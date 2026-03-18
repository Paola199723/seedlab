package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	comand := args[1]
	ver, err := GetVersionFromFilename(NameArchive)
	fileName := fmt.Sprintf("%04d_%s", ver,NameArchive)
	switch comand {

	case "png":
		err = generator.GeneratePNG(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".png")
	case "draw":
		err = generator.GenerateDraw(selectedTables.Tables, selectedTables.ForeignKeys,NameArchive+".drawio")
	case "excel":
		fs := flag.NewFlagSet("excel", flag.ExitOnError)
		rows := fs.Int("rows", 0, "number of fake rows")

    	fs.Parse(args[2:])
		err = generator.GenerateExcel(selectedTables.Tables, NameArchive+".xlsx", *rows, false)
	case "all":
		err = generator.GeneratePNG(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".png")
		if err == nil {
			err = generator.GenerateDraw(selectedTables.Tables, selectedTables.ForeignKeys, NameArchive+".drawio")
		}
		if err != nil {
			fmt.Println("Unknown command")
		}
	case "document":	
		err = generator.GenerateWord(selectedTables.Tables,selectedTables.ForeignKeys, fileName+".docx")
	case "document-md":
		err = generator.GenerateMarkdown(selectedTables.Tables, fileName+".md")
	default:
		fmt.Println("Unknown command")
	}

	return err
}

// GetVersionFromFilename extrae la versión de un archivo con formato "0001_name.json"
func GetVersionFromFilename(filePath string) (int, error) {
	base := filepath.Base(filePath) // solo el nombre, sin ruta
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("archivo no tiene formato esperado: %s", base)
	}
	versionStr := parts[0]              // "0001"
	version, err := strconv.Atoi(versionStr) // 1
	if err != nil {
		return 0, fmt.Errorf("error convirtiendo versión: %v", err)
	}
	return version, nil
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
