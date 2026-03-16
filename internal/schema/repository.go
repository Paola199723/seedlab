package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SaveNewSnapshot(folder string, snap Snapshot) (int, error) {

	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return 0, err
	}

	files, _ := os.ReadDir(folder)

	version := len(files) + 1

	snap.Version = version

	name := fmt.Sprintf("%04d_schema.json", version)

	path := filepath.Join(folder, name)

	data, _ := json.MarshalIndent(snap, "", "  ")

	err = os.WriteFile(path, data, 0644)

	return version, err
}

func IsSchemaDifferent(a, b Snapshot) bool {

	aJson, _ := json.Marshal(a)
	bJson, _ := json.Marshal(b)

	return string(aJson) != string(bJson)
}
