package dbUtils

import (
	"encoding/xml"
	"fmt"
	"os"

	OSMUtil "transform-osm/osm-utils"

	"github.com/pkg/errors"
)

// MapDB maps all elements present in DB-data line files in 'DBDir' using the respective OSM line file present in 'osmDir'.
// First, all anchor-able signals and switches are mapped, second all other non-anchor-able elements.
func MapDB(
	refs []string,
	osmDir string,
	DBDir string,
) error {
	newNodeIdCounter := 0
	linesWithNoAnchors := 0
	for _, line := range refs {
		var anchors map[float64]([]*OSMUtil.Node) = map[float64]([]*OSMUtil.Node){}
		var osm OSMUtil.Osm
		var dbIss XmlIssDaten

		osmLineFilePath := osmDir + "/" + line + ".xml"
		osmFile, err := os.ReadFile(osmLineFilePath)
		if err != nil {
			return errors.Wrap(err, "could not read osm line file: "+osmLineFilePath)
		}
		dbLineFilePath := DBDir + "/" + line + "_DB.xml"
		dbFile, err := os.ReadFile(dbLineFilePath)
		if err != nil {
			return errors.Wrap(err, "could not read DB line file: "+dbLineFilePath)
		}

		if err := xml.Unmarshal([]byte(osmFile), &osm); err != nil {
			return errors.Wrap(err, "could not unmarshal osm file: "+osmLineFilePath)
		}
		if err := xml.Unmarshal([]byte(dbFile), &dbIss); err != nil {
			return errors.Wrap(err, "could not unmarshal db file: "+dbLineFilePath)
		}

		fmt.Printf("Processing line %s \n", line)

		var notFoundSignalsFalling []*Signal = []*Signal{}
		var notFoundSignalsRising []*Signal = []*Signal{}
		var foundAnchorCount = 0
		for _, stelle := range dbIss.Betriebsstellen {
			for _, abschnitt := range stelle.Abschnitte {
				err = findAndMapAnchorMainSignals(
					abschnitt,
					&osm,
					anchors,
					&notFoundSignalsFalling,
					&notFoundSignalsRising,
					&foundAnchorCount,
					&newNodeIdCounter,
				)
				if err != nil {
					return errors.Wrap(err, "could not anchor main signals")
				}
				err = findAndMapAnchorSwitches(
					abschnitt,
					&osm,
					anchors,
					&foundAnchorCount,
					&newNodeIdCounter,
				)
				if err != nil {
					return errors.Wrap(err, "could not anchor switches")
				}
			}
		}

		numSignalsNotFound := (float64)(len(notFoundSignalsFalling) + len(notFoundSignalsRising))
		percentAnchored := ((float64)(foundAnchorCount) / ((float64)(foundAnchorCount) + numSignalsNotFound)) * 100.0
		fmt.Printf("Could anchor %f %% of signals. \n", percentAnchored)

		var issWithMappedSignals = XmlIssDaten{
			Betriebsstellen: []*Spurplanbetriebsstelle{{
				Abschnitte: []*Spurplanabschnitt{{
					Knoten: []*Spurplanknoten{{
						HauptsigF: notFoundSignalsFalling,
						HauptsigS: notFoundSignalsRising,
					}},
				}},
			}},
		}

		for _, stelle := range issWithMappedSignals.Betriebsstellen {
			for _, abschnitt := range stelle.Abschnitte {
				mapUnanchoredMainSignals(&osm, &anchors,
					&newNodeIdCounter, *abschnitt)
			}
		}

		if new_Data, err := xml.MarshalIndent(osm, "", "	"); err != nil {
			return errors.Wrap(err, "could not marshal osm data")
		} else {
			if err := os.WriteFile(osmLineFilePath,
				[]byte(xml.Header+string(new_Data)), 0644); err != nil {
				return errors.Wrap(err, "could not write file: "+osmLineFilePath)
			}
		}
	}

	fmt.Printf("Lines with no anchors: %d out of %d \n", linesWithNoAnchors, len(refs))
	return nil
}
