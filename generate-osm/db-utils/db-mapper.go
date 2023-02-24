package dbUtils

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	OSMUtil "transform-osm/osm-utils"
)

var XML_TAG_NAME_CONST = xml.Name{Space: " ", Local: "tag"}
var num_found = 0

func MapDB(nodeIdCounter *int, refs []string, osmDir string, DBDir string) {
	for _, line := range refs {
		var anchors = make(map[string]([]*OSMUtil.Node))

		var dbData XmlIssDaten
		var osmData OSMUtil.Osm
		osmData = OSMUtil.Osm{}

		osm_file, err := os.ReadFile(osmDir + "/" + line + ".xml")
		if err != nil {
			log.Fatal(err)
		}
		db_file, err := os.ReadFile(DBDir + "/" + line + "_DB.xml")
		if err != nil {
			log.Fatal(err)
		}

		if err := xml.Unmarshal([]byte(osm_file), &osmData); err != nil {
			panic(err)
		}
		if err := xml.Unmarshal([]byte(db_file), &dbData); err != nil {
			panic(err)
		}

		print("Processing line ")
		print(line)
		print("\n")

		mainF, mainS := findAndMapAnchorMainSignals(&osmData, &anchors, nodeIdCounter, dbData)
		// mapPoints(&osmData, dbData)

		var restData = XmlIssDaten{
			Betriebsstellen: []*Spurplanbetriebsstelle{{
				Abschnitte: []*Spurplanabschnitt{{
					Knoten: []*Spurplanknoten{{
						HauptsigF: mainF,
						HauptsigS: mainS,
					}},
				}},
			}},
		}

		mapUnanchoredMainSignals(&osmData, &anchors, nodeIdCounter, restData)
		// mapPoints(restData)
		// mapRest(dbData)

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header+string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
	fmt.Printf("Could not find: %d \n", num_found)
}

func findAndMapAnchorMainSignals(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int, dbData XmlIssDaten) ([]*Signal, []*Signal) {
	var main_sigF []*Signal = []*Signal{}
	var main_sigS []*Signal = []*Signal{}
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				main_sigF = append(main_sigF, anchorMainSignal(osmData, anchors, nodeIdCounter, *knoten, true)...)
				main_sigS = append(main_sigS, anchorMainSignal(osmData, anchors, nodeIdCounter, *knoten, false)...)
			}
		}
	}
	return main_sigF, main_sigS
}

func mapUnanchoredMainSignals(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int, dbData XmlIssDaten) {
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				searchUnanchoredMainSignal(osmData, anchors, nodeIdCounter, *knoten, true)
				searchUnanchoredMainSignal(osmData, anchors, nodeIdCounter, *knoten, false)
			}
		}
	}
}

func anchorMainSignal(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int, knoten Spurplanknoten, isFalling bool) []*Signal {
	var notFoundSignals = []*Signal{}

	directionString := "falling"
	signals := knoten.HauptsigF
	if !isFalling {
		directionString = "rising"
		signals = knoten.HauptsigS
	}

	for _, signal := range signals {
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value

		for _, node := range osmData.Node {
			if len(node.Tag) != 0 {
				is_signal := false
				has_correct_id := false

				railwayTag, _ := OSMUtil.FindTagOnNode(*node, "railway")
				is_signal = railwayTag == "signal"

				idTag, _ := OSMUtil.FindTagOnNode(*node, "ref")
				has_correct_id = strings.ReplaceAll(idTag, " ", "") == signal.Name[0].Value

				if is_signal && has_correct_id {
					found = insertNewHauptsig(osmData, anchors, nodeIdCounter, node, kilometrage, *signal, directionString, &notFoundSignals)
				}
			}
		}

		if !found {
			notFoundSignals = append(notFoundSignals, signal)
		}
	}

	return notFoundSignals
}

