package usecase

import (
	"context"
	"seedlab/internal/domain"
	"seedlab/internal/repository"
	"strings"
)

type TableUseCase interface {
	GetOrderedTables(ctx context.Context) ([]domain.Table, []domain.ForeignKey, error)
	GetRelatedTables(tableName string, tables []domain.Table, fks []domain.ForeignKey) []domain.Table
	SearchTables(query string, tables []domain.Table) []domain.Table
}

type tableUseCase struct {
	repo repository.DatabaseRepository
}

func NewTableUseCase(repo repository.DatabaseRepository) TableUseCase {
	return &tableUseCase{repo: repo}
}

func (uc *tableUseCase) GetOrderedTables(ctx context.Context) ([]domain.Table, []domain.ForeignKey, error) {
	tables, err := uc.repo.GetTables(ctx)
	if err != nil {
		return nil, nil, err
	}
	fks, err := uc.repo.GetForeignKeys(ctx)
	if err != nil {
		return nil, nil, err
	}
	// Ordenar por dependencias usando topological sort
	orderedTables := uc.topologicalSort(tables, fks)
	return orderedTables, fks, nil
}

func (uc *tableUseCase) GetRelatedTables(tableName string, tables []domain.Table, fks []domain.ForeignKey) []domain.Table {
	// Implementar lógica para obtener tablas relacionadas recursivamente
	// Usar un set para evitar duplicados
	selected := make(map[string]bool)
	uc.selectRelated(tableName, tables, fks, selected)
	var result []domain.Table
	for _, t := range tables {
		if selected[t.Name] {
			result = append(result, t)
		}
	}
	return result
}

func (uc *tableUseCase) selectRelated(tableName string, tables []domain.Table, fks []domain.ForeignKey, selected map[string]bool) {
	if selected[tableName] {
		return
	}
	selected[tableName] = true
	// Tablas que referencian a esta (padres)
	for _, fk := range fks {
		if fk.ReferencedTable == tableName {
			uc.selectRelated(fk.Table, tables, fks, selected)
		}
	}
	// Tablas referenciadas por esta (hijos)
	for _, fk := range fks {
		if fk.Table == tableName {
			uc.selectRelated(fk.ReferencedTable, tables, fks, selected)
		}
	}
}

func (uc *tableUseCase) SearchTables(query string, tables []domain.Table) []domain.Table {
	var result []domain.Table
	for _, t := range tables {
		if strings.Contains(strings.ToLower(t.Name), strings.ToLower(query)) {
			result = append(result, t)
		}
	}
	return result
}

func (uc *tableUseCase) topologicalSort(tables []domain.Table, fks []domain.ForeignKey) []domain.Table {
	// Implementar topological sort basado en FK
	// Crear grafo de dependencias
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	tableMap := make(map[string]domain.Table)
	for _, t := range tables {
		tableMap[t.Name] = t
		inDegree[t.Name] = 0
		graph[t.Name] = []string{}
	}
	for _, fk := range fks {
		graph[fk.ReferencedTable] = append(graph[fk.ReferencedTable], fk.Table)
		inDegree[fk.Table]++
	}
	// Queue para tablas sin dependencias
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	var result []domain.Table
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, tableMap[current])
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	// Si hay ciclos, agregar las restantes al final
	for name, t := range tableMap {
		found := false
		for _, r := range result {
			if r.Name == name {
				found = true
				break
			}
		}
		if !found {
			result = append(result, t)
		}
	}
	return result
}