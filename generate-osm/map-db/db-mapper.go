package DBMapper 

import (
	"encoding/xml"
	"os"
	"log"
	"strconv"
	"strings"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
)

var TagName = xml.Name{" ", "tag"}
var id_counter = 1

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {	
		var mappedItems map[string](*OSMUtil.Node)
		mappedItems = make(map[string](*OSMUtil.Node))
		
		var osmData OSMUtil.Osm
		var dbData DBUtil.XmlIssDaten

		osm_file, err := os.ReadFile(osmDir+"/"+line+".xml")
		if err != nil {
			log.Fatal(err)
		}
		db_file, err := os.ReadFile(DBDir+"/"+line+"_DB.xml")
		if err != nil {
			log.Fatal(err)
		}

		if err := xml.Unmarshal([]byte(osm_file), &osmData); err != nil { 
			panic(err)	
		}
		if err := xml.Unmarshal([]byte(db_file), &dbData); err != nil { 
			panic(err)	
		}

		print("Processing line ")
		print(line)
		print("\n")

		mapSignals(&osmData, dbData, &mappedItems)
		// mapPoints(&osmData, dbData, &mappedItems)
		// mapRest(&osmData, dbData, &mappedItems) 

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
}

func mapSignals(OSMData *OSMUtil.Osm, DBData DBUtil.XmlIssDaten, anchors *map[string](*OSMUtil.Node)) {
	for _, stelle := range DBData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				processHauptsigF(*knoten, OSMData, anchors)
				processHauptsigS(*knoten, OSMData, anchors)
			}
		}
	}
}

func processHauptsigF(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string](*OSMUtil.Node)) {
	for _, signal := range knoten.HauptsigF {
		
		if node := (*anchors)[signal.KnotenTyp.Kilometrierung[0].Value]; node != nil {
			typ, err1 := OSMUtil.GetTag(*node, "type")
			subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
			id, err3 := OSMUtil.GetTag(*node, "id")
			direction, err4 := OSMUtil.GetTag(*node, "direction")
			if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "falling" {
				continue
			}			
			newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
				&OSMUtil.Tag{TagName, "type", "element"}, 
				&OSMUtil.Tag{TagName, "subtype", "ms"}, 
				&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
				&OSMUtil.Tag{TagName, "direction", "falling"}}}
			OSMUtil.InsertNode(&newNode, node.Id, OSMData)
			//(*anchors)[signal.KnotenTyp.Kilometrierung[0].Value] = &newNode
			print("Added node \n")
			id_counter++
			continue 
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}
			is_signal := false
			has_correct_id := false
			for _, tag := range (*node).Tag {
				if tag.K == "railway" && tag.V == "signal" {
					is_signal = true
				}	
				if tag.K == "ref" && tag.V == signal.Name[0].Value {
					has_correct_id = true
				}	
			}

			if has_correct_id && is_signal {
				tags := &OSMData.Node[i].Tag 
				*tags = append(*tags, []*OSMUtil.Tag{
						&OSMUtil.Tag{TagName, "type", "element"}, 
						&OSMUtil.Tag{TagName, "subtype", "ms"}, 
						&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
						&OSMUtil.Tag{TagName, "direction", "falling"}}...)
				(*anchors)[signal.KnotenTyp.Kilometrierung[0].Value] = OSMData.Node[i]
			} 
		}
		// TODO: Node not found, find closest mapped Node and work from there
	}
}

func processHauptsigS(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string](*OSMUtil.Node)) {
	for _, signal := range knoten.HauptsigS {
		
		if node := (*anchors)[signal.KnotenTyp.Kilometrierung[0].Value]; node != nil {
			typ, err1 := OSMUtil.GetTag(*node, "type")
			subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
			id, err3 := OSMUtil.GetTag(*node, "id")
			direction, err4 := OSMUtil.GetTag(*node, "direction")
			if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "rising" {
				continue
			}	
			newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
				&OSMUtil.Tag{TagName, "type", "element"}, 
				&OSMUtil.Tag{TagName, "subtype", "ms"}, 
				&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
				&OSMUtil.Tag{TagName, "direction", "rising"}}}
			OSMUtil.InsertNode(&newNode, node.Id, OSMData)
			//(*anchors)[signal.KnotenTyp.Kilometrierung[0].Value] = &newNode
			id_counter++
			continue 
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}
			is_signal := false
			has_correct_id := false
			for _, tag := range (*node).Tag {
				if tag.K == "railway" && tag.V == "signal" {
					is_signal = true
				}	
				if tag.K == "ref" && strings.ReplaceAll(tag.V, " ", "") == signal.Name[0].Value {
					has_correct_id = true
				}	
			}

			if has_correct_id && is_signal {
				tags := &OSMData.Node[i].Tag 
				*tags = append(*tags, []*OSMUtil.Tag{
						&OSMUtil.Tag{TagName, "type", "element"}, 
						&OSMUtil.Tag{TagName, "subtype", "ms"}, 
						&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
						&OSMUtil.Tag{TagName, "direction", "rising"}}...)
				(*anchors)[signal.KnotenTyp.Kilometrierung[0].Value] = OSMData.Node[i]
			} 
		}
		// TODO: Node not found, find closest mapped Node and work from there
	}
}