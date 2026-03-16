package schema

import (
	"context"
	"seedlab/internal/domain"
	"seedlab/internal/repository"
)

type SchemaService struct {
	repo repository.DatabaseRepository
}

func NewSchemaService(repo repository.DatabaseRepository) *SchemaService {
	return &SchemaService{
		repo: repo,
	}
}

func (s *SchemaService) LoadDatabaseSchema(ctx context.Context) (domain.DatabaseSchema, error) {

	tables, err := s.repo.GetTables(ctx)
	if err != nil {
		return domain.DatabaseSchema{}, err
	}

	fks, err := s.repo.GetForeignKeys(ctx)
	if err != nil {
		return domain.DatabaseSchema{}, err
	}

	return domain.DatabaseSchema{
		Tables:      tables,
		ForeignKeys: fks,
	}, nil
}
