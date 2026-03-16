package main

import (
	"context"
	"log"
	"os"
	"seedlab/internal/adapter/cli"
	"seedlab/internal/repository"
	"seedlab/internal/schema"
	"seedlab/internal/usecase"

	"seedlab/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Cargar .env
	cfg := config.Load()

	if err := cfg; err != nil {
		if os.Getenv("DEBUG") == "true" {
        	log.Println("No se pudo cargar .env")
   		 }
	}

	args := os.Args

	if len(args) > 1 {
		err := config.RunCLICommand(args, cfg)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// modo interactivo
	
	// Conectar a BD
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error conectando a BD:", err)
	}
	defer pool.Close()

	// Inicializar capas
	ctx := context.Background()
	repo := repository.NewDatabaseRepository(pool)
	schemaService := schema.NewSchemaService(repo)

	dbSchema, err := schemaService.LoadDatabaseSchema(ctx)
	if err != nil {
		panic(err)
	}

	snapshot := schema.FromDomainSchema(dbSchema)

	last, _ := schema.LoadLastSnapshot(cfg.NameArchive)

	var version int

	//--------------------------------
	// primera ejecución
	//--------------------------------

	if last == nil {

		version = 1

		snapshot.Version = version
		cfg.Version = snapshot.Version
		err := schema.SaveSnapshot(cfg.NameArchive, snapshot)
		if err != nil {
			log.Println(err)
		}

	} else if schema.HasSchemaChanged(last, snapshot) {

		version = last.Version + 1

		snapshot.Version = version
		cfg.Version = snapshot.Version
		err := schema.SaveSnapshot(cfg.NameArchive, snapshot)
		if err != nil {
			log.Println(err)
		}
	}

	uc := usecase.NewTableUseCase(repo)
	cliAdapter := cli.NewCLIAdapter(uc, cfg)

	// Correr aplicación
	if err := cliAdapter.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
	

}
