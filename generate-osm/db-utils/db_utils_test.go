package DBUtils 

import (
	"testing"

	OSMUtil "transform-osm/osm-utils"
)

func TestDistance(t *testing.T) {
	print("Testing distance function. \n")
	var simpleTestCases = [][]float64{ // {phi1, phi2, lambda1, lambda2, expectedResult}
		[]float64{50, 50, 8, 8, 0},
		[]float64{0, -180, 0, 0, 20015.086796020572}, // ~half earth circumference
		[]float64{0, 0, 90, 0, 10007.543398010284}} // ~quarter earth circumference
	
	var idemTestCases = [][]float64{ // {phi1_1, phi2_1, lambda1_1, labda2_1, phi1_2, phi2_2, lambda1_2, labda2_2}
		[]float64{50, 49, 8, 8, 49, 50, 8, 8}, 
		[]float64{0, -180, 0, 0, 0, 180, 0, 0},
		[]float64{0, 0, 90, 0, 0, 0, -90, 0},
		[]float64{0, 1, 0, 0, 1, 0, 0, 0},
		[]float64{0, 0, 1, 0, 0, 0, 0, 1}}
	
	for _, vals := range simpleTestCases {
		if dist := distance(vals[0], vals[1], vals[2], vals[3]); vals[4] != dist {
			t.Log("result was wrongly", dist)
			t.Fail()
		}
	}
	for _, vals := range idemTestCases {
		if dist1, dist2 := distance(vals[0], vals[1], vals[2], vals[3]), distance(vals[4], vals[5], vals[6], vals[7]); dist1 != dist2{
			t.Log("Function not symmetric!", dist1, dist2)
			t.Fail()
		}
	}
}

type testIndexTuple struct {
	way OSMUtil.Way
	id string
	expectedResult int
}

func TestGetIndex(t *testing.T) {
	print("Testing getIndex function. \n")
	testWay := OSMUtil.Way{Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref:"1"}, &OSMUtil.Nd{Ref:"2"}, &OSMUtil.Nd{Ref:"3"}}}

	var testCases = []testIndexTuple{
		{testWay, "1", 0},
		{testWay, "3", 2},
		{testWay, "42", -1},
		{OSMUtil.Way{}, "1", -1}}

	for _, vals := range testCases {
		if index := getIndex(vals.id, vals.way); index != vals.expectedResult {
			t.Log("Wrong index: ", vals.expectedResult, index)
			t.Fail()
		}
	}
}

type testGetNodeTuple struct {
	osmData OSMUtil.Osm
	id string
	expectedNode *OSMUtil.Node
	expectedError error
}

func TestGetNode(t *testing.T) {
	print("Testing getNode function. \n")
	testNode1 := OSMUtil.Node{Id:"1"}
	testNode3 := OSMUtil.Node{Id:"3"}
	testData := OSMUtil.Osm{Node: []*OSMUtil.Node{
		&testNode1, &OSMUtil.Node{Id:"2"}, &testNode3}}

	var testCases = []testGetNodeTuple{
		{testData, "1", &testNode1, nil},
		{testData, "3", &testNode3, nil},
		{testData, "42", nil, nodeNotFound("42")},
		{OSMUtil.Osm{}, "1", nil, nodeNotFound("1")}}	

	for _, vals := range testCases {
		SetOSMData(&vals.osmData)
		node, err := getNode(vals.id)
		if (err != nil || vals.expectedError != nil) && err.Error() != vals.expectedError.Error() {
			t.Log("Wrong error: ", err, vals.expectedError)
			t.Fail()
		}
		if node != vals.expectedNode {
			t.Log("Expected "+vals.expectedNode.Id+", got "+node.Id)
			t.Fail()
		}
	}
}

type testFindWayTuple struct {
	osmData OSMUtil.Osm
	id string
	expectedWays []OSMUtil.Way
	expectedError error
}

func TestFindWay(t *testing.T) {
	print("Testing findWay function. \n")
	testWay1 := OSMUtil.Way{Id:"1", Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref:"100"}, &OSMUtil.Nd{Ref:"101"}, &OSMUtil.Nd{Ref:"102"}}}
	testWay2 := OSMUtil.Way{Id:"2", Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref:"102"}, &OSMUtil.Nd{Ref:"103"}, &OSMUtil.Nd{Ref:"104"}}}
	testData1 := OSMUtil.Osm{Way:[]*OSMUtil.Way{&testWay1, &testWay2}}
	testData2 := OSMUtil.Osm{Way:[]*OSMUtil.Way{&testWay1, &OSMUtil.Way{}, &testWay2}}

	testCases := []testFindWayTuple{
		{testData1, "103", []OSMUtil.Way{testWay2}, nil},
		{testData1, "102", []OSMUtil.Way{testWay1, testWay2}, nil},
		{testData1, "42", []OSMUtil.Way{}, wayNotFound("42")},
		{testData2, "101", []OSMUtil.Way{testWay1}, nil},
		{testData2, "102", []OSMUtil.Way{testWay1, testWay2}, nil},
		{testData2, "42", []OSMUtil.Way{}, wayNotFound("42")},
		{OSMUtil.Osm{}, "100", []OSMUtil.Way{}, wayNotFound("100")}}

	for _, vals := range testCases {
		SetOSMData(&vals.osmData)
		ways, err := findWay(vals.id)
		if (err != nil || vals.expectedError != nil) && err.Error() != vals.expectedError.Error() {
			t.Log("Wrong error: ", err, vals.expectedError)
			t.Fail()
		}
		for i, _ := range ways {
			if ways[i].Id != vals.expectedWays[i].Id {
				t.Log("Wrong Id, got:", ways[i].Id)
				t.Fail()
			}
		}
	}
}

