package generator

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"seedlab/internal/domain"

	"github.com/fogleman/gg"
)

type Pos struct {
	X      float64
	Y      float64
	Height float64
	Width  float64
}

type Node struct {
	Name     string
	Children []string
	Level    int
}

func GeneratePNG(tables []domain.Table, fks []domain.ForeignKey, filename string) error {

	const (
		imgWidth     = 2200
		imgHeight    = 1600
		tableHeader  = 30
		columnHeight = 20
		tableMargin  = 120
		fontSize     = 12
	)

	dc := gg.NewContext(imgWidth, imgHeight)

	//--------------------------------
	// fondo blanco
	//--------------------------------

	dc.SetRGB(1, 1, 1)
	dc.Clear()

	//--------------------------------
	// texto
	//--------------------------------

	dc.SetRGB(0, 0, 0)

	//--------------------------------
	// mapa de tablas
	//--------------------------------

	tableMap := map[string]domain.Table{}

	for _, t := range tables {
		tableMap[t.Name] = t
	}

	//--------------------------------
	// filtrar FKs válidas
	//--------------------------------

	var validFKs []domain.ForeignKey

	for _, fk := range fks {
		if _, ok := tableMap[fk.Table]; ok {
			if _, ok2 := tableMap[fk.ReferencedTable]; ok2 {
				validFKs = append(validFKs, fk)
			}
		}
	}

	//--------------------------------
	// construir grafo
	//--------------------------------

	nodes := map[string]*Node{}

	for _, t := range tables {
		nodes[t.Name] = &Node{Name: t.Name}
	}

	for _, fk := range validFKs {

		parent := nodes[fk.ReferencedTable]
		child := nodes[fk.Table]

		if parent != nil && child != nil {
			parent.Children = append(parent.Children, child.Name)
		}
	}

	assignLevels(nodes, validFKs)

	//--------------------------------
	// agrupar por niveles
	//--------------------------------

	levels := map[int][]string{}

	for name, node := range nodes {
		levels[node.Level] = append(levels[node.Level], name)
	}

	positions := map[string]Pos{}
	levelHeights := map[int]float64{}

	//--------------------------------
	// calcular alturas
	//--------------------------------

	for level, tableNames := range levels {

		maxHeight := 0.0

		for _, name := range tableNames {

			table := tableMap[name]

			h := tableHeader + float64(len(table.Columns))*columnHeight

			if h > maxHeight {
				maxHeight = h
			}
		}

		levelHeights[level] = maxHeight
	}

	//--------------------------------
	// calcular anchos
	//--------------------------------

	levelWidths := map[int]float64{}

	for level, tableNames := range levels {

		maxWidth := 0.0

		for _, name := range tableNames {

			table := tableMap[name]

			w, _ := dc.MeasureString(table.Name)

			if w > maxWidth {
				maxWidth = w
			}

			for _, col := range table.Columns {

				txt := fmt.Sprintf("%s : %s", col.Name, col.Type)

				w, _ := dc.MeasureString(txt)

				if w > maxWidth {
					maxWidth = w
				}
			}
		}

		levelWidths[level] = maxWidth + 20
	}

	//--------------------------------
	// dibujar tablas
	//--------------------------------

	for level, tableNames := range levels {

		tableWidth := levelWidths[level]

		for i, name := range tableNames {

			table := tableMap[name]

			h := tableHeader + float64(len(table.Columns))*columnHeight

			x := float64(tableMargin) + float64(i)*(tableWidth+tableMargin)

			y := float64(tableMargin)

			for l := 0; l < level; l++ {
				y += levelHeights[l] + tableMargin
			}

			positions[name] = Pos{
				X:      x,
				Y:      y,
				Height: h,
				Width:  tableWidth,
			}

			drawTable(dc, table, x, y, tableWidth, h, tableHeader, columnHeight)
		}
	}

	//--------------------------------
	// dibujar relaciones
	//--------------------------------

	dc.SetRGB(0.8, 0, 0)
	dc.SetLineWidth(2)

	for _, fk := range validFKs {

		src := positions[fk.Table]
		dst := positions[fk.ReferencedTable]

		var srcX, dstX float64

		if src.X < dst.X {
			srcX = src.X + src.Width
			dstX = dst.X
		} else {
			srcX = src.X
			dstX = dst.X + dst.Width
		}

		srcY := src.Y + src.Height/2
		dstY := dst.Y + dst.Height/2

		cp1X := srcX + (dstX-srcX)*0.3
		cp1Y := srcY
		cp2X := dstX - (dstX-srcX)*0.3
		cp2Y := dstY

		dc.NewSubPath()
		dc.MoveTo(srcX, srcY)
		dc.CubicTo(cp1X, cp1Y, cp2X, cp2Y, dstX, dstY)
		dc.Stroke()

		angle := math.Atan2(dstY-srcY, dstX-srcX)

		drawArrow(dc, dstX, dstY, angle)

		midX := (srcX + dstX) / 2
		midY := (srcY + dstY) / 2

		dc.SetRGB(0, 0, 0)
		dc.DrawString(getCardinality(tables, validFKs, fk), midX-10, midY)
	}

	//--------------------------------
	// crear carpeta png
	//--------------------------------

	folder := "png"

	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}

	//--------------------------------
	// guardar archivo dentro de png
	//--------------------------------

	outputPath := filepath.Join(folder, filename)

	return dc.SavePNG(outputPath)
}

