package findNodes

import (
	"fmt"
	"math"
	"strings"
	OSMUtil "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

var ErrNoSuitableAnchors = errors.New("failed to find suitable anchors")

// FindBestOSMNode determines a pair of anchors based on which a new Node is searched.
// Based on those, a new Node is then determined.
func FindBestOSMNode(
	osmData *OSMUtil.Osm,
	anchors map[float64]([]*OSMUtil.Node),
	kilometrage float64,
) (*OSMUtil.Node, error) {
	sortedAnchors := SortAnchors(anchors, kilometrage)
	if len(sortedAnchors) < 2 {
		return nil, errors.Wrap(ErrNoSuitableAnchors, "less than two anchors found")
	}
	fmt.Printf("sortedAnchors: %v %v", sortedAnchors[0], sortedAnchors[1])
	nearest, secondNearest := sortedAnchors[0], sortedAnchors[1]

	anchor1 := (anchors[nearest])[0]
	anchor2 := (anchors[secondNearest])[0]
	distance1 := math.Abs(nearest - kilometrage)
	distance2 := math.Abs(secondNearest - kilometrage)
	fmt.Println("anchor1: ", anchor1)
	fmt.Println("anchor2: ", anchor2)
	fmt.Println("distance1: ", distance1)
	fmt.Println("distance2: ", distance2)
	newNode, err := findNewNode(
		osmData,
		anchor1,
		anchor2,
		distance1,
		distance2,
	)
	fmt.Println("newNode: ", newNode)
	if err == nil {
		return newNode, nil
	}

	newAnchorCounter := 2
	for err != nil && newAnchorCounter < len(sortedAnchors) {
		innerError := errors.Unwrap(err)
		errorParts := strings.Split(innerError.Error(), ": ")
		if errorParts[0] != "insufficient anchor" {
			return nil, errors.Wrap(err, "failed to find OSM-node")
		}

		faultyNodeID := strings.ReplaceAll(errorParts[1], " ", "")

		if faultyNodeID == anchor1.Id {
			nearest = sortedAnchors[newAnchorCounter]
			anchor1 = (anchors[nearest])[0]
			distance1 = math.Abs(nearest - kilometrage)
			newAnchorCounter++
		} else {
			secondNearest = sortedAnchors[newAnchorCounter]
			anchor2 = (anchors[secondNearest])[0]
			distance2 = math.Abs(secondNearest - kilometrage)
			newAnchorCounter++
		}
		newNode, err = findNewNode(
			osmData,
			anchor1,
			anchor2,
			distance1,
			distance2,
		)
	}

	if newAnchorCounter == len(sortedAnchors) {
		return nil, ErrNoSuitableAnchors
	}

	return newNode, nil
}
