package dbUtils

import (
	"encoding/xml"
	"fmt"
	"os"
	osmUtils "transform-osm/osm-utils"
)

func MapDB(refs []string, osmDir string, dbDir string) {
	var optionalNewId int = 1
	for _, line := range refs {
		fmt.Printf("Mapping into %s \n", line)
		var anchors map[string]([]*osmUtils.Node) = map[string]([]*osmUtils.Node){}

		var osm osmUtils.Osm
		var dbIss XmlIssDaten
		osmFile, err := os.ReadFile(osmDir + "/" + line + ".xml")
		if err != nil {
			panic(err)
		}
		dbFile, err := os.ReadFile(dbDir + "/" + line + "_DB.xml")
		if err != nil {
			panic(err)
		}
		err = xml.Unmarshal([]byte(osmFile), &osm)
		if err != nil {
			panic(err)
		}
		err = xml.Unmarshal([]byte(dbFile), &dbIss)
		if err != nil {
			panic(err)
		}

		var notFoundSignalsFalling []*Signal = []*Signal{}
		var notFoundSignalsRising []*Signal = []*Signal{}

		numberNotFoundSignals := 0

		MapMainSignalsWithAnchorSearch(
			&osm,
			dbIss,
			&anchors,
			&notFoundSignalsFalling,
			&notFoundSignalsRising,
			&optionalNewId,
			&numberNotFoundSignals,
		)
		fmt.Printf("Inserted %d anchors \n", len(anchors))
		var issUnmappedMainSignals = XmlIssDaten{
			Betriebsstellen: []*Spurplanbetriebsstelle{{
				Abschnitte: []*Spurplanabschnitt{{
					Knoten: []*Spurplanknoten{{
						HauptsigF: notFoundSignalsFalling,
						HauptsigS: notFoundSignalsRising,
					}},
				}},
			}},
		}
		fmt.Printf("Got %d conflicts \n", numberNotFoundSignals)
		numberNotFoundSignals = 0
		notFoundSignalsFalling = []*Signal{}
		notFoundSignalsRising = []*Signal{}
		MapMainSignalsExistingAnchors(
			&osm,
			issUnmappedMainSignals,
			&anchors,
			&notFoundSignalsFalling,
			&notFoundSignalsRising,
			&optionalNewId,
			&numberNotFoundSignals,
		)
		fmt.Printf("Could not find: %d \n", numberNotFoundSignals)

		/*
			mapPoints(&osmData, dbData, &mappedItems)
			mapRest(&osmData, dbData, &mappedItems)
		*/

		updatedOsm, err := xml.MarshalIndent(osm, "", "	")
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header+string(updatedOsm)), 0644)
		if err != nil {
			panic(err)
		}
	}

}
