package dbUtils

import (
	"encoding/xml"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	OSMUtil "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

var XML_TAG_NAME_CONSTR = xml.Name{Space: " ", Local: "tag"}

func findAndMapAnchorMainSignals(
	dbIss XmlIssDaten,
	osm *OSMUtil.Osm,
	anchors map[string][]*OSMUtil.Node,
	notFoundSignalsFalling *[]*Signal,
	notFoundSignalsRising *[]*Signal,
	optionalNewId *int,
) {
	for _, stelle := range dbIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				processHauptsignal(
					*knoten,
					notFoundSignalsFalling,
					anchors,
					osm,
					true,
					optionalNewId,
				)
				processHauptsignal(
					*knoten,
					notFoundSignalsRising,
					anchors,
					osm,
					false,
					optionalNewId,
				)
			}
		}
	}
}

func processHauptsignal(
	knoten Spurplanknoten,
	notFoundSignals *[]*Signal,
	anchors map[string][]*OSMUtil.Node,
	osm *OSMUtil.Osm,
	isFalling bool,
	optionalNewId *int,
) {
	signals := knoten.HauptsigF
	if !isFalling {
		signals = knoten.HauptsigS
	}

	for _, signal := range signals {
		conflictFreeSignal := false
		matchingSignalNodes := []*OSMUtil.Node{}

		for _, node := range osm.Node {
			if len(node.Tag) != 0 {
				railwayTag, _ := OSMUtil.FindTagOnNode(node, "railway")
				refTag, _ := OSMUtil.FindTagOnNode(node, "ref")

				if railwayTag == "signal" &&
					strings.ReplaceAll(refTag, " ", "") == signal.Name[0].Value {
					matchingSignalNodes = append(matchingSignalNodes, node)
				}
			}
		}

		if len(matchingSignalNodes) == 1 {
			conflictFreeSignal = insertNewHauptsignal(
				optionalNewId,
				matchingSignalNodes[0],
				signal,
				isFalling,
				notFoundSignals,
				anchors,
				osm,
			)
			if !conflictFreeSignal {
				*notFoundSignals = append(*notFoundSignals, signal)
			}
		} else {
			*notFoundSignals = append(*notFoundSignals, signal)
		}
	}
}

func insertNewHauptsignal(
	newId *int,
	signalNode *OSMUtil.Node,
	signal *Signal,
	isFalling bool,
	notFound *[]*Signal,
	anchors map[string][]*OSMUtil.Node,
	osm *OSMUtil.Osm,
) bool {
	signalKilometrage := signal.KnotenTyp.Kilometrierung[0].Value
	for anchorKilometrage, possibleAnchors := range anchors {
		for _, possibleAnchorPair := range possibleAnchors {
			if possibleAnchorPair.Id == signalNode.Id && anchorKilometrage != signalKilometrage {
				for _, errorAnchor := range possibleAnchors {
					errorSignal := Signal{}
					errorSignal.KnotenTyp = KnotenTyp{
						Kilometrierung: []*Wert{{
							Value: anchorKilometrage,
						}},
					}
					errorSignal.Name = []*Wert{{
						Value: signal.Name[0].Value,
					}}
					*notFound = append(*notFound, &errorSignal)

					errorAnchor.Tag = errorAnchor.Tag[:(len(errorAnchor.Tag) - 4)]
				}
				delete(anchors, anchorKilometrage)
				print("Been here \n")
				return false
			}
		}
	}
	newSignalNode := createNewHauptsignal(
		newId,
		signalNode,
		signal,
		isFalling,
	)
	OSMUtil.InsertNewNodeWithReferenceNode(osm, &newSignalNode, signalNode)
	if len(anchors[signalKilometrage]) == 0 {
		anchors[signalKilometrage] = []*OSMUtil.Node{&newSignalNode}
	} else {
		anchors[signalKilometrage] = append(anchors[signalKilometrage], &newSignalNode)
	}
	return true
}

func createNewHauptsignal(
	id *int,
	node *OSMUtil.Node,
	signal *Signal,
	isFalling bool,
) OSMUtil.Node {
	directionString := "falling"
	if !isFalling {
		directionString = "rising"
	}
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONSTR, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONSTR, K: "subtype", V: "ms"},
			{XMLName: XML_TAG_NAME_CONSTR, K: "id", V: signal.Name[0].Value},
			{XMLName: XML_TAG_NAME_CONSTR, K: "direction", V: directionString},
		},
	}
}

func mapUnanchoredMainSignals(
	osmData *OSMUtil.Osm,
	anchors *map[string]([]*OSMUtil.Node),
	nodeIdCounter *int,
	dbData XmlIssDaten,
) {
	for _, stelle := range dbData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				searchUnanchoredMainSignal(
					osmData,
					anchors,
					nodeIdCounter,
					*knoten,
					true)
				searchUnanchoredMainSignal(
					osmData,
					anchors,
					nodeIdCounter,
					*knoten,
					false)
			}
		}
	}
}

func searchUnanchoredMainSignal(
	osmData *OSMUtil.Osm,
	anchors *map[string]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	isFalling bool,
) {
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
		kilometrage, _ := strconv.ParseFloat(
			strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."),
			64)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			fmt.Printf("Error: %s \n", err.Error())
			continue
		}

		*nodeIdCounter++
		newSignalNode := OSMUtil.Node{
			Id:  strconv.Itoa(*nodeIdCounter),
			Lat: maxNode.Lat,
			Lon: maxNode.Lon,
			Tag: []*OSMUtil.Tag{
				{XMLName: XML_TAG_NAME_CONSTR, K: "type", V: "element"},
				{XMLName: XML_TAG_NAME_CONSTR, K: "subtype", V: "ms"},
				{XMLName: XML_TAG_NAME_CONSTR, K: "id", V: signal.Name[0].Value},
				{XMLName: XML_TAG_NAME_CONSTR, K: "direction", V: directionString},
			},
		}
		OSMUtil.InsertNewNodeWithReferenceNode(osmData, &newSignalNode, maxNode)
	}
}

func findBestOSMNode(
	osmData *OSMUtil.Osm,
	anchors *map[string]([]*OSMUtil.Node),
	kilometrage float64,
) (*OSMUtil.Node, error) {
	sortedAnchors := sortAnchors(anchors, kilometrage)

	nearest, _ := strconv.ParseFloat(sortedAnchors[0], 64)
	secondNearest, _ := strconv.ParseFloat(sortedAnchors[0], 64)

	anchor1 := ((*anchors)[sortedAnchors[0]])[0]
	anchor2 := ((*anchors)[sortedAnchors[1]])[0]
	distance1 := math.Abs(nearest - kilometrage)
	distance2 := math.Abs(secondNearest - kilometrage)

	newNode, err := findNewNode(
		osmData,
		anchor1,
		anchor2,
		distance1,
		distance2,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not find OSM-node")
	}

	return newNode, nil
}

func sortAnchors(
	anchors *map[string]([]*OSMUtil.Node),
	kilometrage float64,
) []string {
	anchorKeys := []string{}
	for anchorKey := range *anchors {
		anchorKeys = append(anchorKeys, anchorKey)
	}

	sort.SliceStable(anchorKeys, func(i, j int) bool {
		floatKilometrage1, _ := strconv.ParseFloat(anchorKeys[i], 64)
		floatKilometrage2, _ := strconv.ParseFloat(anchorKeys[j], 64)
		return math.Abs(kilometrage-floatKilometrage1) < math.Abs(kilometrage-floatKilometrage2)
	})

	return anchorKeys
}

func formatKilometrage(
	anchors *map[string]([]*OSMUtil.Node),
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