type testFindNextWayTuple struct {
	osmData OSMUtil.Osm
	currWayDirUp bool
	currIndex int
	currRunningNode *OSMUtil.Node
	oldNode *OSMUtil.Node
	currRunningWay OSMUtil.Way
	expectedRunningWay OSMUtil.Way
	expectedIndex int
	expectedWayDirUp bool
	expectedNextNode *OSMUtil.Node
}

func TestFindNextWay(t *testing.T) {	
	print("Testing findNextWay function. \n")
	testNode1 := OSMUtil.Node{Id: "1"}
	testNode2 := OSMUtil.Node{Id: "2"}
	testNode3 := OSMUtil.Node{Id: "3"}
	testNode4 := OSMUtil.Node{Id: "4"}
	testNode5 := OSMUtil.Node{Id: "5"}
	testNode6 := OSMUtil.Node{Id: "6"}
	testNode7 := OSMUtil.Node{Id: "7"}

	testWay1 := OSMUtil.Way{Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref: "1"}, &OSMUtil.Nd{Ref: "2"}, &OSMUtil.Nd{Ref: "3"}}}
	testWay2 := OSMUtil.Way{Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref: "3"}, &OSMUtil.Nd{Ref: "4"}}}
	testWay3 := OSMUtil.Way{Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref: "5"}, &OSMUtil.Nd{Ref: "4"}}}
	testWay4 := OSMUtil.Way{Nd: []*OSMUtil.Nd{
		&OSMUtil.Nd{Ref: "5"}, &OSMUtil.Nd{Ref: "6"}, &OSMUtil.Nd{Ref: "7"}}}

	testData1 := OSMUtil.Osm{
		Node: []*OSMUtil.Node{
			&testNode1, &testNode2, &testNode3, &testNode4, &testNode5, &testNode6, &testNode7},
		Way: []*OSMUtil.Way{
			&testWay1, &testWay2, &testWay3, &testWay4}}
	testData2 := OSMUtil.Osm{
		Node: []*OSMUtil.Node{
			&testNode1, &testNode2, &testNode3, &testNode4, &testNode5, &testNode6, &testNode7},
		Way: []*OSMUtil.Way{
			&testWay4, &testWay3, &testWay2, &testWay1}}

	testCases := []testFindNextWayTuple{
		{testData1, true, 1, &testNode2, &testNode3, testWay1, testWay1, 1, true, &testNode1}, // No searching of any next ways
		{testData1, false, 1, &testNode2, &testNode1, testWay1, testWay1, 1, false, &testNode3},
		{testData1, true, 0, &testNode1, &testNode2, testWay1, testWay1, -1, false, nil}, // Reached upper/lower end of ways
		{testData1, false, 2, &testNode7, &testNode6, testWay4, testWay4, -1, false, nil},
		{testData1, true, 0, &testNode3, &testNode4, testWay2, testWay1, 2, true, &testNode2}, // No change in direction when going up
		{testData2, true, 0, &testNode3, &testNode4, testWay2, testWay1, 2, true, &testNode2}, // independent of order of ways
		{testData1, false, 2, &testNode3, &testNode2, testWay1, testWay2, 0, false, &testNode4}, // No change in direction when going down
		{testData2, false, 2, &testNode3, &testNode2, testWay1, testWay2, 0, false, &testNode4},
		{testData1, true, 0, &testNode5, &testNode6, testWay4, testWay3, 0, false, &testNode4}, // Change of direction when going up
		{testData2, true, 0, &testNode5, &testNode6, testWay4, testWay3, 0, false, &testNode4},
		{testData1, false, 1, &testNode4, &testNode3, testWay2, testWay3, 1, true, &testNode5}, // Change of direction when going down
		{testData2, false, 1, &testNode4, &testNode3, testWay2, testWay3, 1, true, &testNode5}} 

	for _, vals := range testCases {
		SetOSMData(&vals.osmData)
		runningWay, index, wayDirUp, nextNode := findNextWay(vals.currWayDirUp, vals.currIndex, vals.currRunningNode, vals.oldNode, vals.currRunningWay)

		if runningWay.Id != vals.expectedRunningWay.Id {
			t.Log("Wrong way:", runningWay.Id)
			t.Fail()
		}
		if index != vals.expectedIndex {
			t.Log("Wrong index:", index)
			t.Fail()
		}
		if wayDirUp != vals.expectedWayDirUp {
			t.Log("Wrong direction:", wayDirUp)
			t.Fail()
		}
		if !(nextNode == nil && vals.expectedNextNode == nil) && nextNode.Id != vals.expectedNextNode.Id {
			t.Log("Wrong node:", nextNode.Id, vals.expectedNextNode.Id)
			t.Fail()
		}
	}
}