package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"seedlab/internal/domain"

	"baliance.com/gooxml/color"
	"baliance.com/gooxml/document"
	"baliance.com/gooxml/schema/soo/wml"
)

// GenerateMarkdown genera la documentación de las tablas en formato .md
func GenerateMarkdown(tables []domain.Table, filename string) error {
	// Crear carpeta docs si no existe
	docDir := "docs"
	if err := os.MkdirAll(docDir, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(docDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, t := range tables {
		fmt.Fprintf(f, "## Tabla: %s\n\n", t.Name)
		fmt.Fprintf(f, "| Nombre | Tipo | Nulo | Único | Default |\n")
		fmt.Fprintf(f, "|-------|------|------|-------|--------|\n")
		for _, c := range t.Columns {
			def := ""
			if c.DefaultValue != nil {
				def = *c.DefaultValue
			}
			fmt.Fprintf(f, "| %s | %s | %v | %v | %s |\n",
				c.Name, c.Type, c.IsNullable, c.IsUnique, def)
		}
		fmt.Fprintln(f, "\n")
	}

	return nil
}

// GenerateWord genera documentación profesional en Word
func GenerateWord(tables []domain.Table, fks []domain.ForeignKey, filename string) error {

	docDir := "docs"
	if err := os.MkdirAll(docDir, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(docDir, filename)
	doc := document.New()

	for _, t := range tables {

		//--------------------------------
		// TITULO
		//--------------------------------
		title := doc.AddParagraph()
		run := title.AddRun()
		run.Properties().SetBold(true)
		run.Properties().SetSize(28)
		run.AddText(fmt.Sprintf("Tabla: %s", t.Name))

		//--------------------------------
		// TABLA
		//--------------------------------
		tableDoc := doc.AddTable()

		// ✅ BORDES CORRECTOS
		borders := tableDoc.Properties().Borders()

		borders.SetTop(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 2)
		borders.SetBottom(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 2)
		borders.SetLeft(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 2)
		borders.SetRight(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 2)
		borders.SetInsideHorizontal(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 1)
		borders.SetInsideVertical(wml.ST_BorderSingle, color.FromHex("B8CCE4"), 1)
		//--------------------------------
		// HEADER
		//--------------------------------
		headers := []string{"Nombre", "Tipo", "Nulo", "Único", "Default"}
		headerRow := tableDoc.AddRow()

		for _, h := range headers {
			cell := headerRow.AddCell()

			cell.Properties().SetShading(
				wml.ST_ShdSolid,
				color.FromHex("DCE6F1"),
				color.Auto,
			)

			p := cell.AddParagraph()
			r := p.AddRun()
			r.Properties().SetBold(true)
			r.AddText(h)
		}

		//--------------------------------
		// COLUMNAS
		//--------------------------------
		for i, c := range t.Columns {

			row := tableDoc.AddRow()

			bg := color.Auto
			if i%2 == 0 {
				bg = color.FromHex("F2F6FB")
			}

			def := ""
			if c.DefaultValue != nil {
				def = *c.DefaultValue
			}

			values := []string{
				c.Name,
				c.Type,
				fmt.Sprintf("%v", c.IsNullable),
				fmt.Sprintf("%v", c.IsUnique),
				def,
			}

			for _, v := range values {
				cell := row.AddCell()

				cell.Properties().SetShading(
					wml.ST_ShdSolid,
					bg,
					color.Auto,
				)

				cell.AddParagraph().AddRun().AddText(v)
			}
		}

		doc.AddParagraph()

		//--------------------------------
		// RELACIONES
		//--------------------------------
		relTitle := doc.AddParagraph()
		relRun := relTitle.AddRun()
		relRun.Properties().SetBold(true)
		relRun.AddText("Relaciones:")

		hasRelations := false

		for _, fk := range fks {
			if fk.Table == t.Name {

				hasRelations = true

				p := doc.AddParagraph()

				// ⚠️ AJUSTA ESTO SEGÚN TU STRUCT REAL
				relatedTable := fk.ReferencedTable // <-- cambia aquí

				p.AddRun().AddText(fmt.Sprintf(
					"%s → %s (%s)",
					fk.Column,
					relatedTable,
					detectRelationType(fk, tables, fks),
				))
			}
		}

		if !hasRelations {
			doc.AddParagraph().AddRun().AddText("Sin relaciones")
		}

		doc.AddParagraph()
		doc.AddParagraph()
	}

	return doc.SaveToFile(filePath)
}

func detectRelationType(fk domain.ForeignKey, tables []domain.Table, fks []domain.ForeignKey) string {

	//--------------------------------
	// 1. Buscar tabla actual
	//--------------------------------
	var currentTable *domain.Table

	for _, t := range tables {
		if t.Name == fk.Table {
			currentTable = &t
			break
		}
	}

	if currentTable == nil {
		return "Desconocida"
	}

	//--------------------------------
	// 2. Verificar si la columna es UNIQUE
	//--------------------------------
	for _, col := range currentTable.Columns {
		if col.Name == fk.Column {
			if col.IsUnique {
				return "Uno a Uno (1:1)"
			}
		}
	}

	//--------------------------------
	// 3. Detectar tabla intermedia (Muchos a Muchos)
	//--------------------------------
	// contar cuántas FKs tiene esta tabla
	fkCount := 0
	for _, f := range fks {
		if f.Table == fk.Table {
			fkCount++
		}
	}

	// si tiene 2 o más FKs → posible tabla puente
	if fkCount >= 2 {

		// verificar si casi todas sus columnas son FKs
		totalCols := len(currentTable.Columns)

		if totalCols <= fkCount+1 {
			return "Muchos a Muchos (N:M)"
		}
	}

	//--------------------------------
	// 4. Default → Muchos a Uno
	//--------------------------------
	return "Muchos a Uno (N:1)"
}