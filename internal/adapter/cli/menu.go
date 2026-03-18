package cli

import (
	"context"
	"fmt"
	"strconv"

	"seedlab/internal/config"
	"seedlab/internal/domain"
	"seedlab/internal/usecase"
	"seedlab/pkg/generator"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CLIAdapter struct {
	app       *tview.Application
	useCase   usecase.TableUseCase
	tables    []domain.Table
	fks       []domain.ForeignKey
	selected  map[string]bool
	cfgConfig *config.Config
}

func NewCLIAdapter(useCase usecase.TableUseCase, cfg *config.Config) *CLIAdapter {
	return &CLIAdapter{
		app:       tview.NewApplication(),
		useCase:   useCase,
		selected:  make(map[string]bool),
		cfgConfig: cfg,
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

func (c *CLIAdapter) showExcelMenu() {

	menu := tview.NewList().
		AddItem("1. Generar Excel vacío (llenado manual)", "", '1', func() {
			c.showTableSelection("excel_empty")
		}).
		AddItem("2. Generar Excel con Fake Data", "", '2', func() {
			c.showTableSelection("excel_fake")
		}).
		AddItem("3. Volver", "", '3', func() {
			c.showMainMenu()
		})

	menu.SetBorder(true).SetTitle("Generación de Excel")

	c.app.SetRoot(menu, true)
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
			c.showExcelMenu()
		}).
		AddItem("4. Generar SQL de INSERT + ROLLBACK desde Excel", "", '4', func() {
			c.showTableSelection("sql")
		}).
		AddItem("5. Generar Documento + .md", "", '5', func() {
			c.showTableSelection("document")
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
		c.generate(action)
		return
	}

	list := tview.NewList()

	for _, table := range c.tables {

		name := table.Name

		if c.selected[name] {
			name = "[green]" + name + " ✓[-]"
		}

		list.AddItem(name, "", 0, nil)
	}

	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {

		// ENTER genera
		c.generate(action)

	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		index := list.GetCurrentItem()
		tableName := c.tables[index].Name

		switch {

		case event.Key() == tcell.KeyRune && event.Rune() == ' ':

			// seleccionar / deseleccionar tabla
			if !c.selected[tableName] {

				related := c.useCase.GetRelatedTables(tableName, c.tables, c.fks)

				for _, t := range related {
					c.selected[t.Name] = true
				}

			} else {

				related := c.useCase.GetRelatedTables(tableName, c.tables, c.fks)

				for _, t := range related {
					delete(c.selected, t.Name)
				}

			}

			c.showTableSelection(action)
			return nil

		case event.Key() == tcell.KeyEsc:

			c.showMainMenu()
			return nil
		}

		return event
	})

	list.SetBorder(true).
		SetTitle(fmt.Sprintf(
			"Seleccionar Tablas (%s)\nESPACIO seleccionar | ENTER generar | ESC volver",
			action,
		))

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

	fileName := fmt.Sprintf("%04d_%s", c.cfgConfig.Version, c.cfgConfig.NameArchive)

	var err error

	switch action {

	case "png":

		err = generator.GeneratePNG(selectedTables, c.fks, fileName+".png")

	case "draw":

		err = generator.GenerateDraw(selectedTables, c.fks, fileName+".drawio")

	case "excel_empty":

		err = generator.GenerateExcel(selectedTables, fileName+".xlsx", 0, false)

	case "excel_fake":

		c.askColumnsForFakeData(selectedTables, fileName)
		return

	case "document":

		err = generator.GenerateWord(selectedTables, c.fks, fileName+".docx")

	case "document-md":
		err = generator.GenerateMarkdown(selectedTables, fileName+".md")

	case "sql":

		err = generator.GenerateInsertRollbackFromExcel(c.cfgConfig.Version, fileName+".xlsx")
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

func (c *CLIAdapter) askColumnsForFakeData(selectedTables []domain.Table, fileName string) {

	input := tview.NewInputField().
		SetLabel("¿Cuántas columnas requiere?: ").
		SetFieldWidth(10)

	input.SetAcceptanceFunc(func(text string, lastChar rune) bool {

		return lastChar >= '0' && lastChar <= '9'
	})

	input.SetDoneFunc(func(key tcell.Key) {

		if key == tcell.KeyEnter {

			value := input.GetText()

			columns, err := strconv.Atoi(value)

			if err != nil {

				c.showError("Debe ingresar un número válido")
				return
			}
			err = generator.GenerateExcel(selectedTables, fileName+".xlsx", columns, true)
			message := "Excel generado correctamente"
			if err != nil {
				message = fmt.Sprintf("Error al generar Excel: %v", err)
			}
			modal := tview.NewModal().
				SetText(message).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					c.showMainMenu()
				})
			c.app.SetRoot(modal, false)
		}
	})

	input.SetBorder(true).
		SetTitle("Configuración Fake Data")

	c.app.SetRoot(input, true)
}

func (c *CLIAdapter) showError(msg string) {

	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			c.showMainMenu()
		})

	c.app.SetRoot(modal, false)
}