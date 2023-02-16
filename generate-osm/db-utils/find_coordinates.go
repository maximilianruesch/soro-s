package DBUtils

import(
	"math"
	"errors"
	"strconv"
	"sort"
	"fmt"
	OSMUtil "transform-osm/osm-utils"
)

type nodePair struct {
	node1 OSMUtil.Node
	node2 OSMUtil.Node
	dist float64
	remDist1 float64
	remDist2 float64
}

const r = 6371.0
var osmData OSMUtil.Osm

func FindNewNode(node1 OSMUtil.Node, node2 OSMUtil.Node, dist1 float64, dist2 float64, data OSMUtil.Osm) (node OSMUtil.Node, err error) {
	osmData = data
	err = nil

	if dist1 == 0.0 {
		return node1, nil
	}

	up1, upDist1, down1, downDist1, err1 := findNodes(node1, dist1)
	up2, upDist2, down2, downDist2, err2 := findNodes(node2, dist2)

	if err1 != nil || err2 != nil {
		return OSMUtil.Node{}, errors.New("Insufficient anchor!")
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
			nodePair{upNode1, upNode2, distUp1Up2, upDist1, upDist2},
			nodePair{upNode1, downNode1, distUp1Down2, upDist1, downDist2},
			nodePair{downNode1, upNode2, distDown1Up2, downDist1, upDist2},
			nodePair{downNode1, downNode2, distDown1Down2, downDist1, downDist2}} 

		sort.Slice(allPairs, func(i, j int) bool {
			dist1 := allPairs[i].dist
			dist2 := allPairs[j].dist
			return dist1 < dist2
		})

		if allPairs[0].remDist1 <= allPairs[0].remDist2 {
			node = allPairs[0].node1
		} else {
			node = allPairs[0].node2
		}
	}
	return 
}

func findNodes(node OSMUtil.Node, dist float64) (upId string, upDist float64, downId string, downDist float64, err error) {
	startWay, err := findWay(node.Id)
	if err != nil {
		panic(err)
	}	

	fmt.Printf("node: %s, length: %d \n", node.Id, len(startWay))

	switch (len(startWay)) {
	case 1:
		runningWay := startWay[0]
		index := getIndex(node.Id, runningWay)
		upId, upDist = goUp(runningWay, index, dist)
		downId, downDist = goDown(runningWay, index, dist)
	case 2:
		if startWay[0].Nd[0].Ref == node.Id && startWay[1].Nd[len(startWay[1].Nd)-1].Ref == node.Id {
			runningWay := startWay[1]
			upId, upDist = goUp(runningWay, len(runningWay.Nd)-1, dist)
			runningWay = startWay[0]
			downId, downDist = goDown(runningWay, 0, dist)
		} else if startWay[1].Nd[0].Ref == node.Id && startWay[0].Nd[len(startWay[0].Nd)-1].Ref == node.Id {
			runningWay := startWay[0]
			upId, upDist = goUp(runningWay, len(runningWay.Nd)-1, dist)
			runningWay = startWay[1]
			downId, downDist = goDown(runningWay, 0, dist)
		} else {
			panic(errors.New("Error with ways!"))
		}
	default: 
		err = errors.New("Too many ways!")
		return
	}
	err = nil
	return 
}

