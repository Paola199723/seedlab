package main

import (
	"context"
	"log"
	"os"
	"seedlab/internal/adapter/cli"
	"seedlab/internal/repository"
	"seedlab/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar .env, usando variables de entorno")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL no configurada")
	}

	// Conectar a BD
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Error conectando a BD:", err)
	}
	defer pool.Close()

	// Inicializar capas
	repo := repository.NewDatabaseRepository(pool)
	uc := usecase.NewTableUseCase(repo)
	cliAdapter := cli.NewCLIAdapter(uc)

	// Correr aplicación
	if err := cliAdapter.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
