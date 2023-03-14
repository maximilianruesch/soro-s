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
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
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

		newSignalNode := createSimpleNode(
			nodeIdCounter,
			maxNode,
			"bumper",
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapBumper processes all bumpers.
func mapBorder(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
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

		newSignalNode := createSimpleNode(
			nodeIdCounter,
			maxNode,
			"border",
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapTrackEnd processes all track ends.
func mapTrackEnd(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	for _, track_end := range knoten.Gleisende {
		kilometrage, _ := formatKilometrageStringInFloat(track_end.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["track ends"] = append(elementsNotFound["track ends"], track_end.Name.Value)
				continue
			}
			return errors.Wrap(err, "failed to map track end "+track_end.Name.Value)
		}

		newSignalNode := createNamedSimpleNode(
			nodeIdCounter,
			maxNode,
			"track_end",
			track_end.Name.Value,
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapKmJump processes all kilometrage jumps.
func mapKmJump(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	for _, jump := range knoten.KmSprungAnf {
		kilometrage, _ := formatKilometrageStringInFloat(jump.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["kilometrage jumps"] = append(elementsNotFound["kilometrage jumps"], jump.Kilometrierung.Value)
				continue
			}
			return errors.Wrap(err, "failed to map kilometrage jump "+jump.Kilometrierung.Value)
		}

		newSignalNode := createSimpleNode(
			nodeIdCounter,
			maxNode,
			"km_jump",
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapSpeedLimits processes all speed limits.
func mapSpeedLimits(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	err := searchSpeedLimit(
		osmData,
		anchors,
		nodeIdCounter,
		knoten,
		elementsNotFound,
		true)
	if err != nil {
		return errors.Wrap(err, "failed finding falling speed limit")
	}

	err = searchSpeedLimit(
		osmData,
		anchors,
		nodeIdCounter,
		knoten,
		elementsNotFound,
		false)
	if err != nil {
		return errors.Wrap(err, "failed finding rising speed limit")
	}
	return nil
}

// searchSpeedLimit searches for a Node, that best fits the speed limit to be mapped.
// This search is based on at least two anchored elements and their respective distance to the signal at hand.
// If no ore only one anchor could be identified, or all anchors are otherwise insufficient, no mapping can be done.
func searchSpeedLimit(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
	isFalling bool,
) error {
	speeds := knoten.MaxSpeedF
	if !isFalling {
		speeds = knoten.MaxSpeedS
	}

	for _, speed := range speeds {
		kilometrage, _ := formatKilometrageStringInFloat(speed.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["speed limits"] = append(elementsNotFound["speed limits"], speed.Kilometrierung.Value)
				continue
			}
			return errors.Wrap(err, "failed to map speed limit "+speed.Kilometrierung.Value)
		}

		newSignalNode := createDirectionalNode(
			nodeIdCounter,
			maxNode,
			"spl",
			isFalling,
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapSlopes processes all slopes.
func mapSlopes(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	for _, slope := range knoten.Neigung {
		kilometrage, _ := formatKilometrageStringInFloat(slope.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["slopes"] = append(elementsNotFound["slopes"], slope.Kilometrierung.Value)
				continue
			}
			return errors.Wrap(err, "failed to map slope "+slope.Kilometrierung.Value)
		}

		newSignalNode := createSimpleNode(
			nodeIdCounter,
			maxNode,
			"slope",
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapSlopes processes all slopes.
func mapTunnels(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	for _, tunnel := range knoten.Tunnel {
		kilometrage, _ := formatKilometrageStringInFloat(tunnel.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["tunnels"] = append(elementsNotFound["tunnels"], tunnel.Name.Value)
				continue
			}
			return errors.Wrap(err, "failed to map tunnel "+tunnel.Name.Value)
		}

		newSignalNode := createNamedSimpleNode(
			nodeIdCounter,
			maxNode,
			"tunnel",
			tunnel.Name.Value,
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapEoTDs processes all speed limits.
func mapEoTDs(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	err := searchEoTD(
		osmData,
		anchors,
		nodeIdCounter,
		knoten,
		elementsNotFound,
		true)
	if err != nil {
		return errors.Wrap(err, "failed finding falling end of train detector")
	}

	err = searchEoTD(
		osmData,
		anchors,
		nodeIdCounter,
		knoten,
		elementsNotFound,
		false)
	if err != nil {
		return errors.Wrap(err, "failed finding rising end of train detector")
	}
	return nil
}

// searchEoTD searches for a Node, that best fits the Signal to be mapped.
// This search is based on at least two anchored elements and their respective distance to the signal at hand.
// If no ore only one anchor could be identified, or all anchors are otherwise insufficient, no mapping can be done.
func searchEoTD(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
	isFalling bool,
) error {
	track_ends := append(knoten.FstrZugschlussstelleF, knoten.SignalZugschlussstelleF...)
	if !isFalling {
		track_ends = append(knoten.FstrZugschlussstelleS, knoten.SignalZugschlussstelleS...)
	}

	for _, eotd := range track_ends {
		kilometrage, _ := formatKilometrageStringInFloat(eotd.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["eotds"] = append(elementsNotFound["eotds"], eotd.Kilometrierung.Value)
				continue
			}
			return errors.Wrap(err, "failed to map eotd "+eotd.Kilometrierung.Value)
		}

		newSignalNode := createDirectionalNode(
			nodeIdCounter,
			maxNode,
			"eotd",
			isFalling,
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}

// mapLineSwitches processes all line switches.
func mapLineSwitches(
	osmData *OSMUtil.Osm,
	anchors *map[float64]([]*OSMUtil.Node),
	nodeIdCounter *int,
	knoten Spurplanknoten,
	elementsNotFound map[string]([]string),
) error {
	for _, line_switch := range knoten.Streckenwechsel0 {
		kilometrage, _ := formatKilometrageStringInFloat(line_switch.KnotenTyp.Kilometrierung.Value)

		maxNode, err := findBestOSMNode(osmData, anchors, kilometrage)
		if err != nil {
			if errors.Cause(err) == errNoSuitableAnchors {
				elementsNotFound["line switches"] = append(elementsNotFound["line switches"], line_switch.Kilometrierung.Value)
				continue
			}
			return errors.Wrap(err, "failed to map line switche "+line_switch.Kilometrierung.Value)
		}

		newSignalNode := createSimpleNode(
			nodeIdCounter,
			maxNode,
			"line_switch",
		)
		OSMUtil.InsertNewNodeWithReferenceNode(
			osmData,
			&newSignalNode,
			maxNode,
		)
	}
	return nil
}
