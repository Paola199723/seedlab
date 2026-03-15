package schema

import "seedlab/internal/domain"

func FromDomainSchema(db domain.DatabaseSchema) Snapshot {

	// Inicializar slices como vacíos
	snap := Snapshot{
		Tables:      make([]Table, 0),
		ForeignKeys: make([]ForeignKey, 0),
	}

	for _, t := range db.Tables {

		cols := make([]Column, 0) // inicializar slice de columnas vacío

		for _, c := range t.Columns {

			cols = append(cols, Column{
				Name:         c.Name,
				Type:         c.Type,
				IsNullable:   c.IsNullable,
				IsUnique:     c.IsUnique,
				DefaultValue: c.DefaultValue,
			})

		}

		snap.Tables = append(snap.Tables, Table{
			Name:       t.Name,
			Columns:    cols,
			PrimaryKey: t.PrimaryKey,
		})
	}

	for _, fk := range db.ForeignKeys {

		snap.ForeignKeys = append(snap.ForeignKeys, ForeignKey{
			Table:            fk.Table,
			Column:           fk.Column,
			ReferencedTable:  fk.ReferencedTable,
			ReferencedColumn: fk.ReferencedColumn,
		})

	}

	return snap
}
