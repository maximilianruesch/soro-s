package stationsHaltsDisplay

import (
	osmUtils "transform-osm/osm-utils"
)

func StationsHaltsDisplay(stationData osmUtils.Osm, osmData *osmUtils.Osm) map[string]map[string]map[string]string {
	stations := make(map[string]map[string]string)
	halts := make(map[string]map[string]string)

	for _, n := range stationData.Node {
		var name string = ""
		for _, t := range n.Tag {
			if t.K == "name" {
				name = t.V
			}

			if name != "" && t.K == "railway" {
				if t.V == "station" || t.V == "facility" {
					n.Tag = append(n.Tag, &osmUtils.Tag{K: "type", V: "station"})
					stations[n.Id] = map[string]string{
						"name": name,
						"lat":  n.Lat,
						"lon":  n.Lon,
					}
				}

				if t.V == "halt" {					
					n.Tag = append(n.Tag, &osmUtils.Tag{K: "type", V: "element"})
					n.Tag = append(n.Tag, &osmUtils.Tag{K: "subtype", V: "hlt"})
					halts[n.Id] = map[string]string{
						"name": name,
						"lat":  n.Lat,
						"lon":  n.Lon,
					}
				}
			}
		}
	}
	
	var found bool
	for id := range stations {
		found = false
		for _, node := range osmData.Node {		
			if node.Id == id {
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "type", V: "station"})
				found = true
				break
			}
		}
		
		if !found {
			for _, node := range stationData.Node {
				if node.Id == id {
					osmData.Node = append(osmData.Node, node)
					break
				}
			}
		}
	}	

	for id := range halts {
		found = false
		for _, node := range osmData.Node {
			if node.Id == id {
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "type", V: "element"})
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "subtype", V: "hlt"})
				found = true
				break
			}
		}
		if !found {
			for _, node := range stationData.Node {
				if node.Id == id {
					osmData.Node = append(osmData.Node, node)
					break
				}
			}
		}
	}

	return map[string]map[string]map[string]string{
		"stations": stations,
		"halts":    halts,
	}
}
