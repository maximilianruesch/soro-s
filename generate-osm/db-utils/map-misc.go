package dbUtils

import (
	OSMUtil "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

// mapBumper processes all bumpers.
func mapBumper(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	abschnitt Spurplanabschnitt,
	elementsNotFound map[string]([]string),
) error {
	for _, knoten := range abschnitt.Knoten {
		for _, bumper := range knoten.Prellbock {
			kilometrage, _ := formatKilometrageStringInFloat(bumper.KnotenTyp.Kilometrierung.Value)

			maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
			if err != nil {
				if errors.Cause(err) == errNoSuitableAnchors {
					elementsNotFound["bumpers"] = append(elementsNotFound["bumpers"], bumper.Kilometrierung.Value)
					continue
				}
				return errors.Wrap(err, "failed to map bumper "+bumper.Kilometrierung.Value)
			}

			newSignalNode := createNewBumper(
				nodeIdCounter,
				maxNode,
			)
			OSMUtil.InsertNewNodeWithReferenceNode(
				osmData,
				&newSignalNode,
				maxNode,
			)
		}
	}
	return nil
}

// mapBumper processes all bumpers.
func mapBorder(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	abschnitt Spurplanabschnitt,
	elementsNotFound map[string]([]string),
) error {
	for _, knoten := range abschnitt.Knoten {
		for _, border := range knoten.BetriebsStGr {
			kilometrage, _ := formatKilometrageStringInFloat(border.KnotenTyp.Kilometrierung.Value)

			maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
			if err != nil {
				if errors.Cause(err) == errNoSuitableAnchors {
					elementsNotFound["borders"] = append(elementsNotFound["borders"], border.Kilometrierung.Value)
					continue
				}
				return errors.Wrap(err, "failed to map border "+border.Kilometrierung.Value)
			}

			newSignalNode := createNewBorder(
				nodeIdCounter,
				maxNode,
			)
			OSMUtil.InsertNewNodeWithReferenceNode(
				osmData,
				&newSignalNode,
				maxNode,
			)
		}
	}
	return nil
}

// mapBumper processes all bumpers.
func mapTrackEnd(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	abschnitt Spurplanabschnitt,
	elementsNotFound map[string]([]string),
) error {
	for _, knoten := range abschnitt.Knoten {
		for _, border := range knoten.BetriebsStGr {
			kilometrage, _ := formatKilometrageStringInFloat(border.KnotenTyp.Kilometrierung.Value)

			maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
			if err != nil {
				if errors.Cause(err) == errNoSuitableAnchors {
					elementsNotFound["track ends"] = append(elementsNotFound["track ends"], border.Kilometrierung.Value)
					continue
				}
				return errors.Wrap(err, "failed to map track end "+border.Kilometrierung.Value)
			}

			newSignalNode := createNewTrackEnd(
				nodeIdCounter,
				maxNode,
			)
			OSMUtil.InsertNewNodeWithReferenceNode(
				osmData,
				&newSignalNode,
				maxNode,
			)
		}
	}
	return nil
}
