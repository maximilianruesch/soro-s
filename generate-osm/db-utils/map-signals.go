package dbUtils

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	osmUtils "transform-osm/osm-utils"
)

var XML_TAG_NAME_CONST = xml.Name{Space: " ", Local: "tag"}

func MapMainSignalsWithAnchorSearch(
	osm *osmUtils.Osm,
	dbIss XmlIssDaten,
	anchors *map[string][]*osmUtils.Node,
	notFoundSignalsFalling *[]*Signal,
	notFoundSignalsRising *[]*Signal,
	optionalNewId *int,
	numberNotFoundSignals *int,
) {
	for _, stelle := range dbIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				processHauptsignal(
					osm,
					*knoten,
					notFoundSignalsFalling,
					anchors,
					true,
					optionalNewId,
					numberNotFoundSignals,
				)
				processHauptsignal(
					osm,
					*knoten,
					notFoundSignalsRising,
					anchors,
					false,
					optionalNewId,
					numberNotFoundSignals,
				)
			}
		}
	}
}

func MapMainSignalsExistingAnchors(
	osm *osmUtils.Osm,
	dbIss XmlIssDaten,
	anchors *map[string][]*osmUtils.Node,
	notFoundSignalsFalling *[]*Signal,
	notFoundSignalsRising *[]*Signal,
	optionalNewId *int,
	numberNotFoundSignals *int,
) {
	for _, stelle := range dbIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				searchHauptsignal(
					osm,
					*knoten,
					numberNotFoundSignals,
					notFoundSignalsFalling,
					anchors,
					true,
					optionalNewId,
				)
				searchHauptsignal(
					osm,
					*knoten,
					numberNotFoundSignals,
					notFoundSignalsRising,
					anchors,
					false,
					optionalNewId,
				)
			}
		}
	}
}

func processHauptsignal(
	osm *osmUtils.Osm,
	knoten Spurplanknoten,
	notFoundSignals *[]*Signal,
	anchors *map[string][]*osmUtils.Node,
	isFalling bool,
	optionalNewId *int,
	numberNotFoundSignals *int,
) {
	signals := knoten.HauptsigF
	if !isFalling {
		signals = knoten.HauptsigS
	}
	for _, signal := range signals {
		/*
			found := searchAnchorForSignal(signal, isFalling, anchors)

			if found {
				print("Continued")
				continue
			}
		*/
		var found bool
		for _, node := range osm.Node {
			if len(node.Tag) != 0 {
				railwayTag, _ := osmUtils.FindTagOnNode(node, "railway")
				refTag, _ := osmUtils.FindTagOnNode(node, "ref")

				if railwayTag == "signal" &&
					strings.ReplaceAll(refTag, " ", "") == signal.Name[0].Value {
					found = insertNewHauptsignal(
						osm,
						optionalNewId,
						node,
						signal,
						isFalling,
						notFoundSignals,
						anchors,
						numberNotFoundSignals,
					)
				}
			}
		}

		if !found {
			*notFoundSignals = append(*notFoundSignals, signal)
		}
	}
}

func searchHauptsignal(
	osm *osmUtils.Osm,
	knoten Spurplanknoten,
	numberNotFoundSignals *int,
	notFoundSignals *[]*Signal,
	anchors *map[string][]*osmUtils.Node,
	isFalling bool,
	optionalNewId *int,
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

	signals := knoten.HauptsigF
	if !isFalling {
		signals = knoten.HauptsigS
	}
	for _, signal := range signals {
		kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

		bestNode, err := findBestOSMNode(osm, kilometrage, anchors)
		if err != nil {
			*notFoundSignals = append(*notFoundSignals, signal)
		} else {
			found := insertNewHauptsignal(
				osm,
				optionalNewId,
				bestNode,
				signal,
				isFalling,
				notFoundSignals,
				anchors,
				numberNotFoundSignals,
			)

			if !found {
				print("Could not find signal\n")
				*numberNotFoundSignals += 1
			}
		}
	}
}

func searchAnchorForSignal(
	signal *Signal,
	isFalling bool,
	anchors *map[string][]*osmUtils.Node,
) bool {
	kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
	possibleAnchors := (*anchors)[kilometrage]
	if possibleAnchors == nil {
		return false
	}

	directionString := "falling"
	if !isFalling {
		directionString = "rising"
	}
	for _, anchorNode := range possibleAnchors {
		typ, _ := osmUtils.FindTagOnNode(anchorNode, "type")
		subtyp, _ := osmUtils.FindTagOnNode(anchorNode, "subtype")
		id, _ := osmUtils.FindTagOnNode(anchorNode, "id")
		direction, _ := osmUtils.FindTagOnNode(anchorNode, "direction")

		if typ == "element" &&
			subtyp == "ms" &&
			id == signal.Name[0].Value &&
			direction == directionString {
			return true
		}
	}

	return false
}