func searchUnanchoredMainSignal(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int, knoten Spurplanknoten, isFalling bool) {
	if len(*anchors) == 0 {
		fmt.Print("Could not find anchors! \n")
		return
	}
	if len(*anchors) == 1 {
		fmt.Print("Could not find enough anchors! \n")
		// TODO: Node not found, find closest mapped Node and work from there
		return
	}

	directionString := "falling"
	signals := knoten.HauptsigF
	if !isFalling {
		directionString = "rising"
		signals = knoten.HauptsigS
	}

	for _, signal := range signals {
		kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			num_found++
			return
		}

		found := insertNewHauptsig(osmData, anchors, nodeIdCounter, maxNode, signal.KnotenTyp.Kilometrierung[0].Value, *signal, directionString, &[]*Signal{})
		if !found {
			num_found++
		}
	}
}

func findBestOSMNode(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), kilometrage float64) (*OSMUtil.Node, error) {
	nearest, second_nearest := findTwoNearest(anchors, kilometrage)

	if nearest == -1.0 || second_nearest == -1.0 {
		return nil, errors.New("Could not find anchors.")
	}

	nearest_string := formatKilometrage(anchors, nearest)
	second_nearest_string := formatKilometrage(anchors, second_nearest)

	newNode, err := FindNewNode(osmData, ((*anchors)[nearest_string])[0], ((*anchors)[second_nearest_string])[0], math.Abs(nearest-kilometrage), math.Abs(second_nearest-kilometrage))
	if err != nil {
		return nil, errors.New("Could not find node.")
	}

	return newNode, nil
}

func findTwoNearest(anchors *map[string]([]*OSMUtil.Node), kilometrage float64) (nearest float64, second_nearest float64) {
	nearest = -1.0
	second_nearest = -1.0

	for key := range *anchors {
		if !strings.Contains(key, "+") {
			float_key, _ := strconv.ParseFloat(strings.ReplaceAll(key, ",", "."), 64)
			if nearest == -1.0 {
				nearest = float_key
			}
			if math.Abs(float_key-kilometrage) < math.Abs(nearest-kilometrage) {
				second_nearest = nearest
				nearest = float_key
			}
		}
	}

	if second_nearest != -1.0 {
		return
	}
	for key := range *anchors {
		if !strings.Contains(key, "+") {
			float_key, _ := strconv.ParseFloat(strings.ReplaceAll(key, ",", "."), 64)
			if float_key != nearest {
				if second_nearest == -1.0 {
					second_nearest = float_key
				}
				if math.Abs(float_key-kilometrage) < math.Abs(second_nearest-kilometrage) {
					second_nearest = float_key
				}
			}
		}
	}
	return
}

func formatKilometrage(anchors *map[string]([]*OSMUtil.Node), in float64) (out string) {
	out = strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len((*anchors)[out]) == 0; out += "0" {
		if !strings.Contains(out, ",") {
			out += ","
		}
	}
	return
}

func insertNewHauptsig(osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int, node *OSMUtil.Node, kilometrage string, signal Signal, direction string, notFound *[]*Signal) bool {
	for key, value_list := range *anchors {
		for _, value := range value_list {
			if value == node {
				if key == kilometrage {
					newNode := OSMUtil.Node{Id: strconv.Itoa(*nodeIdCounter), Lat: node.Lat, Lon: node.Lon, Tag: []*OSMUtil.Tag{
						{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
						{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
						{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name[0].Value},
						{XMLName: XML_TAG_NAME_CONST, K: "direction", V: direction}}}
					OSMUtil.InsertNode(&newNode, node.Id, osmData)
					*nodeIdCounter++
					(*anchors)[key] = append((*anchors)[key], &newNode)
					return true
				} else {
					for _, error_val := range value_list {
						*notFound = append(*notFound, &Signal{
							KnotenTyp{Kilometrierung: []*Wert{{Value: key}}},
							[]*Wert{{Value: signal.Name[0].Value}}})
						error_val.Tag = error_val.Tag[:(len(error_val.Tag) - 4)]
					}
					delete(*anchors, key)
					return false
				}
			}
		}
	}

	node.Tag = append(node.Tag, []*OSMUtil.Tag{
		{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
		{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
		{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name[0].Value},
		{XMLName: XML_TAG_NAME_CONST, K: "direction", V: direction}}...)
	if len((*anchors)[kilometrage]) == 0 {
		(*anchors)[kilometrage] = []*OSMUtil.Node{node}
	} else {
		(*anchors)[kilometrage] = append((*anchors)[kilometrage], node)
	}
	return true
}