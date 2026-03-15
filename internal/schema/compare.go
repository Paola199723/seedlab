package schema

import (
	"encoding/json"
	"sort"
)

func normalizeSnapshot(snap Snapshot) Snapshot {
	// Ordenar tablas por nombre
	sort.Slice(snap.Tables, func(i, j int) bool {
		return snap.Tables[i].Name < snap.Tables[j].Name
	})
	for i := range snap.Tables {
		// Ordenar columnas por nombre
		sort.Slice(snap.Tables[i].Columns, func(a, b int) bool {
			return snap.Tables[i].Columns[a].Name < snap.Tables[i].Columns[b].Name
		})
	}
	// Ordenar foreign keys
	sort.Slice(snap.ForeignKeys, func(i, j int) bool {
		if snap.ForeignKeys[i].Table != snap.ForeignKeys[j].Table {
			return snap.ForeignKeys[i].Table < snap.ForeignKeys[j].Table
		}
		return snap.ForeignKeys[i].Column < snap.ForeignKeys[j].Column
	})

	// Omitir la versión al comparar
	snap.Version = 0

	return snap
}

func HasSchemaChanged(old *Snapshot, new Snapshot) bool {
	if old == nil {
		return true
	}

	oldNorm := normalizeSnapshot(*old)
	newNorm := normalizeSnapshot(new)

	oldJSON, _ := json.Marshal(oldNorm)
	newJSON, _ := json.Marshal(newNorm)

	return string(oldJSON) != string(newJSON)
}
