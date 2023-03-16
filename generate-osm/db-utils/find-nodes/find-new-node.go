package findNodes

import (
	"fmt"
	OSMUtil "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

// findNewNode takes two Nodes and a desired distance to both Nodes and computes an optimal Node in the provided 'osmData'.
// The returned Node satisfies the distance as best as possible.
func findNewNode(
	osmData *OSMUtil.Osm,
	node1 *OSMUtil.Node,
	node2 *OSMUtil.Node,
	dist1 float64,
	dist2 float64,
) (*OSMUtil.Node, error) {

	if dist1 == 0.0 {
		return node1, nil
	}

	up1, upDist1, down1, downDist1, err1 := findNodes(osmData, node1, dist1)
	up2, upDist2, down2, downDist2, err2 := findNodes(osmData, node2, dist2)
	fmt.Println("up1: ", up1)
	fmt.Println("up2: ", up2)
	fmt.Println("down1: ", down1)
	fmt.Println("down2: ", down2)
	fmt.Println("err1: ", err1)
	fmt.Println("err2: ", err2)
	if err1 != nil {
		return nil, errors.Wrap(err1, "insufficient anchor: "+node1.Id)
	}
	if err2 != nil {
		return nil, errors.Wrap(err2, "insufficient anchor: "+node2.Id)
	}

	if up1 == up2 || up1 == down2 {
		return OSMUtil.GetNodeById(osmData, up1)
	} else if down1 == up2 || down1 == down2 {
		return OSMUtil.GetNodeById(osmData, down1)
	} else {
		return findClosestMatch(osmData, up1, up2, down1, down2, upDist1, upDist2, downDist1, downDist2)
	}
}
