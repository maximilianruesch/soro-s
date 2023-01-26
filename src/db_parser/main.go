package main

import (
	"encoding/xml"
	"os"
	"log"
	Util "db-parse/DBUtils"
)

func main() {
	const line string = "3601"

	file, err := os.ReadFile("Spurplanbetriebsstellen_FB__S-FD.xml")
	if err != nil {
		log.Fatal(err)
	}

	var input_data Util.XmlIssDaten

	if err := xml.Unmarshal([]byte(file), &input_data); err != nil {
		panic(err)
	}

	var output_data Util.XmlIssDaten
	output_data.Betriebsstellen = []*Util.Spurplanbetriebsstelle{}

	var i int
	i = 0

	var used bool
	used = false

	for _, stelle := range input_data.Betriebsstellen {	
		for _, abschnitt := range stelle.Abschnitte {			
			if nr := (*abschnitt.Strecken_Nr[0]).Nummer; nr == line {
				if len(output_data.Betriebsstellen) == i {
					output_data.Betriebsstellen = append(output_data.Betriebsstellen, &Util.Spurplanbetriebsstelle{stelle.XMLName, stelle.Name, []*Util.Spurplanabschnitt{}})
					used = true
				}				
				output_data.Betriebsstellen[i].Abschnitte = append(output_data.Betriebsstellen[i].Abschnitte, abschnitt)
			}			
		}
		if used {
			i++
			used = false
		}
	}

	var new_Data []byte 
	

	if new_Data, err = xml.MarshalIndent(output_data, "", "	"); err != nil {
		panic(err)
	} else {
		if err := os.WriteFile("new_data.xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
			panic(err)
		}
	}
}