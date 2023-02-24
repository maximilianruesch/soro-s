package dbUtils

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	OSMUtil "transform-osm/osm-utils"
)

var XML_TAG_NAME_CONSTR = xml.Name{Space: " ", Local: "tag"}

func MapSignalsWithAnchorSearch(
	dbIss XmlIssDaten,
	osm *OSMUtil.Osm,
	anchors map[string][]*OSMUtil.Node,
	notFoundedSignalsFalling *[]*Signal,
	notFoundedSignalsRising *[]*Signal,
	optionalNewId *int,
) {
	for _, stelle := range dbIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				processHauptsignal(
					*knoten,
					notFoundedSignalsFalling,
					anchors,
					osm,
					true,
					optionalNewId,
				)
				processHauptsignal(
					*knoten,
					notFoundedSignalsRising,
					anchors,
					osm,
					false,
					optionalNewId,
				)
			}
		}
	}
}

func MapSignalsExistingAnchors(
	dbIss XmlIssDaten,
	osm *OSMUtil.Osm,
	anchors map[string][]*OSMUtil.Node,
	notFoundedSignalsFalling *[]*Signal,
	notFoundedSignalsRising *[]*Signal,
	optionalNewId *int,
) int {
	numberFoundSignals := 0
	for _, stelle := range dbIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				searchHauptsignal(
					*knoten,
					&numberFoundSignals,
					notFoundedSignalsFalling,
					anchors,
					osm,
					true,
					optionalNewId,
				)
				searchHauptsignal(
					*knoten,
					&numberFoundSignals,
					notFoundedSignalsRising,
					anchors,
					osm,
					false,
					optionalNewId,
				)
			}
		}
	}

	return numberFoundSignals
}

func processHauptsignal(
	knoten Spurplanknoten,
	notFoundedSignals *[]*Signal,
	anchors map[string][]*OSMUtil.Node,
	osm *OSMUtil.Osm,
	isFalling bool,
	optionalNewId *int,
) {
	for _, signal := range knoten.HauptsigF {
		found := searchAnchorForSignal(signal, isFalling, anchors)

		if found {
			continue
		}

		for _, node := range osm.Node {
			if len(node.Tag) != 0 {
				railwayTag, _ := OSMUtil.FindTagOnNode(node, "railway")
				refTag, _ := OSMUtil.FindTagOnNode(node, "ref")

				if railwayTag == "signal" &&
					strings.ReplaceAll(refTag, " ", "") == signal.Name[0].Value {
					found = insertNewHauptsignal(
						optionalNewId,
						node,
						signal,
						isFalling,
						notFoundedSignals,
						anchors,
						osm,
					)
				}
			}
		}

		if !found {
			*notFoundedSignals = append(*notFoundedSignals, signal)
		}
	}
}

func searchHauptsignal(
	knoten Spurplanknoten,
	numberFoundSignals *int,
	notFoundedSignals *[]*Signal,
	anchors map[string][]*OSMUtil.Node,
	osm *OSMUtil.Osm,
	isFalling bool,
	optionalNewId *int,
) {
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

		maxNode, err := findBestOSMNode(kilometrage, anchors, osm)
		if err != nil {
			*notFoundedSignals = append(*notFoundedSignals, signal)
		} else {
			found := insertNewHauptsignal(
				optionalNewId,
				maxNode,
				signal,
				isFalling,
				notFoundedSignals,
				anchors,
				osm,
			)

			if !found {
				*numberFoundSignals++
			}
		}
	}
}

func searchAnchorForSignal(
	signal *Signal,
	isFalling bool,
	anchors map[string][]*OSMUtil.Node,
) bool {
	kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
	possibleAnchors := anchors[kilometrage]
	if possibleAnchors == nil {
		return false
	}

	found := false
	for _, anchorNode := range possibleAnchors {
		typ, _ := OSMUtil.FindTagOnNode(anchorNode, "type")
		subtyp, _ := OSMUtil.FindTagOnNode(anchorNode, "subtype")
		id, _ := OSMUtil.FindTagOnNode(anchorNode, "id")
		direction, _ := OSMUtil.FindTagOnNode(anchorNode, "direction")

		if typ == "element" &&
			subtyp == "ms" &&
			id == signal.Name[0].Value &&
			direction == "falling" {
			found = true
			break
		}
	}

	return found
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
		for _, possibleAnchor := range possibleAnchors {
			if possibleAnchor == signalNode {

				if anchorKilometrage == signalKilometrage {
					newSignalNode := createNewHauptsignal(
						*newId,
						signalNode,
						signal,
						isFalling,
					)
					OSMUtil.InsertNodeBasedOnExistingNode(osm, &newSignalNode, signalNode.Id)
					*newId++
					anchors[anchorKilometrage] = append(anchors[anchorKilometrage], &newSignalNode)

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

					errorAnchor.Tag = errorAnchor.Tag[:(len(errorAnchor.Tag) - 4)]
				}
				delete(anchors, anchorKilometrage)

				return false
			}
		}
	}

	newSignalNode := createNewHauptsignal(
		*newId,
		signalNode,
		signal,
		isFalling,
	)
	osm.Node = append(osm.Node, &newSignalNode)
	if len(anchors[signalKilometrage]) == 0 {
		anchors[signalKilometrage] = []*OSMUtil.Node{&newSignalNode}
	} else {
		anchors[signalKilometrage] = append(anchors[signalKilometrage], &newSignalNode)
	}
	return true
}

func createNewHauptsignal(
	id int,
	node *OSMUtil.Node,
	signal *Signal,
	isFalling bool,
) OSMUtil.Node {
	directionString := "falling"
	if !isFalling {
		directionString = "rising"
	}

	return OSMUtil.Node{
		Id:  strconv.Itoa(id),
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

func findBestOSMNode(
	kilometrage float64,
	anchors map[string][]*OSMUtil.Node,
	osm *OSMUtil.Osm,
) (*OSMUtil.Node, error) {
	nearestAnchor, secondNearestAnchor := findTwoNearestAnchor(kilometrage, anchors)

	if nearestAnchor == -1.0 || secondNearestAnchor == -1.0 {
		return nil, errors.New("could not find node")
	}

	nearestAnchorString := formatKilometrage(nearestAnchor, anchors)
	secondNearestAnchorString := formatKilometrage(secondNearestAnchor, anchors)

	newNode, err := FindNewNode(
		osm,
		(anchors[nearestAnchorString])[0],
		(anchors[secondNearestAnchorString])[0],
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
	anchors map[string][]*OSMUtil.Node,
) (float64, float64) {
	nearestAnchor := -1.0
	secondNearestAnchor := -1.0

	for anchorKilometrageString := range anchors {
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
	for anchorKilometrageString := range anchors {
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
	anchors map[string][]*OSMUtil.Node,
) string {
	result := strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len(anchors[result]) == 0; result += "0" {
		if !strings.Contains(result, ",") {
			result += ","
		}
	}
	return result
}
