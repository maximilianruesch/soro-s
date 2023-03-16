package mapper

import (
	"strconv"
	"strings"
	findNodes "transform-osm/db-utils/find-nodes"
	osmUtils "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

// FindAndMapAnchorMainSignals identifies Hauptsignal(S/F)-Node pairs, that match up.
// Matching up in this context means, that the Node has the tags 'railway:signal' and 'name:...' where ... is the same
// (excluding spaces) as the Signal-name.
// In notFoundSignals... all signals that could not be identified will be returned.
func FindAndMapAnchorMainSignals(
	knoten Spurplanknoten,
	osm *osmUtils.Osm,
	anchors map[float64][]*osmUtils.Node,
	notFoundSignalsFalling *[]*NamedSimpleElement,
	notFoundSignalsRising *[]*NamedSimpleElement,
	signalList map[string]osmUtils.Signal,
	foundAnchorCount *int,
	nodeIdCounter *int,
) error {
	conflictingSignalNames := map[string]bool{}
	err := processHauptsignal(
		knoten.HauptsigF,
		notFoundSignalsFalling,
		anchors,
		signalList,
		conflictingSignalNames,
		osm,
		true,
		foundAnchorCount,
		nodeIdCounter,
	)
	if err != nil {
		return errors.Wrap(err, "failed processing falling main signals")
	}

	err = processHauptsignal(
		knoten.HauptsigS,
		notFoundSignalsRising,
		anchors,
		signalList,
		conflictingSignalNames,
		osm,
		false,
		foundAnchorCount,
		nodeIdCounter,
	)
	if err != nil {
		return errors.Wrap(err, "failed processing rising main signals")
	}
	return nil
}

// processHauptsignal iterates over all Hautpsignal[S/F] (depending on 'is Falling')
// and does all the identification.
func processHauptsignal(
	signals []*NamedSimpleElement,
	notFoundSignals *[]*NamedSimpleElement,
	anchors map[float64][]*osmUtils.Node,
	signalList map[string]osmUtils.Signal,
	conflictingSignalNames map[string]bool,
	osm *osmUtils.Osm,
	isFalling bool,
	foundAnchorCount *int,
	nodeIdCounter *int,
) error {
	for _, signal := range signals {
		matchingSignalNodes := []*osmUtils.Node{}

		for _, node := range osm.Node {
			if len(node.Tag) != 0 {
				railwayTag, _ := osmUtils.FindTagOnNode(node, "railway")
				refTag, _ := osmUtils.FindTagOnNode(node, "ref")

				if railwayTag == "signal" &&
					strings.ReplaceAll(refTag, " ", "") == signal.Name.Value {
					matchingSignalNodes = append(matchingSignalNodes, node)
				}
			}
		}

		if len(matchingSignalNodes) == 1 {
			conflictFreeSignal, err := insertNewHauptsignal(
				nodeIdCounter,
				matchingSignalNodes[0],
				signal,
				isFalling,
				notFoundSignals,
				anchors,
				conflictingSignalNames,
				osm,
				foundAnchorCount,
			)
			if err != nil {
				return errors.Wrap(err, "failed to insert signal")
			}
			if conflictFreeSignal {
				*foundAnchorCount++
				signalList[matchingSignalNodes[0].Id] = osmUtils.Signal{
					Name: signal.Name.Value,
					Lat:  matchingSignalNodes[0].Lat,
					Lon:  matchingSignalNodes[0].Lon,
				}
				return nil
			}
		}
		conflictingSignalNames[signal.Name.Value] = true
		*notFoundSignals = append(*notFoundSignals, signal)
	}
	return nil
}

// insertNewHauptsignal tries to insert a new main signal.
// Inserting is prohibited, if the signal name is already conflicting or a conflict could be found.
// A conflict exists, when there are either multiple Signals with the same name but different kilometrages,
// or when there exists more than one node, that could be identified as a certian Signal (i.e. with the same name).
func insertNewHauptsignal(
	nodeIdCounter *int,
	signalNode *osmUtils.Node,
	signal *NamedSimpleElement,
	isFalling bool,
	notFound *[]*NamedSimpleElement,
	anchors map[float64][]*osmUtils.Node,
	conflictingSignalNames map[string]bool,
	osm *osmUtils.Osm,
	foundAnchorCount *int,
) (bool, error) {
	if conflictingSignalNames[signal.Name.Value] {
		return false, nil
	}
	signalKilometrage, err := findNodes.FormatKilometrageStringInFloat(signal.KnotenTyp.Kilometrierung.Value)
	if err != nil {
		return false, errors.Wrap(err, "failed to format kilometrage")
	}
	for anchorKilometrage, currentAnchors := range anchors {
		for _, possibleAnchor := range currentAnchors {
			if possibleAnchor.Lat == signalNode.Lat && possibleAnchor.Lon == signalNode.Lon && anchorKilometrage != signalKilometrage {
				for _, errorAnchor := range currentAnchors {
					errorSignal := NamedSimpleElement{}
					errorSignal.KnotenTyp = KnotenTyp{
						Kilometrierung: Wert{
							Value: strconv.FormatFloat(anchorKilometrage, 'f', 3, 64),
						},
					}
					errorSignal.Name = Wert{
						Value: errorAnchor.Tag[3].V,
					}
					*notFound = append(*notFound, &errorSignal)
					*foundAnchorCount--

					errorAnchor.Tag = errorAnchor.Tag[:(len(errorAnchor.Tag) - 4)]
				}
				delete(anchors, anchorKilometrage)
				return false, nil
			}
		}
	}

	newSignalNode := createNamedDirectionalNode(
		nodeIdCounter,
		signalNode,
		"ms",
		signal.Name.Value,
		isFalling,
	)
	osmUtils.InsertNewNodeWithReferenceNode(
		osm,
		&newSignalNode,
		signalNode,
	)

	if len(anchors[signalKilometrage]) == 0 {
		anchors[signalKilometrage] = []*osmUtils.Node{&newSignalNode}
	} else {
		anchors[signalKilometrage] = append(anchors[signalKilometrage], &newSignalNode)
	}
	return true, nil
}

// mapUnanchoredSignals processes all previously unmapped signals.
// This is main signals, for which no unique Node could be determined ("signalType" = 'ms'),
// or approach and protection signals ("signalType" = 'as'/'ps').
func MapUnanchoredSignals(
	osmData *osmUtils.Osm,
	anchors map[float64]([]*osmUtils.Node),
	signalList map[string]osmUtils.Signal,
	nodeIdCounter *int,
	knoten Spurplanknoten,
	signalType string,
	elementsNotFound map[string]([]string),
) error {
	err := searchUnanchoredSignal(
		osmData,
		anchors,
		signalList,
		nodeIdCounter,
		knoten,
		signalType,
		elementsNotFound,
		true)
	if err != nil {
		return errors.Wrap(err, "failed finding falling main signal")
	}

	err = searchUnanchoredSignal(
		osmData,
		anchors,
		signalList,
		nodeIdCounter,
		knoten,
		signalType,
		elementsNotFound,
		false)
	if err != nil {
		return errors.Wrap(err, "failed finding rising main signal")
	}
	return nil
}

// serachUnanchoredSignal searches for a Node, that best fits the Signal to be mapped.
// This search is based on at least two anchored elements and their respective distance to the signal at hand.
// If no ore only one anchor could be identified, or all anchors are otherwise insufficient, no mapping can be done.
func searchUnanchoredSignal(
	osmData *osmUtils.Osm,
	anchors map[float64]([]*osmUtils.Node),
	signalList map[string]osmUtils.Signal,
	nodeIdCounter *int,
	knoten Spurplanknoten,
	signalType string,
	elementsNotFound map[string]([]string),
	isFalling bool,
) error {
	var signalTypeLong string
	var signals []*NamedSimpleElement

	switch signalType {
	case "ms":
		signalTypeLong = "main signal"
		signals = knoten.HauptsigF
		if !isFalling {
			signals = knoten.HauptsigS
		}
	case "as":
		signalTypeLong = "approach signal"
		signals = knoten.VorsigF
		if !isFalling {
			signals = knoten.VorsigS
		}
	case "ps":
		signalTypeLong = "protection signal"
		signals = knoten.SchutzsigF
		if !isFalling {
			signals = knoten.SchutzsigS
		}
	}

	for _, signal := range signals {
		kilometrage, _ := findNodes.FormatKilometrageStringInFloat(signal.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findNodes.FindBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == findNodes.ErrNoSuitableAnchors {
				elementsNotFound[signalTypeLong+"s"] = append(elementsNotFound[signalTypeLong+"s"], signal.Name.Value)
				continue
			}
			return errors.Wrap(err, "failed to map "+signalTypeLong+" "+signal.Name.Value)

		}

		newSignalNode := createNamedDirectionalNode(
			nodeIdCounter,
			maxNode,
			signalType,
			signal.Name.Value,
			isFalling,
		)
		osmUtils.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
		signalList[newSignalNode.Id] = osmUtils.Signal{
			Name: signal.Name.Value,
			Lat:  newSignalNode.Lat,
			Lon:  newSignalNode.Lon,
		}
	}
	return nil
}
