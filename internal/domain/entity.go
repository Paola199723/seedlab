package domain

type Table struct {
	Name       string
	Columns    []Column
	PrimaryKey []string
}

type Column struct {
	Name         string
	Type         string
	IsNullable   bool
	IsUnique     bool
	DefaultValue *string
}

type ForeignKey struct {
	Table        string
	Column       string
	ReferencedTable string
	ReferencedColumn string
}

type DatabaseSchema struct {
	Tables      []Table
	ForeignKeys []ForeignKey
}