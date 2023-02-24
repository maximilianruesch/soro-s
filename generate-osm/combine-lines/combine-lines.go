package combineLines

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	osmUtils "transform-osm/osm-utils"
)

var ErrLinesDirNotFound = errors.New("lines directory not found")

func CombineAllLines(tempLineDir string) (osmUtils.Osm, error) {
	files, err := os.ReadDir(tempLineDir)

	if err != nil {
		return osmUtils.Osm{}, ErrLinesDirNotFound
	}

	var osmData osmUtils.Osm

	for _, file := range files {
		fmt.Printf("Processing %s... ", file.Name())
		data, _ := os.ReadFile(tempLineDir + "/" + file.Name())

		var fileOsmData osmUtils.Osm
		if err := xml.Unmarshal([]byte(data), &fileOsmData); err != nil {
			panic(err)
		}
		// Gernerate random colours for the lines
		var color = getRandomColor()
		for i := range fileOsmData.Way {
			fileOsmData.Way[i].Tag = append(fileOsmData.Way[i].Tag, &osmUtils.Tag{
				K: "color",
				V: color,
			})
		}

		osmData.Node = append(osmData.Node, fileOsmData.Node...)
		osmData.Way = append(osmData.Way, fileOsmData.Way...)
		osmData.Relation = append(osmData.Relation, fileOsmData.Relation...)
		fmt.Println("Done")
	}
	fmt.Println("Done processing files")

	return osmData, nil
}

func getRandomColor() string {

	//This fixed the problem, but not in the best way possible
	//HSV may be the way to go

	var r, g, b int64

	//Only allow colors within a certain range
	//this prevents white and black
	r = (int64)(rand.Intn((100)) + 50)
	g = (int64)(rand.Intn((100)) + 50)
	b = (int64)(rand.Intn((100)) + 50)

	color := "#"

	//Grey colors are generated if the RGB values are too close to each other
	//If a Grey color is detected, a new random color is generated
	for (math.Abs((float64)(r-g)) < 20) || (math.Abs((float64)(r-b)) < 20) || (math.Abs((float64)(b-r)) < 20) || (math.Abs((float64)(b-g)) < 20) || (math.Abs((float64)(g-b)) < 20) || (math.Abs((float64)(g-r)) < 20) {
		r = (int64)(rand.Intn(100) + 50)
		g = (int64)(rand.Intn(100) + 50)
		b = (int64)(rand.Intn(100) + 50)
	}

	//translate to hex code
	color += strconv.FormatInt((r), 16) + strconv.FormatInt((g), 16) + strconv.FormatInt((b), 16)

	return color
}
