package DBMapper 

import (
	"encoding/xml"
	"os"
	"log"
	//"fmt"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
)

var TagName = xml.Name{" ", "tag"}

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {	
		var mappedItems map[string]OSMUtil.Node
		mappedItems = make(map[string]OSMUtil.Node)
		
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

func mapSignals(OSMData *OSMUtil.Osm, DBData DBUtil.XmlIssDaten, anchors *map[string]OSMUtil.Node) {
	for _, stelle := range DBData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				processHauptsigF(*knoten, OSMData, anchors)
			}
		}
	}
}

func processHauptsigF(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string]OSMUtil.Node) {
	for _, signal := range knoten.HauptsigF {
		/*
		if (*anchors)[signal.KnotenTyp.Kilometrierung[0].Value] != 0 {
			continue // TODO: Check, if Node is correct
		}
		*/
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
			}
		}
		// TODO: Node not found, find closest mapped Node and work from there
	}
}