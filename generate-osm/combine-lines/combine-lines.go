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

	var r, g, b int64

	var h, s, v float64

	//s=1 and v=1 leads to bright and saturated colors
	h = (float64)(rand.Intn((360)))
	s = 1
	v = 1

	//convert HSV to RGB
	h_i := math.Floor(h / 60)

	f := (h / 60) - h_i

	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch {
	case h_i == 0 || h_i == 6:
		r = int64(math.Floor(v * 255))
		g = int64(math.Floor(t * 255))
		b = int64(math.Floor(p * 255))
	case h_i == 1:
		r = int64(math.Floor(q * 255))
		g = int64(math.Floor(v * 255))
		b = int64(math.Floor(p * 255))
	case h_i == 2:
		r = int64(math.Floor(p * 255))
		g = int64(math.Floor(v * 255))
		b = int64(math.Floor(t * 255))
	case h_i == 3:
		r = int64(math.Floor(p * 255))
		g = int64(math.Floor(q * 255))
		b = int64(math.Floor(v * 255))
	case h_i == 4:
		r = int64(math.Floor(t * 255))
		g = int64(math.Floor(p * 255))
		b = int64(math.Floor(v * 255))
	case h_i == 5:
		r = int64(math.Floor(v * 255))
		g = int64(math.Floor(p * 255))
		b = int64(math.Floor(q * 255))

	}

	color := "#"

	//translate rgb values to hex
	color += leftPad(strconv.FormatInt((r), 16), 2, "0") + leftPad(strconv.FormatInt((g), 16), 2, "0") + leftPad(strconv.FormatInt((b), 16), 2, "0")

	return color
}

func leftPad(stringToPad string, length int, padding string) string {
	for len(stringToPad) < length {
		stringToPad = padding + stringToPad
	}
	return stringToPad
}
