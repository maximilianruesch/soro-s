package osmUtils_test

import (
	"testing"
	osmUtils "transform-osm/osm-utils"
)

func TestGenerateSearchFile(t *testing.T) {
	type args struct {
		osm osmUtils.Osm
	}
	test := struct {
		name string
		args args
	}{
		name: "test the generation of the search file and adding the stations to the osm data",
		args: args{
			osmUtils.Osm{
				Node: []*osmUtils.Node{
					{
						Id:  "1",
						Lat: "1",
						Lon: "2",
						Tag: []*osmUtils.Tag{
							{
								K: "name",
								V: "testStation",
							},
							{
								K: "railway",
								V: "station",
							},
						},
					},
					{
						Id:  "2",
						Lat: "3",
						Lon: "4",
						Tag: []*osmUtils.Tag{
							{
								K: "name",
								V: "testHalt",
							},
							{
								K: "railway",
								V: "halt",
							},
						},
					},
				},
			},
		},
	}

	t.Run(test.name, func(t *testing.T) {
		searchFile, stationHaltsOsm := osmUtils.GenerateOsmAndSearchFile(&test.args.osm)

		if searchFile.Stations["1"].Name != "testStation" {
			t.Errorf("Expected testStation, got %s", searchFile.Stations["1"].Name)
		}
		if searchFile.Stations["1"].Lat != "1" {
			t.Errorf("Expected 1, got %s", searchFile.Stations["1"].Lat)
		}
		if searchFile.Stations["1"].Lon != "2" {
			t.Errorf("Expected 2, got %s", searchFile.Stations["1"].Lon)
		}

		if searchFile.Halts["2"].Name != "testHalt" {
			t.Errorf("Expected testHalt, got %s", searchFile.Halts["2"].Name)
		}
		if searchFile.Halts["2"].Lat != "3" {
			t.Errorf("Expected 3, got %s", searchFile.Halts["2"].Lat)
		}
		if searchFile.Halts["2"].Lon != "4" {
			t.Errorf("Expected 4, got %s", searchFile.Halts["2"].Lon)
		}

		for _, stationHaltNode := range stationHaltsOsm.Node {
			if stationHaltNode.Id == "1" {
				if stationHaltNode.Tag[2].K != "type" || stationHaltNode.Tag[2].V != "station" {
					t.Errorf("Expected station, got %s", stationHaltNode.Tag[2].V)
				}
			}
			if stationHaltNode.Id == "2" {
				if stationHaltNode.Tag[2].K != "type" || stationHaltNode.Tag[2].V != "element" {
					t.Errorf("Expected element, got %s", stationHaltNode.Tag[2].V)
				}
				if stationHaltNode.Tag[3].K != "subtype" || stationHaltNode.Tag[3].V != "hlt" {
					t.Errorf("Expected hlt, got %s", stationHaltNode.Tag[3].V)
				}
			}
		}
	})
}
