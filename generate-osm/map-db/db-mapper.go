package DBMapper 

import {
	"encoding/xml"
	"os"
	"log"
	"fmt"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
}

func MapDB([]string refs, string osmDir, string DBDir) {
	for _, line := range refs {	
		var mappedItems map[string]OSMUtil.Node
		mappedItems = make(map[string]OSMUtil.Node)
		
		var osmData OSMUtil.Osm
		var dbData DBUtil.XmlIssDaten

		osm_file, err := os.ReadFile(osmDir+"/"+line+".xml")
		if err != nil {
			log.Fatal(err)
		}
		db_file, err := os.ReadFile(osmDir+"/"+line+"_DB.xml")
		if err != nil {
			log.Fatal(err)
		}

		if err := xml.Unmarshal([]byte(osm_file), &osmData); err != nil { 
			panic(err)	
		}
		if err := xml.Unmarshal([]byte(db_file), &dbData); err != nil { 
			panic(err)	
		}

		/*
		mapSignals(&osmData, dbData, &mappedItems)
		mapPoints(&osmData, dbData, &mappedItems)
		mapRest(&osmData, dbData, &mappedItems) 
		*/
	}
}