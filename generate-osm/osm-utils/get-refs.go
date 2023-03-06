package osmUtils

import (
	"encoding/xml"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func GenerateOsmTrackRefs(inputFilePath string, tempFilePath string) (refs []string, err error) {
	refsPath, _ := filepath.Abs(tempFilePath + "/refs.xml")

	err = ExecuteOsmFilterCommand([]string{
		"-R",
		inputFilePath,
		"-o",
		refsPath,
		"r/route=tracks,railway",
		"--overwrite",
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute osmium command")
	}

	var data []byte
	if data, err = os.ReadFile(refsPath); err != nil {
		return nil, errors.Wrap(err, "Failed to read refs file")
	}
	var osmData Osm
	if err := xml.Unmarshal([]byte(data), &osmData); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal refs file")
	}

	return getRefIds(osmData), nil
}

func getRefIds(trackRefOsm Osm) []string {
	var refs []string
	for _, s := range trackRefOsm.Relation {
		for _, m := range s.Tag {
			if m.K == "ref" &&
				len(m.V) == 4 {
				refs = append(refs, m.V)
			}
		}
	}

	return refs
}
