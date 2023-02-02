package DBParser

import (
	"encoding/xml"
	"os"
	"log"
	"fmt"
	Util "transform-osm/db-utils"
)

func Parse(refs []string) []string {
	const resourceDir = "db-parser/resources"
	const tempDir = "DBLines"

	// read all files and unmarshal them into one XmlIssDaten-struct
	files, err := os.ReadDir(resourceDir)
	if err != nil {
		log.Fatal(err)
	}
	var input_data Util.XmlIssDaten	
	for _, file := range files {
		data, _ := os.ReadFile(resourceDir+"/"+file.Name())
		fmt.Printf("Processing %s... \r", file.Name())
		
		if err := xml.Unmarshal([]byte(data), &input_data); err != nil { 
			panic(err)	
		}
	}

	// these three maps all map XmlIssDaten, indices and "valid-bits" onto the line-numbers.
	// This is due to bookkeeping concerns.
	var lineMap map[string]Util.XmlIssDaten
	lineMap = make(map[string]Util.XmlIssDaten)
	var indexMap map[string]int
	indexMap = make(map[string]int)
	var usedMap map[string]bool
	usedMap = make(map[string]bool)
	// in missingMap, all lines, for which DB-data exists but no OSM-data (i.e. not appearing in refs) are listed
	var missingMap map[string]bool
	missingMap = make(map[string]bool)

	// all datastructures are being intialized
	for _, line := range refs {
		lineMap[line] = Util.XmlIssDaten{xml.Name{" ", "XmlIssDaten"}, []*Util.Spurplanbetriebsstelle{}}

		indexMap[line] = 0
		usedMap[line] = false
	}

	// main work-loop: For all "Betriebsstellen" and for all "Spurplanabschnitte" of these, we check, whether the respective line
	// appears in refs and if so add the "Abschnitt" to the "Betriebsstelle" in the respective line
	for _, stelle := range input_data.Betriebsstellen {	
		for _, abschnitt := range stelle.Abschnitte {
			abschnitt_nummer := (*abschnitt.Strecken_Nr[0]).Nummer
			temp, ok := lineMap[abschnitt_nummer]
			
			if !ok {
				missingMap[abschnitt_nummer] = true
				continue
			} 

			i := indexMap[abschnitt_nummer]
			// if no "Abschnitt" has yet been added to this particular "Betriebsstelle", we must first create one
			if len(temp.Betriebsstellen) == i {
				temp.Betriebsstellen = append(temp.Betriebsstellen, &Util.Spurplanbetriebsstelle{stelle.XMLName, stelle.Name, []*Util.Spurplanabschnitt{}})
				usedMap[abschnitt_nummer] = true
			}
			temp.Betriebsstellen[i].Abschnitte = append(temp.Betriebsstellen[i].Abschnitte, abschnitt)
			lineMap[abschnitt_nummer] = temp // write-back due to not being a pointer...
		}

		// final increment of all "Betriebsstellen"-counters where neccessary
		for key, used := range usedMap {
			if used {
				indexMap[key] += 1
				usedMap[key] = false
			}
		}
	}
	
	relevant_refs := []string{}
	os.Mkdir("temp/"+tempDir+"/", 0755)
	var new_Data []byte 	
	//final work-loop: For all collected lines, .xml-files must be marshelled
	for line, data := range lineMap {
		if (len(data.Betriebsstellen) == 0) {
			continue
		}

		if new_Data, err = xml.MarshalIndent(data, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile("temp/"+tempDir+"/"+line+"_DB.xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
				panic(err)
			}
			relevant_refs = append(relevant_refs, line)
		}
	}	
	
	return relevant_refs
}