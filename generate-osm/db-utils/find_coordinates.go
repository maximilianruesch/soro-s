package dbUtils

import (
	"errors"
	"math"
	"sort"
	"strconv"
	OSMUtil "transform-osm/osm-utils"
)

type nodePair struct {
	node1    *OSMUtil.Node
	node2    *OSMUtil.Node
	dist     float64
	remDist1 float64
	remDist2 float64
}

const r = 6371.0

var osmData *OSMUtil.Osm

func nodeNotFound(id string) error { return errors.New("Could not find node: " + id) }
func wayNotFound(id string) error  { return errors.New("Could not find way: " + id) }

var endReached = errors.New("Reached end of track.")

func FindNewNode(node1 *OSMUtil.Node, node2 *OSMUtil.Node, dist1 float64, dist2 float64) (*OSMUtil.Node, error) {
	var node *OSMUtil.Node
	if dist1 == 0.0 {
		return node1, nil
	}

	up1, upDist1, down1, downDist1, err1 := findNodes(node1, dist1)
	up2, upDist2, down2, downDist2, err2 := findNodes(node2, dist2)

	if err1 != nil || err2 != nil {
		return nil, errors.New("Insufficient anchor!")
	}

	if up1 == up2 {
		node, _ = getNode(up1)
	} else if up1 == down2 {
		node, _ = getNode(up1)
	} else if down1 == up2 {
		node, _ = getNode(down1)
	} else if down1 == down2 {
		node, _ = getNode(down1)
	} else {
		node = getClosestMatch(up1, up2, down1, down2, upDist1, upDist2, downDist1, downDist2)
	}
	return node, nil
}

func getClosestMatch(up1, up2, down1, down2 string, upDist1, upDist2, downDist1, downDist2 float64) *OSMUtil.Node {
	upNode1, _ := getNode(up1)
	upNode2, _ := getNode(up2)
	downNode1, _ := getNode(down1)
	downNode2, _ := getNode(down2)

	upNode1Lat, _ := strconv.ParseFloat(upNode1.Lat, 64)
	upNode1Lon, _ := strconv.ParseFloat(upNode1.Lon, 64)
	downNode1Lat, _ := strconv.ParseFloat(downNode1.Lat, 64)
	downNode1Lon, _ := strconv.ParseFloat(downNode1.Lon, 64)

	upNode2Lat, _ := strconv.ParseFloat(upNode2.Lat, 64)
	upNode2Lon, _ := strconv.ParseFloat(upNode2.Lon, 64)
	downNode2Lat, _ := strconv.ParseFloat(downNode2.Lat, 64)
	downNode2Lon, _ := strconv.ParseFloat(downNode2.Lon, 64)

	distUp1Up2 := distance(upNode1Lat, upNode2Lat, upNode1Lon, upNode2Lon)
	distUp1Down2 := distance(upNode1Lat, downNode2Lat, upNode1Lon, downNode2Lon)
	distDown1Up2 := distance(downNode1Lat, upNode2Lat, downNode1Lon, upNode2Lon)
	distDown1Down2 := distance(downNode1Lat, downNode2Lat, downNode1Lon, downNode2Lon)

	var allPairs = []nodePair{
		{upNode1, upNode2, distUp1Up2, upDist1, upDist2},
		{upNode1, downNode1, distUp1Down2, upDist1, downDist2},
		{downNode1, upNode2, distDown1Up2, downDist1, upDist2},
		{downNode1, downNode2, distDown1Down2, downDist1, downDist2}}

	sort.Slice(allPairs, func(i, j int) bool {
		dist1 := allPairs[i].dist
		dist2 := allPairs[j].dist
		return dist1 < dist2
	})

	if allPairs[0].remDist1 <= allPairs[0].remDist2 {
		return allPairs[0].node1
	}
	return allPairs[0].node2
}

