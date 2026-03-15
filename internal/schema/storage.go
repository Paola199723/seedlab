package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func SaveSnapshot(project string, snap Snapshot) error {

	folder := "schema"

	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%04d_%s_schema.json", snap.Version, project)

	path := filepath.Join(folder, filename)

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
func LoadLastSnapshot(project string) (*Snapshot, error) {

	folder := "schema"

	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, nil
	}

	var projectFiles []string

	for _, f := range files {

		if filepath.Ext(f.Name()) == ".json" {
			projectFiles = append(projectFiles, f.Name())
		}

	}

	if len(projectFiles) == 0 {
		return nil, nil
	}

	sort.Strings(projectFiles)

	last := projectFiles[len(projectFiles)-1]

	path := filepath.Join(folder, last)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var snap Snapshot

	err = json.Unmarshal(data, &snap)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}
