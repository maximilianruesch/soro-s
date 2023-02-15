package DBUtils

import(
	"math"
	"errors"
	"strconv"
	"fmt"
	OSMUtil "transform-osm/osm-utils"
)

const r = 6371.0
var osmData OSMUtil.Osm

func FindNewNode(node1 OSMUtil.Node, node2 OSMUtil.Node, dist1 float64, dist2 float64, data OSMUtil.Osm) (node OSMUtil.Node) {
	osmData = data

	if dist1 == 0.0 {
		return node1
	}

	up1, down1 := findNodes(node1, dist1)
	up2, down2 := findNodes(node2, dist2)

	if up1 == up2 {
		node, _ = getNode(up1)
	} else if up1 == down2 {
		node, _ = getNode(up1)
	} else if down1 == up2 {
		node, _ = getNode(down1)
	} else if down1 == down2 {
		node, _ = getNode(down1)
	} else {/*
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
		
		fmt.Printf("%s, %s, %s, %s \n", upNode1.Lat, upNode2.Lat, upNode1.Lon, upNode2.Lon)
		fmt.Printf("%s, %s, %s, %s \n", upNode1.Lat, downNode2.Lat, upNode1.Lon, downNode2.Lon)
		fmt.Printf("%s, %s, %s, %s \n", downNode1.Lat, upNode2.Lat, downNode1.Lon, downNode2.Lon)
		fmt.Printf("%s, %s, %s, %s \n", downNode1.Lat, downNode2.Lat, downNode1.Lon, downNode2.Lon)
		*/
		panic(errors.New("Fail"))
	}
	return 
}

func findNodes(node OSMUtil.Node, dist float64) (upId string, downId string) {
	startWay, err := findWay(node.Id)
	if err != nil {
		panic(err)
	}	

	switch (len(startWay)) {
	case 1:
		runningWay := startWay[0]
		index := getIndex(node.Id, runningWay)
		upId = goUp(runningWay, index, dist)
		downId = goDown(runningWay, index, dist)
	case 2:
		if startWay[0].Nd[0].Ref == node.Id && startWay[1].Nd[len(startWay[1].Nd)-1].Ref == node.Id {
			runningWay := startWay[1]
			upId = goUp(runningWay, len(runningWay.Nd)-1, dist)
			runningWay = startWay[0]
			downId = goDown(runningWay, 0, dist)
		} else if startWay[1].Nd[0].Ref == node.Id && startWay[0].Nd[len(startWay[0].Nd)-1].Ref == node.Id {
			runningWay := startWay[0]
			upId = goUp(runningWay, len(runningWay.Nd)-1, dist)
			runningWay = startWay[1]
			downId = goDown(runningWay, 0, dist)
		} else {
			panic(errors.New("Error with ways!"))
		}
	default: 
		panic(errors.New("Too many ways!"))
	}

	return
}

func goUp(runningWay OSMUtil.Way, index int, dist float64) string {	
	runningNode, _ := getNode(runningWay.Nd[index].Ref)
	totalDist := 0.0

	for ; totalDist < dist; {
		if index == 0 {
			nextWays, err := findWay(runningNode.Id)
			if err != nil || len(nextWays) != 2 {
				panic(errors.New("Wrong number of ways!"))
			}

			if nextWays[0].Nd[0].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[0].Ref == runningNode.Id {
				runningWay = nextWays[1]
				index = len(nextWays[1].Nd)-1
			} else if nextWays[1].Nd[0].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[0].Ref == runningNode.Id {
				runningWay = nextWays[0]
				index = len(nextWays[0].Nd)-1
			} else {
				panic(errors.New("Could not find way!"))
			}
		} 
		nextNode, err := getNode(runningWay.Nd[index-1].Ref)
		if err != nil {
			panic(err)
		}

		phi1, _ := strconv.ParseFloat(runningNode.Lat, 64)
		phi2, _ := strconv.ParseFloat(nextNode.Lat, 64)
		lambda1, _ := strconv.ParseFloat(runningNode.Lon, 64)
		lambda2, _ := strconv.ParseFloat(nextNode.Lon, 64)
		
		newDist := distance(phi1, phi2, lambda1, lambda2)
		fmt.Printf("%f between %s:%f,%f and %s:%f,%f \n", newDist, runningNode.Id, phi1, lambda1, nextNode.Id, phi2, lambda2)

		totalDist += newDist
		fmt.Printf("%f of %f \n", totalDist, dist)

		if totalDist == dist {
			print("Dist is gud \n")
			return runningWay.Nd[index].Ref			
		}

		index--
		runningNode = nextNode
	}
	return runningNode.Id
}

func goDown(runningWay OSMUtil.Way, index int, dist float64) string {	
	runningNode, _ := getNode(runningWay.Nd[index].Ref)
	totalDist := 0.0

	for ; totalDist < dist; {
		if index == len(runningWay.Nd)-1 {
			nextWays, err := findWay(runningNode.Id)
			if err != nil || len(nextWays) != 2 {
				panic(errors.New("Wrong number of ways!"))
			}

			if nextWays[0].Nd[0].Ref == nextWays[1].Nd[len(nextWays[1].Nd)-1].Ref && nextWays[0].Nd[0].Ref == runningNode.Id {
				runningWay = nextWays[0]
				index = 0
			} else if nextWays[1].Nd[0].Ref == nextWays[0].Nd[len(nextWays[0].Nd)-1].Ref && nextWays[1].Nd[0].Ref == runningNode.Id {
				runningWay = nextWays[1]
				index = 0
			} else {
				panic(errors.New("Could not find way!"))
			}
		} 
		nextNode, err := getNode(runningWay.Nd[index+1].Ref)
		if err != nil {
			panic(err)
		}

		phi1, _ := strconv.ParseFloat(runningNode.Lat, 64)
		phi2, _ := strconv.ParseFloat(nextNode.Lat, 64)
		lambda1, _ := strconv.ParseFloat(runningNode.Lon, 64)
		lambda2, _ := strconv.ParseFloat(nextNode.Lon, 64)
		
		newDist := distance(phi1, phi2, lambda1, lambda2)
		fmt.Printf("%f between %s:%f,%f and %s:%f,%f \n", newDist, runningNode.Id, phi1, lambda1, nextNode.Id, phi2, lambda2)

		totalDist += newDist
		fmt.Printf("%f of %f \n", totalDist, dist)

		if totalDist == dist {
			print("Dist is gud \n")
			return runningWay.Nd[index].Ref			
		}

		index++
		runningNode = nextNode
	}
	return runningNode.Id
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