func findNodes(node *OSMUtil.Node, dist float64) (string, float64, string, float64, error) {
	var upId, downId string
	var upDist, downDist float64

	startWay, err := findWay(node.Id)
	if err != nil {
		panic(err)
	}

	if len(startWay) > 2 {
		return "", 0, "", 0, errors.New("Too many ways!")
	}

	if len(startWay) == 1 {
		runningWay := startWay[0]
		index, err := getIndex(node.Id, runningWay)
		if err != nil {
			panic(err)
		}
		upId, upDist = goDir(runningWay, index, dist, true)      // going "up" first
		downId, downDist = goDir(runningWay, index, dist, false) // then going "down"
		return upId, upDist, downId, downDist, nil
	}

	if startWay[0].Nd[0].Ref == node.Id && startWay[1].Nd[len(startWay[1].Nd)-1].Ref == node.Id {
		runningWay := startWay[1]
		upId, upDist = goDir(runningWay, len(runningWay.Nd)-1, dist, true)
		runningWay = startWay[0]
		downId, downDist = goDir(runningWay, 0, dist, false)
	} else if startWay[1].Nd[0].Ref == node.Id && startWay[0].Nd[len(startWay[0].Nd)-1].Ref == node.Id {
		runningWay := startWay[0]
		upId, upDist = goDir(runningWay, len(runningWay.Nd)-1, dist, true)
		runningWay = startWay[1]
		downId, downDist = goDir(runningWay, 0, dist, false)
	} else {
		return "", 0, "", 0, errors.New("Error with ways!")
	}

	return upId, upDist, downId, downDist, nil
}

func goDir(runningWay OSMUtil.Way, index int, dist float64, initialWayDirUp bool) (string, float64) {
	runningNode, _ := getNode(runningWay.Nd[index].Ref)
	var oldNode *OSMUtil.Node
	var nextNode *OSMUtil.Node
	totalDist := 0.0
	wayDirUp := initialWayDirUp

	var err error

	for totalDist < dist {
		if (wayDirUp && index == 0) || (!wayDirUp && index == len(runningWay.Nd)-1) {
			runningWay, index, wayDirUp, err = findNextWay(wayDirUp, index, runningNode, oldNode, runningWay)
			if err == endReached {
				return runningNode.Id, totalDist
			}
			if err != nil {
				panic(err)
			}
		}

		nextNode = getNextNode(wayDirUp, index, runningWay)

		phi1, _ := strconv.ParseFloat(runningNode.Lat, 64)
		phi2, _ := strconv.ParseFloat(nextNode.Lat, 64)
		lambda1, _ := strconv.ParseFloat(runningNode.Lon, 64)
		lambda2, _ := strconv.ParseFloat(nextNode.Lon, 64)

		totalDist += distance(phi1, phi2, lambda1, lambda2)

		if totalDist == dist {
			print("Dist is gud \n")
			return runningWay.Nd[index].Ref, totalDist
		}

		if wayDirUp {
			index--
		} else {
			index++
		}
		oldNode = runningNode
		runningNode = nextNode
	}
	return runningNode.Id, totalDist
}

func getNextNode(wayDirUp bool, index int, runningWay OSMUtil.Way) *OSMUtil.Node {
	if wayDirUp {
		nextNode, err := getNode(runningWay.Nd[index-1].Ref)
		if err != nil {
			panic(err)
		}
		return nextNode
	}

	nextNode, err := getNode(runningWay.Nd[index+1].Ref)
	if err != nil {
		panic(err)
	}
	return nextNode
}

