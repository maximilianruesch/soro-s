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

var TagName = xml.Name{" ", "tag"}
var id_counter = 1
var anchors map[string]([]*OSMUtil.Node)
var num_found = 0

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {
		anchors = make(map[string]([]*OSMUtil.Node))

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

		mainF, mainS := findAndMapAnchorMainSignals(&osmData, dbData)
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

		mapUnanchoredMainSignals(&osmData, restData)
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

func findAndMapAnchorMainSignals(osmData *OSMUtil.Osm, dbData XmlIssDaten) ([]*Signal, []*Signal) {
	var main_sigF []*Signal = []*Signal{}
	var main_sigS []*Signal = []*Signal{}
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				main_sigF = append(main_sigF, anchorMainSignal(osmData, *knoten, true)...)
				main_sigS = append(main_sigS, anchorMainSignal(osmData, *knoten, false)...)
			}
		}
	}
	return main_sigF, main_sigS
}

func mapUnanchoredMainSignals(osmData *OSMUtil.Osm, dbData XmlIssDaten) {
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				searchHauptsigF(osmData, *knoten)
				searchHauptsigS(osmData, *knoten)
			}
		}
	}
}

func anchorMainSignal(osmData *OSMUtil.Osm, knoten Spurplanknoten, isFalling bool) []*Signal {
	var notFoundSignals = []*Signal{}

	signals := knoten.HauptsigF
	if !isFalling {
		signals = knoten.HauptsigS
	}

	for _, signal := range signals {
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value

		for _, node := range osmData.Node {
			if len(node.Tag) != 0 {
				is_signal := false
				has_correct_id := false
				if railwayTag, err := OSMUtil.FindTagOnNode(*node, "railway"); err == nil && railwayTag == "signal" {
					is_signal = true
				}
				if idTag, err := OSMUtil.FindTagOnNode(*node, "ref"); err == nil && strings.ReplaceAll(idTag, " ", "") == signal.Name[0].Value {
					has_correct_id = true
				}

				if is_signal && has_correct_id {
					found = insertNewHauptsig(osmData, node, kilometrage, *signal, "falling", &notFoundSignals)
				}
			}
		}

		if !found {
			notFoundSignals = append(notFoundSignals, signal)
		}
	}

	return notFoundSignals
}

func searchHauptsigF(osmData *OSMUtil.Osm, knoten Spurplanknoten) {
	var not_found = []*Signal{}

	if len(anchors) == 0 {
		fmt.Print("Could not find anchors! \n")
		return
	}
	if len(anchors) == 1 {
		fmt.Print("Could not find enough anchors! \n")
		// TODO: Node not found, find closest mapped Node and work from there
		return
	}

	for _, signal := range knoten.HauptsigF {
		kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

		maxNode, err := findBestOSMNode(osmData, kilometrage)
		if err != nil {
			not_found = append(not_found, signal)
		} else {
			found := insertNewHauptsig(osmData, maxNode, signal.KnotenTyp.Kilometrierung[0].Value, *signal, "falling", &not_found)
			if !found {
				num_found++
			}
		}
	}
}

func searchHauptsigS(osmData *OSMUtil.Osm, knoten Spurplanknoten) {
	var not_found = []*Signal{}

	if len(anchors) == 0 {
		fmt.Print("Could not find anchors! \n")
		return
	}
	if len(anchors) == 1 {
		fmt.Print("Could not find enough anchors! \n")
		// TODO: Node not found, find closest mapped Node and work from there
		return
	}

	for _, signal := range knoten.HauptsigS {
		kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

		maxNode, err := findBestOSMNode(osmData, kilometrage)
		if err != nil {
			not_found = append(not_found, signal)
		} else {
			found := insertNewHauptsig(osmData, maxNode, signal.KnotenTyp.Kilometrierung[0].Value, *signal, "rising", &not_found)
			if !found {
				num_found++
			}
		}
	}
}

func findTwoNearest(kilometrage float64) (nearest float64, second_nearest float64) {
	nearest = -1.0
	second_nearest = -1.0

	for key := range anchors {
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
	for key := range anchors {
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

func formatKilometrage(in float64) (out string) {
	out = strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len(anchors[out]) == 0; out += "0" {
		if !strings.Contains(out, ",") {
			out += ","
		}
	}
	return
}

func findBestOSMNode(osmData *OSMUtil.Osm, kilometrage float64) (*OSMUtil.Node, error) {
	nearest, second_nearest := findTwoNearest(kilometrage)

	if nearest == -1.0 || second_nearest == -1.0 {
		return nil, errors.New("Could not find node.")
	}

	nearest_string := formatKilometrage(nearest)
	second_nearest_string := formatKilometrage(second_nearest)

	newNode, err := FindNewNode(osmData, (anchors[nearest_string])[0], (anchors[second_nearest_string])[0], math.Abs(nearest-kilometrage), math.Abs(second_nearest-kilometrage))
	if err != nil {
		return nil, errors.New("Could not find node.")
	}

	return newNode, nil
}

func insertNewHauptsig(osmData *OSMUtil.Osm, node *OSMUtil.Node, kilometrage string, signal Signal, direction string, notFound *[]*Signal) bool {
	for key, value_list := range anchors {
		for _, value := range value_list {
			if value == node {
				if key == kilometrage {
					newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag: []*OSMUtil.Tag{
						{XMLName: TagName, K: "type", V: "element"},
						{XMLName: TagName, K: "subtype", V: "ms"},
						{XMLName: TagName, K: "id", V: signal.Name[0].Value},
						{XMLName: TagName, K: "direction", V: direction}}}
					OSMUtil.InsertNode(&newNode, node.Id, osmData)
					id_counter++
					anchors[key] = append(anchors[key], &newNode)
					return true
				} else {
					for _, error_val := range value_list {
						*notFound = append(*notFound, &Signal{
							KnotenTyp{Kilometrierung: []*Wert{{Value: key}}},
							[]*Wert{{Value: signal.Name[0].Value}}})
						error_val.Tag = error_val.Tag[:(len(error_val.Tag) - 4)]
					}
					delete(anchors, key)
					return false
				}
			}
		}
	}

	node.Tag = append(node.Tag, []*OSMUtil.Tag{
		{XMLName: TagName, K: "type", V: "element"},
		{XMLName: TagName, K: "subtype", V: "ms"},
		{XMLName: TagName, K: "id", V: signal.Name[0].Value},
		{XMLName: TagName, K: "direction", V: direction}}...)
	if len(anchors[kilometrage]) == 0 {
		anchors[kilometrage] = []*OSMUtil.Node{node}
	} else {
		anchors[kilometrage] = append(anchors[kilometrage], node)
	}
	return true
}
