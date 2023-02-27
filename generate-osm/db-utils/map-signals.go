package dbUtils

import (
	"encoding/xml"
	"strconv"
	"strings"
	OSMUtil "transform-osm/osm-utils"
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
			*optionalNewId++
			if !conflictFreeSignal {
				*notFoundSignals = append(*notFoundSignals, signal)
			}
		} else {
			*notFoundSignals = append(*notFoundSignals, signal)
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
			if possibleAnchor.Id == signalNode.Id {
				if anchorKilometrage == signalKilometrage {
					newSignalNode := createNewHauptsignal(
						*newId,
						signalNode,
						signal,
						isFalling,
					)
					OSMUtil.InsertSignalWithWayRef(osm, &newSignalNode, signalNode.Id)
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
