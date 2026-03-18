package ai

import (
	"encoding/json"
	"seedlab/internal/domain"
	"strconv"
)

func GenerateFakeRows(table domain.Table, rows int) (string, error) {

	prompt := "Generate fake realistic data.\n\n"

	prompt += "Table: " + table.Name + "\n"
	prompt += "Columns:\n"

	for _, col := range table.Columns {
		prompt += "- " + col.Name + " (" + col.Type + ")\n"
	}

	prompt += "\nGenerate " + strconv.Itoa(rows) + " rows.\n"
	prompt += "Return ONLY valid JSON array. No explanation."

	return Generate(prompt)
}

func ParseRows(jsonStr string) ([]map[string]interface{}, error) {

	var rows []map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}