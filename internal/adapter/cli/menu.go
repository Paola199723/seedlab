package cli

import (
	"context"
	"fmt"
	"seedlab/internal/domain"
	"seedlab/internal/usecase"
	"seedlab/pkg/generator"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CLIAdapter struct {
	app      *tview.Application
	useCase  usecase.TableUseCase
	tables   []domain.Table
	fks      []domain.ForeignKey
	selected map[string]bool
}

func NewCLIAdapter(useCase usecase.TableUseCase) *CLIAdapter {
	return &CLIAdapter{
		app:      tview.NewApplication(),
		useCase:  useCase,
		selected: make(map[string]bool),
	}
}

func (c *CLIAdapter) Run(ctx context.Context) error {
	tables, fks, err := c.useCase.GetOrderedTables(ctx)
	if err != nil {
		return err
	}
	c.tables = tables
	c.fks = fks

	return c.showMainMenu()
}

func (c *CLIAdapter) showMainMenu() error {
	menu := tview.NewList().
		AddItem("1. Generar PNG de las tablas", "", '1', func() {
			c.showTableSelection("png")
		}).
		AddItem("2. Generar .draw de las tablas", "", '2', func() {
			c.showTableSelection("draw")
		}).
		AddItem("3. Generar Excel de las tablas", "", '3', func() {
			c.showTableSelection("excel")
		}).
		AddItem("4. Generar SQL de INSERT + ROLLBACK desde Excel", "", '4', func() {
			c.showTableSelection("sql")
		}).
		AddItem("Salir", "", 'q', func() {
			c.app.Stop()
		})

	menu.SetBorder(true).SetTitle("Menú Principal - SeedLab")

	if err := c.app.SetRoot(menu, true).Run(); err != nil {
		return err
	}
	return nil
}

func (c *CLIAdapter) showTableSelection(action string) {
	if action == "sql" {
		// Para SQL, no se seleccionan tablas, se genera todo desde el Excel
		c.generate(action)
		return
	}
	list := tview.NewList()
	for i, table := range c.tables {
		name := table.Name
		if c.selected[name] {
			name = "[red]" + name + "[-]"
		}
		list.AddItem(name, "", rune(strconv.Itoa(i + 1)[0]), nil)
	}

	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {

		tableName := c.tables[index].Name
		if !c.selected[tableName] {
			// Seleccionar tabla y relacionadas
			related := c.useCase.GetRelatedTables(tableName, c.tables, c.fks)
			for _, t := range related {
				c.selected[t.Name] = true
			}
		} else {
			// Deseleccionar tabla y relacionadas
			related := c.useCase.GetRelatedTables(tableName, c.tables, c.fks)
			for _, t := range related {
				delete(c.selected, t.Name)
			}
		}
		c.showTableSelection(action) // Refresh
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '1' {
			// Generar
			c.generate(action)
		} else if event.Rune() == '2' {
			// Volver
			c.showMainMenu()
		}
		return event
	})

	list.SetBorder(true).SetTitle(fmt.Sprintf("Seleccionar Tablas para %s (Enter para generar, 2 para volver)", action))

	c.app.SetRoot(list, true)
}

func (c *CLIAdapter) generate(action string) {
	var selectedTables []domain.Table
	if action != "sql" {

		for _, t := range c.tables {
			if c.selected[t.Name] {
				selectedTables = append(selectedTables, t)
			}
		}
	}

	var err error
	switch action {
	case "png":
		err = generator.GeneratePNG(selectedTables, c.fks, "schema.png")
	case "draw":
		err = generator.GenerateDraw(selectedTables, c.fks, "schema.drawio")
	case "excel":
		err = generator.GenerateExcel(selectedTables, "schema.xlsx")
	case "sql":
		err := generator.GenerateInsertRollbackFromExcel("schema.xlsx")
		if err != nil {
			message2 := fmt.Sprintln("Error generando SQL:", err)
			modal := tview.NewModal().
				SetText(message2).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					c.showMainMenu()
				})

			c.app.SetRoot(modal, false)
		}
	}
	message := "Generación completada"
	if err != nil {
		message = fmt.Sprintf("Error: %v", err)
	}

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			c.showMainMenu()
		})

	c.app.SetRoot(modal, false)
}
