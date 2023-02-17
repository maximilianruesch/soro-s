package DBUtils 

import (
	"testing"

	OSMUtil "transform-osm/osm-utils"
)

func TestDistance(t *testing.T) {
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