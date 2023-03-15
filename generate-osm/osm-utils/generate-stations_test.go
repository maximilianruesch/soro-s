package osmUtils_test

import (
	"testing"
	osmUtils "transform-osm/osm-utils"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, "testStation", searchFile.Stations["1"].Name, "Expected testStation, got %s", searchFile.Stations["1"].Name)
		assert.Equal(t, "1", searchFile.Stations["1"].Lat, "Expected 1, got %s", searchFile.Stations["1"].Lat)
		assert.Equal(t, "2", searchFile.Stations["1"].Lon, "Expected 2, got %s", searchFile.Stations["1"].Lon)

		assert.Equal(t, "testHalt", searchFile.Halts["2"].Name, "Expected testHalt, got %s", searchFile.Halts["2"].Name)
		assert.Equal(t, "3", searchFile.Halts["2"].Lat, "Expected 3, got %s", searchFile.Halts["2"].Lat)
		assert.Equal(t, "4", searchFile.Halts["2"].Lon, "Expected 4, got %s", searchFile.Halts["2"].Lon)

		for _, stationHaltNode := range stationHaltsOsm.Node {
			if stationHaltNode.Id == "1" {
				assert.Equal(t, "type", stationHaltNode.Tag[2].K, "Expected type, got %s", stationHaltNode.Tag[2].K)
				assert.Equal(t, "station", stationHaltNode.Tag[2].V, "Expected station, got %s", stationHaltNode.Tag[2].V)
			}
			if stationHaltNode.Id == "2" {
				assert.Equal(t, "type", stationHaltNode.Tag[2].K, "Expected type, got %s", stationHaltNode.Tag[2].K)
				assert.Equal(t, "element", stationHaltNode.Tag[2].V, "Expected element, got %s", stationHaltNode.Tag[2].V)
			}
		}
	})
}
