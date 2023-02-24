package dbUtils

import (
	"encoding/xml"
	"fmt"
	"os"
	OSMUtil "transform-osm/osm-utils"
)

func MapDB(refs []string, osmDir string, dbDir string) {
	var optionalNewId int = 1
	for _, line := range refs {
		var anchors map[string]([]*OSMUtil.Node) = map[string]([]*OSMUtil.Node){}

		var osm OSMUtil.Osm
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

		var notFoundedSignalsFalling []*Signal = []*Signal{}
		var notFoundedSignalsRising []*Signal = []*Signal{}

		MapSignalsWithAnchorSearch(
			dbIss,
			&osm,
			anchors,
			&notFoundedSignalsFalling,
			&notFoundedSignalsRising,
			&optionalNewId,
		)
		var issWithMappedSignals = XmlIssDaten{
			Betriebsstellen: []*Spurplanbetriebsstelle{{
				Abschnitte: []*Spurplanabschnitt{{
					Knoten: []*Spurplanknoten{{
						HauptsigF: notFoundedSignalsFalling,
						HauptsigS: notFoundedSignalsRising,
					}},
				}},
			}},
		}
		numberFoundSignals := MapSignalsExistingAnchors(
			issWithMappedSignals,
			&osm,
			anchors,
			&notFoundedSignalsFalling,
			&notFoundedSignalsRising,
			&optionalNewId,
		)

		/*
			mapPoints(&osmData, dbData, &mappedItems)
			mapRest(&osmData, dbData, &mappedItems)
		*/
		if updatedOsm, err := xml.MarshalIndent(osm, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header+string(updatedOsm)), 0644); err != nil {
				panic(err)
			}
		}
		fmt.Printf("Could not find: %d \n", numberFoundSignals)
	}

}