func drawTable(dc *gg.Context, table domain.Table, x, y, width, height float64, header float64, columnHeight float64) {

	dc.SetRGB(0.92, 0.92, 0.92)
	dc.DrawRectangle(x, y, width, height)
	dc.FillPreserve()

	dc.SetRGB(0, 0, 0)
	dc.Stroke()

	dc.DrawString(table.Name, x+10, y+20)

	colY := y + header

	for _, col := range table.Columns {

		txt := fmt.Sprintf("%s : %s", col.Name, col.Type)

		dc.DrawString(txt, x+10, colY+15)

		colY += columnHeight
	}
}

func drawArrow(dc *gg.Context, x, y, angle float64) {

	size := 8.0

	dc.DrawLine(
		x,
		y,
		x-size*math.Cos(angle-math.Pi/6),
		y-size*math.Sin(angle-math.Pi/6),
	)

	dc.DrawLine(
		x,
		y,
		x-size*math.Cos(angle+math.Pi/6),
		y-size*math.Sin(angle+math.Pi/6),
	)

	dc.Stroke()
}

func assignLevels(nodes map[string]*Node, fks []domain.ForeignKey) {

	hasParent := map[string]bool{}

	for _, fk := range fks {
		hasParent[fk.Table] = true
	}

	queue := []string{}

	for name := range nodes {
		if !hasParent[name] {
			nodes[name].Level = 0
			queue = append(queue, name)
		}
	}

	for len(queue) > 0 {

		current := queue[0]
		queue = queue[1:]

		node := nodes[current]

		for _, child := range node.Children {

			nodes[child].Level = node.Level + 1
			queue = append(queue, child)
		}
	}
}

func getCardinality(tables []domain.Table, fks []domain.ForeignKey, fk domain.ForeignKey) string {

	hasReverse := false

	for _, f := range fks {
		if f.ReferencedTable == fk.Table && f.Table == fk.ReferencedTable {
			hasReverse = true
			break
		}
	}

	if hasReverse {
		return "*..*"
	}

	for _, table := range tables {

		if table.Name == fk.Table {

			for _, col := range table.Columns {

				if col.Name == fk.Column {

					if col.IsUnique {
						return "1..1"
					}

					break
				}
			}

			break
		}
	}

	return "1..*"
}

func isManyToMany(tables []domain.Table, fks []domain.ForeignKey, tableName string) bool {

	var table domain.Table

	for _, t := range tables {
		if t.Name == tableName {
			table = t
			break
		}
	}

	if len(table.Columns) != 2 {
		return false
	}

	fkCount := 0

	for _, fk := range fks {
		if fk.Table == tableName {
			fkCount++
		}
	}

	return fkCount == 2
}
