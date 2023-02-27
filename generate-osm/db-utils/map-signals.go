package dbUtils

import (
	"strconv"
	"strings"
	OSMUtil "transform-osm/osm-utils"
)

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