func findNextWay(wayDirUp bool, index int, runningNode *OSMUtil.Node, oldNode *OSMUtil.Node, runningWay OSMUtil.Way) (OSMUtil.Way, int, bool, error) {
	nextWays, err := findWay(runningNode.Id)
	if err != nil || len(nextWays) == 0 {
		panic(errors.New("No ways!"))
	}
	if len(nextWays) == 1 {
		return OSMUtil.Way{}, 0, false, endReached
	}
	if len(nextWays) > 2 {
		return OSMUtil.Way{}, 0, false, endReached
	}

	// Ways can be "linked" in different ways. The usual ones are:
	// Index0 beginning links with index1 end
	wayConnection01 := nextWays[0].Nd[0].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[0].Ref == runningNode.Id
	// or Index1 beginning links with index0 end
	wayConnection10 := nextWays[1].Nd[0].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[0].Ref == runningNode.Id

	if wayDirUp && index == 0 {
		// More complicated ways, Ways can be "linked" include:
		// Index0 beginning links with Index1 beginning and Index0 is the way we are currently climbing up, as the second item is the old node
		wayConnection00comingFrom0 := nextWays[0].Nd[0].Ref == nextWays[1].Nd[0].Ref && nextWays[0].Nd[0].Ref == runningNode.Id && nextWays[0].Nd[1].Ref == oldNode.Id
		// or both beginnings link up, however we come from index1 [second item is also known!]
		wayConnection00comingFrom1 := nextWays[1].Nd[0].Ref == nextWays[0].Nd[0].Ref && nextWays[1].Nd[0].Ref == runningNode.Id && nextWays[1].Nd[1].Ref == oldNode.Id
		if wayConnection01 {
			runningWay = nextWays[1]
			index = len(nextWays[1].Nd) - 1
			wayDirUp = true
		} else if wayConnection10 {
			runningWay = nextWays[0]
			index = len(nextWays[0].Nd) - 1
			wayDirUp = true
		} else if wayConnection00comingFrom0 {
			runningWay = nextWays[1]
			index = 0
			wayDirUp = false
		} else if wayConnection00comingFrom1 {
			runningWay = nextWays[0]
			index = 0
			wayDirUp = false
		} else {
			return OSMUtil.Way{}, 0, false, errors.New("Could not find way up!")
		}
	} else if !wayDirUp && index == len(runningWay.Nd)-1 {
		// In the downward direction, ways can also be linked via the ends:
		// End-linkage and coming from Index0, as second-to-last is known as oldNode
		wayConnectionEndEndComingFrom0 := nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref == runningNode.Id && nextWays[0].Nd[len(nextWays[0].Nd)-2].Ref == oldNode.Id
		// or end-linkage and coming from Index1
		wayConnectionEndEndComingFrom1 := nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref == runningNode.Id && nextWays[1].Nd[len(nextWays[1].Nd)-2].Ref == oldNode.Id
		if wayConnection01 {
			runningWay = nextWays[0]
			index = 0
			wayDirUp = false
		} else if wayConnection10 {
			runningWay = nextWays[1]
			index = 0
			wayDirUp = false
		} else if wayConnectionEndEndComingFrom0 {
			runningWay = nextWays[1]
			index = len(nextWays[1].Nd) - 1
			wayDirUp = true
		} else if wayConnectionEndEndComingFrom1 {
			runningWay = nextWays[0]
			index = len(nextWays[0].Nd) - 1
			wayDirUp = true
		} else {
			return OSMUtil.Way{}, 0, false, errors.New("Could not find way down!")
		}
	}

	return runningWay, index, wayDirUp, nil
}

func getNode(id string) (*OSMUtil.Node, error) {
	for _, node := range osmData.Node {
		if node.Id == id {
			return node, nil
		}
	}
	return nil, nodeNotFound(id)
}

func getIndex(id string, way OSMUtil.Way) (int, error) {
	for i, nd := range way.Nd {
		if nd.Ref == id {
			return i, nil
		}
	}
	return -1, nodeNotFound(id)
}

func findWay(id string) ([]OSMUtil.Way, error) {
	ways := []OSMUtil.Way{}
	for _, way := range osmData.Way {
		for _, node := range way.Nd {
			if node.Ref == id {
				ways = append(ways, *way)
				break
			}
		}
	}
	if len(ways) == 0 {
		return []OSMUtil.Way{}, wayNotFound(id)
	}
	return ways, nil
}

func distance(phi1 float64, phi2 float64, lambda1 float64, lambda2 float64) float64 {
	phi1, phi2, lambda1, lambda2 = phi1*(math.Pi/180.0), phi2*(math.Pi/180.0), lambda1*(math.Pi/180.0), lambda2*(math.Pi/180.0)
	return 2.0 * r * math.Asin(
		math.Sqrt(
			math.Pow(math.Sin((phi2-phi1)/2), 2)+
				math.Cos(phi1)*math.Cos(phi2)*math.Pow(math.Sin((lambda2-lambda1)/2), 2)))
}

func SetOSMData(data *OSMUtil.Osm) {
	osmData = data
	return
}
