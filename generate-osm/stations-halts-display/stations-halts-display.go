package stationsHaltsDisplay

import (
	"encoding/xml"
	"os"
	osmUtils "transform-osm/osm-utils"
)

func StationsHaltsDisplay(stationsFile string) map[string]map[string]map[string]string {
	data, _ := os.ReadFile(stationsFile)
	var osmData osmUtils.Osm
	if err := xml.Unmarshal([]byte(data), &osmData); err != nil {
		panic(err)
	}

	stations := make(map[string]map[string]string)
	halts := make(map[string]map[string]string)
	for _, n := range osmData.Node {
		var name string = ""
		for _, t := range n.Tag {
			if t.K == "name" {
				name = t.V
			}

			if name != "" && t.K == "railway" {
				if t.V == "station" || t.V == "facility" {
					stations[n.Id] = map[string]string{
						"name": name,
						"lat":  n.Lat,
						"lon":  n.Lon,
					}
				}

				if t.V == "halt" {
					halts[n.Id] = map[string]string{
						"name": name,
						"lat":  n.Lat,
						"lon":  n.Lon,
					}
				}
			}
		}
	}

	return map[string]map[string]map[string]string{
		"stations": stations,
		"halts":    halts,
	}
}
