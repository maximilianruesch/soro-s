package osmUtils

import (
	"errors"
)

func nodeNotFound(id string) error       { return errors.New("Could not find node: " + id) }
func tagOnNodeNotFound(key string) error { return errors.New("Could not find tag on node: " + key) }
func wayNotFound(id string) error        { return errors.New("Could not find way: " + id) }

func FindTagOnNode(node *Node, key string) (string, error) {
	for _, tag := range node.Tag {
		if tag.K == key {
			return tag.V, nil
		}
	}

	return "", tagOnNodeNotFound(key)
}

func GetNodeById(osm *Osm, id string) (*Node, error) {
	for _, node := range osm.Node {
		if node.Id == id {
			return node, nil
		}
	}
	return nil, nodeNotFound(id)
}

func GetNodeIndexInWay(way *Way, id string) (int, error) {
	for i, nd := range way.Nd {
		if nd.Ref == id {
			return i, nil
		}
	}
	return -1, nodeNotFound(id)
}

func FindWaysByNodeId(osm *Osm, id string) ([]Way, error) {
	ways := []Way{}
	for _, way := range osm.Way {
		for _, node := range way.Nd {
			if node.Ref == id {
				ways = append(ways, *way)
				break
			}
		}
	}
	if len(ways) == 0 {
		return []Way{}, wayNotFound(id)
	}
	return ways, nil
}

func InsertSignalWithWayRef(osm *Osm, signalNode *Node, nodeBeforeId string) {
	osm.Node = append(osm.Node, signalNode)

	for _, way := range osm.Way {
		index := -1
		for i, nd := range way.Nd {
			if nd.Ref == nodeBeforeId {
				index = i
				break
			}
		}
		if index == -1 {
			return
		}
		if index == len(way.Nd)-1 {
			element := way.Nd[index]
			temp := append(way.Nd[:index], &Nd{Ref: signalNode.Id})
			way.Nd = append(temp, element)
			return
		}
		temp := append(way.Nd[:index+1], &Nd{Ref: signalNode.Id})
		way.Nd = append(temp, way.Nd[index+1:]...)
	}
}
