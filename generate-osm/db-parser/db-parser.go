package DBParser

import (
	"encoding/xml"
	"os"
	"log"
	"fmt"
	Util "transform-osm/db-parser/DBUtils"
)

func Parse(refs []string) {
	const resourceDir = "db-parser/resources"
	const tempDir = "DBLines"

	files, err := os.ReadDir(resourceDir)
	if err != nil {
		log.Fatal(err)
	}

	var input_data Util.XmlIssDaten	

	var i int
	var abschnitt_nummer string

	var lineMap map[string]Util.XmlIssDaten
	lineMap = make(map[string]Util.XmlIssDaten)
	var indexMap map[string]int
	indexMap = make(map[string]int)
	var usedMap map[string]bool
	usedMap = make(map[string]bool)

	var missingMap map[string]bool
	missingMap = make(map[string]bool)

	for _, line := range refs {
		lineMap[line] = Util.XmlIssDaten{xml.Name{" ", "XmlIssDaten"}, []*Util.Spurplanbetriebsstelle{}}

		indexMap[line] = 0
		usedMap[line] = false
	}

	for _, file := range files {
		data, _ := os.ReadFile(resourceDir+"/"+file.Name())
		fmt.Printf("Processing %s... \r", file.Name())
		
		if err := xml.Unmarshal([]byte(data), &input_data); err != nil { 
			panic(err)	
		}
	}

	for _, stelle := range input_data.Betriebsstellen {	
		for _, abschnitt := range stelle.Abschnitte {
			abschnitt_nummer = (*abschnitt.Strecken_Nr[0]).Nummer
			temp, ok := lineMap[abschnitt_nummer]
			
			if !ok {
				missingMap[abschnitt_nummer] = true
				continue
			} 

			i = indexMap[abschnitt_nummer]

			if len(temp.Betriebsstellen) == i {
				temp.Betriebsstellen = append(temp.Betriebsstellen, &Util.Spurplanbetriebsstelle{stelle.XMLName, stelle.Name, []*Util.Spurplanabschnitt{}})
				usedMap[abschnitt_nummer] = true
			}
			temp.Betriebsstellen[i].Abschnitte = append(temp.Betriebsstellen[i].Abschnitte, abschnitt)
			lineMap[abschnitt_nummer] = temp
		}

		for key, used := range usedMap {
			if used {
				indexMap[key] += 1
				usedMap[key] = false
			}
		}
	}

	print("Didn't find in OSM: [")
	for key, _ := range missingMap {
		fmt.Printf("%s, ", key)
	}
	print("] \n")
	
	os.Mkdir("temp/"+tempDir+"/", 0755)

	var new_Data []byte 	

	print("No DB-data available for: [")

	for line, data := range lineMap {
		if (len(data.Betriebsstellen) == 0) {
			fmt.Printf("%s, ", line)
			continue
		}
		if new_Data, err = xml.MarshalIndent(data, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile("temp/"+tempDir+"/"+line+"_DB.xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}	
	print("] \n")
}