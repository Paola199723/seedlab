package schema

type Snapshot struct {
	Version     int          `json:"version"`
	Tables      []Table      `json:"tables"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
}

type Table struct {
	Name       string   `json:"name"`
	Columns    []Column `json:"columns"`
	PrimaryKey []string `json:"primary_key,omitempty"`
}

type Column struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	IsNullable   bool    `json:"is_nullable"`
	IsUnique     bool    `json:"is_unique"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type ForeignKey struct {
	Table            string `json:"table"`
	Column           string `json:"column"`
	ReferencedTable  string `json:"referenced_table"`
	ReferencedColumn string `json:"referenced_column"`
}
