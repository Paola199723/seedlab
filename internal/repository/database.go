package repository

import (
	"context"
	"seedlab/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseRepository interface {
	GetTables(ctx context.Context) ([]domain.Table, error)
	GetForeignKeys(ctx context.Context) ([]domain.ForeignKey, error)
}

type databaseRepository struct {
	db *pgxpool.Pool
}

func NewDatabaseRepository(db *pgxpool.Pool) DatabaseRepository {
	return &databaseRepository{db: db}
}

func (r *databaseRepository) GetTables(ctx context.Context) ([]domain.Table, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []domain.Table
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		// Get columns for each table
		columns, err := r.getColumns(ctx, name)
		if err != nil {
			return nil, err
		}
		tables = append(tables, domain.Table{Name: name, Columns: columns})
	}
	return tables, nil
}

func (r *databaseRepository) getColumns(ctx context.Context, tableName string) ([]domain.Column, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default,
		CASE WHEN EXISTS (
			SELECT 1 FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			WHERE tc.table_name = c.table_name AND tc.table_schema = c.table_schema AND tc.constraint_type = 'UNIQUE' AND kcu.column_name = c.column_name
		) THEN true ELSE false END as is_unique
		FROM information_schema.columns c
		WHERE table_name = $1 AND table_schema = 'public'
		ORDER BY ordinal_position
	`
	rows, err := r.db.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []domain.Column
	for rows.Next() {
		var col domain.Column
		var nullable string
		var defaultVal *string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &col.IsUnique); err != nil {
			return nil, err
		}
		col.IsNullable = nullable == "YES"
		col.DefaultValue = defaultVal
		columns = append(columns, col)
	}
	return columns, nil
}

func (r *databaseRepository) GetForeignKeys(ctx context.Context) ([]domain.ForeignKey, error) {
	query := `
		SELECT
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS referenced_table,
			ccu.column_name AS referenced_column
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []domain.ForeignKey
	for rows.Next() {
		var fk domain.ForeignKey
		if err := rows.Scan(&fk.Table, &fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, nil
}