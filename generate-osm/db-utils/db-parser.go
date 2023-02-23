package dbUtils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/html/charset"
)

func Parse(refs []string, tempDBLinesPath string, dbResourcesPath string) []string {
	combinedDBIss, err := readDBFiles(dbResourcesPath)
	if err != nil {
		panic(err)
	}

	// these three maps all map XmlIssDaten, indices and "valid-bits" onto the line-numbers.
	// This is due to bookkeeping concerns.
	var lineMap map[string]XmlIssDaten
	lineMap = make(map[string]XmlIssDaten)
	var indexMap map[string]int
	indexMap = make(map[string]int)
	var usedMap map[string]bool
	usedMap = make(map[string]bool)
	// in missingMap, all lines, for which DB-data exists but no OSM-data (i.e. not appearing in refs) are listed
	var missingMap map[string]bool
	missingMap = make(map[string]bool)

	// all datastructures are being intialized
	for _, line := range refs {
		lineMap[line] = XmlIssDaten{xml.Name{" ", "XmlIssDaten"}, []*Spurplanbetriebsstelle{}}

		indexMap[line] = 0
		usedMap[line] = false
	}

	// main work-loop: For all "Betriebsstellen" and for all "Spurplanabschnitte" of these, we check, whether the respective line
	// appears in refs and if so add the "Abschnitt" to the "Betriebsstelle" in the respective line
	for _, stelle := range combinedDBIss.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			abschnitt_nummer := (*abschnitt.StreckenNr[0]).Nummer
			temp, ok := lineMap[abschnitt_nummer]

			if !ok {
				missingMap[abschnitt_nummer] = true
				continue
			}

			i := indexMap[abschnitt_nummer]
			// if no "Abschnitt" has yet been added to this particular "Betriebsstelle", we must first create one
			if len(temp.Betriebsstellen) == i {
				temp.Betriebsstellen = append(temp.Betriebsstellen, &Spurplanbetriebsstelle{stelle.XMLName, stelle.Name, []*Spurplanabschnitt{}})
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

	relevantRefs := []string{}
	os.Mkdir(tempDBLinesPath, 0755)
	//final work-loop: For all collected lines, .xml-files must be marshelled
	for line, data := range lineMap {
		if len(data.Betriebsstellen) == 0 {
			continue
		}

		newIssBytes, err := xml.MarshalIndent(data, "", "	")
		if err != nil {
			panic(err)
		}

		tempLinePath := filepath.Join(tempDBLinesPath, line+"_DB.xml")
		err = os.WriteFile(tempLinePath, []byte(xml.Header+string(newIssBytes)), 0644)
		if err != nil {
			panic(err)
		}

		relevantRefs = append(relevantRefs, line)
	}

	return relevantRefs
}

func readDBFiles(dbResourcesPath string) (XmlIssDaten, error) {
	// read all files and unmarshal them into one XmlIssDaten-struct
	files, err := os.ReadDir(dbResourcesPath)
	if err != nil {
		return XmlIssDaten{}, err
	}
	var inputData XmlIssDaten
	for _, file := range files {
		fmt.Printf("Processing %s... \r", file.Name())
		data, _ := os.ReadFile(dbResourcesPath + "/" + file.Name())
		reader := bytes.NewReader(data)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&inputData)
		if err != nil {
			return XmlIssDaten{}, err
		}
	}

	return inputData, nil
}
