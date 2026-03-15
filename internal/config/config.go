package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
