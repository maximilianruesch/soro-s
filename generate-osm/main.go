package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	combineLines "transform-osm/combine-lines"
	osmUtils "transform-osm/osm-utils"
	stationsHaltsDisplay "transform-osm/stations-halts-display"

	"github.com/urfave/cli/v2"
)

func main() {
	os.Mkdir("./temp", 0755)
	var generateLines bool
	var inputFile string

	app := &cli.App{
		Name:  "generate-osm",
		Usage: "Generate OSM file from OSM PBF file and DB Data",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "generate-lines",
				Aliases:     []string{"gl"},
				Usage:       "Generate lines all lines new",
				Destination: &generateLines,
			},
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Value:       "./temp/base.osm.pbf",
				Usage:       "The input file to read as OSM PBF file",
				Destination: &inputFile,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if err := generateOsm(generateLines, inputFile); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func generateOsm(generateLines bool, inputFile string) error {
	if !filepath.IsAbs(inputFile) {
		inputFile, _ = filepath.Abs(inputFile)
	}
	if _, err := os.Stat(inputFile); err != nil {
		return errors.New("Input file does not exist: " + inputFile)
	}
	if filepath.Ext(inputFile) != ".pbf" {
		return errors.New("Input file is not a PBF file: " + inputFile)
	}

	tracksWithOnlyRelationsFile, _ := filepath.Abs("./temp/tracksWithOnlyRelations.osm.pbf")
	tracksFile, _ := filepath.Abs("./temp/tracks.osm.pbf")
	refOutputFile, _ := filepath.Abs("./temp/trackRefs.xml")

	osmUtils.ExecuteOsmFilterCommand([]string{
		"-R",
		inputFile,
		"-o",
		tracksWithOnlyRelationsFile,
		"r/route=tracks",
		"--overwrite",
	})
	osmUtils.ExecuteOsmFilterCommand([]string{
		inputFile,
		"-o",
		tracksFile,
		"r/route=tracks",
		"--overwrite",
	})
	osmUtils.ExecuteOsmFilterCommand([]string{
		tracksWithOnlyRelationsFile,
		"-o",
		refOutputFile,
		"r/ref",
		"--overwrite",
	})

	refs, err := getRefIds(refOutputFile)
	if err != nil {
		return errors.New("Failed to get ref ids: " + err.Error())
	}

	if generateLines {
		if err = os.RemoveAll("./temp/lines"); err != nil {
			return errors.New("Failed to remove lines folder: " + err.Error())
		}
		if err = os.Mkdir("./temp/lines", 0755); err != nil {
			return errors.New("Failed to create lines folder: " + err.Error())
		}

		for _, refId := range refs {
			lineOsmFile, err := filepath.Abs("./temp/lines/" + refId + ".xml")
			if err != nil {
				return errors.New("Failed to get line file path: " + err.Error())
			}
			osmUtils.ExecuteOsmFilterCommand([]string{
				tracksFile,
				"-o",
				lineOsmFile,
				"ref=" + refId,
				"--overwrite",
			})
		}

		fmt.Println("Generated all lines")
	}

	// Combine all the lines into one file
	osmData, err := combineLines.CombineAllLines()
	if err != nil && errors.Is(err, combineLines.ErrLinesDirNotFound) {
		return errors.New("You need to generate lines first")
	} else if err != nil {
		return errors.New("Failed to combine lines: " + err.Error())
	}
	osmData.Version = "0.6"
	osmData.Generator = "osmium/1.14.0"

	// Create stations file
	stattionsUnfilteredFile, _ := filepath.Abs("./temp/stationsUnfiltered.osm.pbf")
	stationsFile, _ := filepath.Abs("./temp/stations.xml")
	osmUtils.ExecuteOsmFilterCommand([]string{
		inputFile,
		"-o",
		stattionsUnfilteredFile,
		"n/railway=station,halt,facility",
		"--overwrite",
	})
	osmUtils.ExecuteOsmFilterCommand([]string{
		stattionsUnfilteredFile,
		"-o",
		stationsFile,
		"-i",
		"n/subway=yes",
		"n/monorail=yes",
		"n/usage",
		"n/tram=yes",
		"--overwrite",
	})

	jsonData := stationsHaltsDisplay.StationsHaltsDisplay(stationsFile)
	// save stations as json
	output, err := json.MarshalIndent(jsonData, "", "     ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.WriteFile("./temp/stations.json", output, 0644)
	for _, node := range osmData.Node {
		for id := range jsonData["stations"] {
			if node.Id == id {
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "type", V: "station"})
			}
		}

		for id := range jsonData["halts"] {
			if node.Id == id {
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "type", V: "element"})
				node.Tag = append(node.Tag, &osmUtils.Tag{K: "subtype", V: "hlt"})
			}
		}
	}

	sortedOsmData := osmUtils.SortOsm(osmData)
	output, err = xml.MarshalIndent(sortedOsmData, "", "     ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	output = []byte(xml.Header + string(output))
	os.WriteFile("./temp/finalOsm.xml", output, 0644)

	return nil
}

func getRefIds(trackRefFile string) (refs []string, err error) {
	var data []byte
	if data, err = os.ReadFile(trackRefFile); err != nil {
		return nil, errors.New("Failed to read track ref file: " + err.Error())
	}
	var osmData osmUtils.Osm
	if err := xml.Unmarshal([]byte(data), &osmData); err != nil {
		return nil, err
	}
	for _, s := range osmData.Relation {
		for _, m := range s.Tag {
			if m.K == "ref" {
				refs = append(refs, m.V)
			}
		}
	}

	return refs, nil
}