func goUp(runningWay OSMUtil.Way, index int, dist float64) (string, float64) {	
	runningNode, _ := getNode(runningWay.Nd[index].Ref)
	var oldNode OSMUtil.Node
	var nextNode OSMUtil.Node
	totalDist := 0.0
	wayDirUp := true

	for ; totalDist < dist; {
		runningWay, index, wayDirUp, nextNode = findNextWay(wayDirUp, index, runningNode, oldNode, runningWay)

		if index == -1 {
			return runningNode.Id, totalDist
		}
		
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

func goDown(runningWay OSMUtil.Way, index int, dist float64) (string, float64) {	
	runningNode, _ := getNode(runningWay.Nd[index].Ref)	
	var oldNode OSMUtil.Node
	var nextNode OSMUtil.Node
	totalDist := 0.0
	wayDirUp := false

	for ; totalDist < dist; {
		runningWay, index, wayDirUp, nextNode = findNextWay(wayDirUp, index, runningNode, oldNode, runningWay)

		if index == -1 {
			return runningNode.Id, totalDist
		}

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

func findNextWay(wayDirUp bool, index int, runningNode OSMUtil.Node, oldNode OSMUtil.Node, runningWay OSMUtil.Way) (OSMUtil.Way, int, bool, OSMUtil.Node) {
	if wayDirUp && index == 0 {
		nextWays, err := findWay(runningNode.Id)
		if err != nil || len(nextWays) == 0 {
			panic(errors.New("No ways!"))
		}
		if len(nextWays) == 1 || len(nextWays) > 2 {
			return runningWay, -1, false, OSMUtil.Node{}
		}

		if nextWays[0].Nd[0].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[0].Ref == runningNode.Id {
			runningWay = nextWays[1]
			index = len(nextWays[1].Nd)-1
			wayDirUp = true
		} else if nextWays[1].Nd[0].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[0].Ref == runningNode.Id {
			runningWay = nextWays[0]
			index = len(nextWays[0].Nd)-1
			wayDirUp = true
		} else if nextWays[0].Nd[0].Ref == nextWays[1].Nd[0].Ref && nextWays[0].Nd[0].Ref == runningNode.Id && nextWays[0].Nd[1].Ref == oldNode.Id {
			runningWay = nextWays[1]
			index = 0
			wayDirUp = false
		} else if nextWays[1].Nd[0].Ref == nextWays[0].Nd[0].Ref && nextWays[1].Nd[0].Ref == runningNode.Id && nextWays[1].Nd[1].Ref == oldNode.Id {
			runningWay = nextWays[0]
			index = 0
			wayDirUp = false
		} else {
			panic(errors.New("Could not find way!"))
		}
	} else if !wayDirUp && index == len(runningWay.Nd)-1 {
		nextWays, err := findWay(runningNode.Id)
		if err != nil || len(nextWays) == 0 {
			panic(errors.New("No ways!"))
		}
		if len(nextWays) == 1 || len(nextWays) > 2 {
			return runningWay, -1, false, OSMUtil.Node{}
		}

		if nextWays[0].Nd[0].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[0].Ref == runningNode.Id {
			runningWay = nextWays[0]
			index = 0
			wayDirUp = false
		} else if nextWays[1].Nd[0].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[0].Ref == runningNode.Id {
			runningWay = nextWays[1]
			index = 0
			wayDirUp = false
		} else if nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref == runningNode.Id && nextWays[0].Nd[len(nextWays[0].Nd)-2].Ref == oldNode.Id {
			runningWay = nextWays[1]
			index = len(nextWays[1].Nd)-1
			wayDirUp = true
		} else if nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref == runningNode.Id && nextWays[1].Nd[len(nextWays[1].Nd)-2].Ref == oldNode.Id {
			runningWay = nextWays[0]
			index = len(nextWays[0].Nd)-1
			wayDirUp = true
		} else {
			panic(errors.New("Could not find way!"))
		}
	}

	if wayDirUp {
		nextNode, err := getNode(runningWay.Nd[index-1].Ref)
		if err != nil {
			panic(err)
		}
		return runningWay, index, wayDirUp, nextNode
	} else {
		nextNode, err := getNode(runningWay.Nd[index+1].Ref)
		if err != nil {
			panic(err)
		}
		return runningWay, index, wayDirUp, nextNode
	}
}

func getNode(id string) (OSMUtil.Node, error){
	for _, node := range osmData.Node {
		if node.Id == id {
			return *node, nil
		}
	}
	return OSMUtil.Node{}, errors.New("Could not find node!")
}

func getIndex(id string, way OSMUtil.Way) int {
	for i, nd := range way.Nd {
		if nd.Ref == id {
			return i
		}
	}
	return -1
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
		return []OSMUtil.Way{}, errors.New("Could not find way!")
	}
	return ways, nil
}

func distance(phi1 float64, phi2 float64, lambda1 float64, lambda2 float64) float64 {
	phi1, phi2, lambda1, lambda2 = phi1*(math.Pi / 180.0), phi2*(math.Pi / 180.0), lambda1*(math.Pi / 180.0), lambda2*(math.Pi / 180.0)
	return 2.0*r*math.Asin(
		math.Sqrt(
			math.Pow(math.Sin((phi2 - phi1)/2), 2) +
			math.Cos(phi1)*math.Cos(phi2)*math.Pow(math.Sin((lambda2 - lambda1)/2), 2)))
}