package dbUtils

import (
	"strconv"

	OSMUtil "transform-osm/osm-utils"
)

// createNewHauptsignal creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:ms', 'id:(Signal name)' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewMainSignal(
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ms"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name.Value},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewSwitch creates a new node with the following tags:
// 'type:element', 'subtype:simple_switch' and 'id:...' where ... is the name of the provided switch.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewSwitch(
	id *int,
	node *OSMUtil.Node,
	switchBegin *Weichenanfang,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "simple_switch"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: switchBegin.Name.Value},
		},
	}
}

// createNewHauptsignal creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:ms', 'id:(Signal name)' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewApproachSignal(
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "as"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name.Value},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewHauptsignal creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:ms', 'id:(Signal name)' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewProtectionSignal(
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "ps"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: signal.Name.Value},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewHauptsignal creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:ms', 'id:(Signal name)' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewHalt(
	id *int,
	node *OSMUtil.Node,
	halt *Halteplatz,
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "hlt"},
			{XMLName: XML_TAG_NAME_CONST, K: "id", V: halt.Name.Value},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewBumper creates a new node with the following tags:
// 'type:element' and 'subtype:bumper'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewBumper(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "bumper"},
		},
	}
}

// createNewBorder creates a new node with the following tags:
// 'type:element' and 'subtype:border'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewBorder(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "border"},
		},
	}
}

// createNewBorder creates a new node with the following tags:
// 'type:element' and 'subtype:track_end'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewTrackEnd(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "track_end"},
		},
	}
}

// createNewKmJump creates a new node with the following tags:
// 'type:element' and 'subtype:km_jump'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewKmJump(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "km_jump"},
		},
	}
}

// createNewSpeedLimit creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:spl' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewSpeedLimit(
	id *int,
	node *OSMUtil.Node,
	halt *MaxGeschwindigkeit,
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "spl"},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewSlope creates a new node with the following tags:
// 'type:element' and 'subtype:slope'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewSlope(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "slope"},
		},
	}
}

// createNewTunnel creates a new node with the following tags:
// 'type:element' and 'subtype:tunnel'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewTunnel(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "tunnel"},
		},
	}
}

// createNewSpeedLimit creates a new OSM-Node with the following tags:
// 'type:element', 'subtype:spl' and 'direction:...' where ... depends on 'isFalling'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewEoTD(
	id *int,
	node *OSMUtil.Node,
	halt *MaxGeschwindigkeit,
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
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "eotd"},
			{XMLName: XML_TAG_NAME_CONST, K: "direction", V: directionString},
		},
	}
}

// createNewLineSwitch creates a new node with the following tags:
// 'type:element' and 'subtype:line_switch'.
// It also increments the "global" NodeIDCounter provided in 'id'.
func createNewLineSwitch(
	id *int,
	node *OSMUtil.Node,
) OSMUtil.Node {
	*id++

	return OSMUtil.Node{
		Id:  strconv.Itoa(*id),
		Lat: node.Lat,
		Lon: node.Lon,
		Tag: []*OSMUtil.Tag{
			{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
			{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "line_switch"},
		},
	}
}
