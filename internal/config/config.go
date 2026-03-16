package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"seedlab/internal/domain"
	"seedlab/internal/schema"
	"seedlab/pkg/generator"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver    string
	DatabaseURL string
	NameArchive string
	Version     int
}

func Load() *Config {

	// intenta cargar .env
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró .env, usando variables del sistema")
	}

	name := os.Getenv("NAME_ARCHIVE")

	if name == "" {
		name = "seedlab_archive"
	}

	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")

	return &Config{
		DBDriver:    os.Getenv("DB_DRIVER"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		NameArchive: strings.ToLower(name),
		Version:     1,
	}
}
func getNextDrawVersion(folder string) (int, error) {

	files, err := os.ReadDir(folder)
	if err != nil {
		return 1, nil
	}

	maxVersion := 0

	for _, file := range files {

		name := file.Name()

		if filepath.Ext(name) != ".draw" {
			continue
		}

		var version int

		_, err := fmt.Sscanf(name, "%04d_", &version)
		if err != nil {
			continue
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}
func RunCLICommand(args []string) error {
	var err error
	last, NameArchive , _ := schema.LoadLatestSnapshot("schema")
	var selectedTables = MapSnapshotToDomain(*last)
	switch args[1] {

	case "png":
		err = generator.GeneratePNG(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".png")
	case "draw":
		err = generator.GenerateDraw(selectedTables.Tables, selectedTables.ForeignKeys,NameArchive+".drawio")
	//case "excel":
		//err = generator.GenerateExcel(selectedTables, fileName+".xlsx")
	case "all":
		err = generator.GeneratePNG(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".png")
		if err == nil {
			err = generator.GenerateDraw(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".drawio")
		}
		if err != nil {
			fmt.Println("Unknown command")
		}
	default:
		fmt.Println("Unknown command")
	}
	return err
}
func MapSnapshotToDomain(snapshot schema.Snapshot) domain.DatabaseSchema {

	var tables []domain.Table
	var fks []domain.ForeignKey

	for _, t := range snapshot.Tables {

		table := domain.Table{
			Name:       t.Name,
			PrimaryKey: t.PrimaryKey,
		}

		for _, c := range t.Columns {

			column := domain.Column{
				Name:         c.Name,
				Type:         c.Type,
				IsNullable:   c.IsNullable,
				IsUnique:     c.IsUnique,
				DefaultValue: c.DefaultValue,
			}

			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	for _, fk := range snapshot.ForeignKeys {

		domainFK := domain.ForeignKey{
			Table:            fk.Table,
			Column:           fk.Column,
			ReferencedTable:  fk.ReferencedTable,
			ReferencedColumn: fk.ReferencedColumn,
		}

		fks = append(fks, domainFK)
	}

	return domain.DatabaseSchema{
		Tables:      tables,
		ForeignKeys: fks,
	}
}
