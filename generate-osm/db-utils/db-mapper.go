package dbUtils

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	OSMUtil "transform-osm/osm-utils"
)

var XML_TAG_NAME_CONST = xml.Name{Space: " ", Local: "tag"}
var numItemsNotFound int
var numItemsFound int

func MapDB(
	nodeIdCounter *int,
	refs []string, osmDir string, DBDir string,
) {
	for _, line := range refs {
		var anchors = make(map[string]([]*OSMUtil.Node))

		numItemsFound = 0
		numItemsNotFound = 0

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

		fmt.Printf("Processing line %s \n", line)

		mainF, mainS := findAndMapAnchorMainSignals(&osmData, &anchors,
			nodeIdCounter, dbData)

		// anchorPoints(&osmData, dbData)
		fmt.Printf("Found %d anchors and could not find %d \n", numItemsFound, numItemsNotFound)

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
		_ = restData

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml",
				[]byte(xml.Header+string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
}

func findAndMapAnchorMainSignals(
	osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int,
	dbData XmlIssDaten,
) ([]*Signal, []*Signal) {
	var main_sigF []*Signal = []*Signal{}
	var main_sigS []*Signal = []*Signal{}
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				main_sigF = append(main_sigF,
					anchorMainSignal(osmData, anchors, nodeIdCounter,
						*knoten, true)...)
				main_sigS = append(main_sigS,
					anchorMainSignal(osmData, anchors, nodeIdCounter,
						*knoten, false)...)
			}
		}
	}
	return main_sigF, main_sigS
}

func anchorMainSignal(
	osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int,
	knoten Spurplanknoten, isFalling bool,
) []*Signal {
	var notFoundSignals = []*Signal{}

	directionString := "falling"
	signals := knoten.HauptsigF
	if !isFalling {
		directionString = "rising"
		signals = knoten.HauptsigS
	}

	for _, signal := range signals {
		conflictFreeSignal := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		nodesFound := []*OSMUtil.Node{}

		for _, node := range osmData.Node {
			if len(node.Tag) != 0 {
				is_signal := false
				has_correct_id := false

				railwayTag, _ := OSMUtil.FindTagOnNode(*node, "railway")
				is_signal = railwayTag == "signal"

				idTag, _ := OSMUtil.FindTagOnNode(*node, "ref")
				has_correct_id = strings.ReplaceAll(idTag, " ", "") == signal.Name[0].Value

				if is_signal && has_correct_id {
					nodesFound = append(nodesFound, node)

				}
			}
		}

		if len(nodesFound) == 1 {
			conflictFreeSignal = insertNewHauptsig(osmData, anchors, nodeIdCounter,
				nodesFound[0], kilometrage, *signal, directionString, conflictFreeSignal, &notFoundSignals)
			if !conflictFreeSignal {
				notFoundSignals = append(notFoundSignals, signal)
				numItemsNotFound++
				break
			}
		} else {
			notFoundSignals = append(notFoundSignals, signal)
			numItemsNotFound++
		}

		if conflictFreeSignal {
			numItemsFound++
		}
	}

	return notFoundSignals
}

func formatKilometrage(anchors *map[string]([]*OSMUtil.Node),
	in float64,
) (out string) {
	out = strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len((*anchors)[out]) == 0; out += "0" {
		if !strings.Contains(out, ",") {
			out += ","
		}
	}
	return
}

func insertNewHauptsig(
	osmData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node), nodeIdCounter *int,
	node *OSMUtil.Node, kilometrage string, signal Signal, direction string, alreadyFound bool, notFound *[]*Signal,
) bool {
	for anchorKilometrage, anchorList := range *anchors {
		for _, anchor := range anchorList {
			if anchor == node {
				if anchorKilometrage == kilometrage {
					newNode := OSMUtil.Node{
						Id:  strconv.Itoa(*nodeIdCounter),
						Lat: node.Lat,
						Lon: node.Lon,
						Tag: []*OSMUtil.Tag{
							{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
							{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
							{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name[0].Value},
							{XMLName: XML_TAG_NAME_CONST, K: "direction", V: direction}}}
					OSMUtil.InsertNode(&newNode, node.Id, osmData)
					*nodeIdCounter++
					(*anchors)[kilometrage] = append((*anchors)[kilometrage], &newNode)
					return true
				}

				for _, errorSignal := range anchorList {
					errorSignalName, _ := OSMUtil.FindTagOnNode(*errorSignal, "id")
					*notFound = append(*notFound, &Signal{
						KnotenTyp{Kilometrierung: []*Wert{{Value: anchorKilometrage}}},
						[]*Wert{{Value: errorSignalName}}})
					errorSignal.Tag = errorSignal.Tag[:(len(errorSignal.Tag) - 4)]
					numItemsNotFound++
					numItemsFound--
				}
				delete(*anchors, anchorKilometrage)
				return false
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