func insertNewHauptsignal(
	osm *osmUtils.Osm,
	newId *int,
	signalNode *osmUtils.Node,
	signal *Signal,
	isFalling bool,
	notFound *[]*Signal,
	anchors *map[string][]*osmUtils.Node,
	numberNotFoundSignals *int,
) bool {
	signalKilometrage := signal.KnotenTyp.Kilometrierung[0].Value
	for anchorKilometrage, possibleAnchors := range *anchors {
		for _, possibleAnchor := range possibleAnchors {
			if possibleAnchor == signalNode {

				if anchorKilometrage == signalKilometrage {
					newSignalNode := createNewHauptsignal(
						*newId,
						signalNode,
						signal,
						isFalling,
					)
					osmUtils.InsertNodeBasedOnExistingNode(osm, &newSignalNode, signalNode.Id)
					*newId += 1
					(*anchors)[anchorKilometrage] = append((*anchors)[anchorKilometrage], &newSignalNode)

					return true
				}

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
					*numberNotFoundSignals += 1
					errorAnchor.Tag = errorAnchor.Tag[:(len(errorAnchor.Tag) - 4)]
				}
				print("Had a conflict!\n")
				delete(*anchors, anchorKilometrage)

				return false
			}
		}
	}

	directionString := "falling"
	if !isFalling {
		directionString = "rising"
	}

	signalNode.Tag = append(signalNode.Tag, []*osmUtils.Tag{
		{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
		{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
		{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name[0].Value},
		{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
	}...)

	if len((*anchors)[signalKilometrage]) == 0 {
		(*anchors)[signalKilometrage] = []*osmUtils.Node{signalNode}
	} else {
		(*anchors)[signalKilometrage] = append((*anchors)[signalKilometrage], signalNode)
	}

	/*
		newSignalNode := createNewHauptsignal(
			*newId,
			signalNode,
			signal,
			isFalling,
		)
		osm.Node = append(osm.Node, &newSignalNode)
		*newId += 1
		if len((*anchors)[signalKilometrage]) == 0 {
			(*anchors)[signalKilometrage] = []*osmUtils.Node{&newSignalNode}
		} else {
			(*anchors)[signalKilometrage] = append((*anchors)[signalKilometrage], &newSignalNode)
		}
	*/
	return true
}

func createNewHauptsignal(
	id int,
	node *osmUtils.Node,
	signal *Signal,
	isFalling bool,
) osmUtils.Node {
	directionString := "falling"
	if !isFalling {
		directionString = "rising"
	}

	return osmUtils.Node{
		Id:  strconv.Itoa(id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*osmUtils.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name[0].Value},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

func findBestOSMNode(
	osm *osmUtils.Osm,
	kilometrage float64,
	anchors *map[string][]*osmUtils.Node,
) (*osmUtils.Node, error) {
	nearestAnchor, secondNearestAnchor := findTwoNearestAnchor(kilometrage, anchors)

	if nearestAnchor == -1.0 || secondNearestAnchor == -1.0 {
		return nil, errors.New("could not find anchors")
	}

	nearestAnchorString := formatKilometrage(nearestAnchor, anchors)
	secondNearestAnchorString := formatKilometrage(secondNearestAnchor, anchors)

	newNode, err := FindNewNode(
		osm,
		((*anchors)[nearestAnchorString])[0],
		((*anchors)[secondNearestAnchorString])[0],
		math.Abs(nearestAnchor-kilometrage),
		math.Abs(secondNearestAnchor-kilometrage),
	)
	if err != nil {
		return nil, errors.New("could not find node")
	}

	return newNode, nil
}

func findTwoNearestAnchor(
	kilometrage float64,
	anchors *map[string][]*osmUtils.Node,
) (float64, float64) {
	nearestAnchor := -1.0
	secondNearestAnchor := -1.0

	for anchorKilometrageString := range *anchors {
		if strings.Contains(anchorKilometrageString, "+") {
			continue
		}

		anchorKilometrage, _ := strconv.ParseFloat(
			strings.ReplaceAll(anchorKilometrageString, ",", "."),
			64,
		)
		if nearestAnchor == -1.0 {
			nearestAnchor = anchorKilometrage
		}
		if math.Abs(anchorKilometrage-kilometrage) < math.Abs(nearestAnchor-kilometrage) {
			secondNearestAnchor = nearestAnchor
			nearestAnchor = anchorKilometrage
		}

	}

	if secondNearestAnchor != -1.0 {
		return nearestAnchor, secondNearestAnchor
	}
	for anchorKilometrageString := range *anchors {
		if strings.Contains(anchorKilometrageString, "+") {
			continue
		}

		anchorKilometrage, _ := strconv.ParseFloat(
			strings.ReplaceAll(anchorKilometrageString, ",", "."),
			64,
		)

		if anchorKilometrage != nearestAnchor {
			if secondNearestAnchor == -1.0 {
				secondNearestAnchor = anchorKilometrage
			}
			if math.Abs(anchorKilometrage-kilometrage) < math.Abs(secondNearestAnchor-kilometrage) {
				secondNearestAnchor = anchorKilometrage
			}
		}
	}
	return nearestAnchor, secondNearestAnchor
}

func formatKilometrage(
	in float64,
	anchors *map[string][]*osmUtils.Node,
) string {
	result := strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len((*anchors)[result]) == 0; result += "0" {
		if !strings.Contains(result, ",") {
			result += ","
		}
	}
	return result
}
