package dbUtils

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	OSMUtil "transform-osm/osm-utils"
)

var XML_TAG_NAME_CONST = xml.Name{Space: " ", Local: "tag"}
var numItemsNotFound int
var numItemsFound int

func MapDB(
	nodeIdCounter *int,
	refs []string, osmDir string, DBDir string,
) {
	for _, line := range refs {
		var anchors = make(map[string]([]*OSMUtil.Node))

		numItemsFound = 0
		numItemsNotFound = 0

		var dbData XmlIssDaten
		var osmData OSMUtil.Osm
		osmData = OSMUtil.Osm{}

		osm_file, err := os.ReadFile(osmDir + "/" + line + ".xml")
		if err != nil {
			log.Fatal(err)
		}
		db_file, err := os.ReadFile(DBDir + "/" + line + "_DB.xml")
		if err != nil {
			log.Fatal(err)
		}

		if err := xml.Unmarshal([]byte(osm_file), &osmData); err != nil {
			panic(err)
		}
		if err := xml.Unmarshal([]byte(db_file), &dbData); err != nil {
			panic(err)
		}

		fmt.Printf("Processing line %s \n", line)

		mainF, mainS := findAndMapAnchorMainSignals(&osmData, &anchors,
			nodeIdCounter, dbData)

		// anchorPoints(&osmData, dbData)
		fmt.Printf("Found %d anchors and could not find %d \n", numItemsFound, numItemsNotFound)

		var restData = XmlIssDaten{
			Betriebsstellen: []*Spurplanbetriebsstelle{{
				Abschnitte: []*Spurplanabschnitt{{
					Knoten: []*Spurplanknoten{{
						HauptsigF: mainF,
						HauptsigS: mainS,
					}},
				}},
			}},
		}
		_ = restData

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml",
				[]byte(xml.Header+string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
}

func formatKilometrage(anchors *map[string]([]*OSMUtil.Node),
	in float64,
) (out string) {
	out = strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len((*anchors)[out]) == 0; out += "0" {
		if !strings.Contains(out, ",") {
			out += ","
		}
	}
	return
}
