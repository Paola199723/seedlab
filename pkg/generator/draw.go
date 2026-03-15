package generator

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"seedlab/internal/domain"
)

type Mxfile struct {
	XMLName xml.Name `xml:"mxfile"`
	Host    string   `xml:"host,attr,omitempty"`
	Diagram Diagram  `xml:"diagram"`
}

type Diagram struct {
	XMLName      xml.Name     `xml:"diagram"`
	ID           string       `xml:"id,attr"`
	Name         string       `xml:"name,attr"`
	MxGraphModel MxGraphModel `xml:"mxGraphModel"`
}

type MxGraphModel struct {
	XMLName    xml.Name `xml:"mxGraphModel"`
	Dx         string   `xml:"dx,attr"`
	Dy         string   `xml:"dy,attr"`
	Grid       string   `xml:"grid,attr"`
	GridSize   string   `xml:"gridSize,attr"`
	Page       string   `xml:"page,attr"`
	PageScale  string   `xml:"pageScale,attr"`
	PageWidth  string   `xml:"pageWidth,attr"`
	PageHeight string   `xml:"pageHeight,attr"`
	Root       Root     `xml:"root"`
}

type Root struct {
	XMLName xml.Name `xml:"root"`
	Cells   []Cell   `xml:"mxCell"`
}

type Cell struct {
	XMLName  xml.Name  `xml:"mxCell"`
	ID       string    `xml:"id,attr"`
	Value    string    `xml:"value,attr,omitempty"`
	Style    string    `xml:"style,attr,omitempty"`
	Parent   string    `xml:"parent,attr,omitempty"`
	Vertex   string    `xml:"vertex,attr,omitempty"`
	Edge     string    `xml:"edge,attr,omitempty"`
	Source   string    `xml:"source,attr,omitempty"`
	Target   string    `xml:"target,attr,omitempty"`
	Geometry *Geometry `xml:"mxGeometry,omitempty"`
}

type Geometry struct {
	XMLName xml.Name `xml:"mxGeometry"`
	X       string   `xml:"x,attr,omitempty"`
	Y       string   `xml:"y,attr,omitempty"`
	Width   string   `xml:"width,attr,omitempty"`
	Height  string   `xml:"height,attr,omitempty"`
	As      string   `xml:"as,attr"`
}

func buildRelationMap(fks []domain.ForeignKey) map[string][]string {

	relations := map[string][]string{}

	for _, fk := range fks {

		relations[fk.Table] =
			append(relations[fk.Table], fk.ReferencedTable)

		relations[fk.ReferencedTable] =
			append(relations[fk.ReferencedTable], fk.Table)

	}

	return relations
}

func GenerateDraw(
	tables []domain.Table,
	fks []domain.ForeignKey,
	filename string,
) error {

	relationMap := buildRelationMap(fks)

	maxColumns := 0

	for _, t := range tables {
		if len(t.Columns) > maxColumns {
			maxColumns = len(t.Columns)
		}
	}

	rowHeight := 22
	headerHeight := 30

	tableHeight := headerHeight + (maxColumns * rowHeight)

	tableWidth := 260

	nodeGapX := tableWidth + 160
	levelGapY := tableHeight + 200

	canvasWidth := len(tables) * nodeGapX

	startX := canvasWidth / 2
	startY := tableHeight

	tableIDs := map[string]string{}
	positions := map[string][2]int{}
	visited := map[string]bool{}

	if len(tables) > 0 {

		queue := []string{tables[0].Name}

		positions[tables[0].Name] = [2]int{startX, startY}
		visited[tables[0].Name] = true

		level := 1

		for len(queue) > 0 {

			nextQueue := []string{}
			x := startX - (len(queue)/2)*nodeGapX

			for _, tableName := range queue {

				for _, rel := range relationMap[tableName] {

					if visited[rel] {
						continue
					}

					posX := x
					posY := startY + level*levelGapY

					positions[rel] = [2]int{posX, posY}

					x += nodeGapX

					visited[rel] = true
					nextQueue = append(nextQueue, rel)

				}

			}

			queue = nextQueue
			level++

		}

	}

	mxfile := Mxfile{
		Host: "app.diagrams.net",
		Diagram: Diagram{
			ID:   "schema",
			Name: "ERD",
			MxGraphModel: MxGraphModel{
				Dx:         "2000",
				Dy:         "1600",
				Grid:       "1",
				GridSize:   "10",
				Page:       "1",
				PageScale:  "1",
				PageWidth:  "2400",
				PageHeight: "2000",
				Root: Root{
					Cells: []Cell{
						{ID: "0"},
						{ID: "1", Parent: "0"},
					},
				},
			},
		},
	}

	for i, table := range tables {

		id := fmt.Sprintf("t%d", i)
		tableIDs[table.Name] = id

		pos := positions[table.Name]

		if pos == [2]int{0, 0} {
			pos = [2]int{40 + i*nodeGapX, 40}
		}

		height := headerHeight + (len(table.Columns) * rowHeight)

		text := ""

		for _, col := range table.Columns {

			text += fmt.Sprintf(
				"+ %s : %s<br>",
				col.Name,
				col.Type,
			)

		}

		tableCell := Cell{
			ID:     id,
			Value:  table.Name,
			Style:  "shape=swimlane;fontStyle=1;startSize=30;whiteSpace=wrap;html=1;",
			Vertex: "1",
			Parent: "1",
			Geometry: &Geometry{
				X:      fmt.Sprintf("%d", pos[0]),
				Y:      fmt.Sprintf("%d", pos[1]),
				Width:  fmt.Sprintf("%d", tableWidth),
				Height: fmt.Sprintf("%d", height),
				As:     "geometry",
			},
		}

		textCell := Cell{
			ID:     id + "_text",
			Value:  text,
			Style:  "text;align=left;verticalAlign=top;spacingLeft=8;spacingTop=4;whiteSpace=wrap;html=1;",
			Vertex: "1",
			Parent: id,
			Geometry: &Geometry{
				X:      "0",
				Y:      "30",
				Width:  fmt.Sprintf("%d", tableWidth),
				Height: fmt.Sprintf("%d", height-30),
				As:     "geometry",
			},
		}

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, tableCell)

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, textCell)

	}

	for i, fk := range fks {

		source := tableIDs[fk.Table]
		target := tableIDs[fk.ReferencedTable]

		if source == "" || target == "" {
			continue
		}

		cell := Cell{
			ID:     fmt.Sprintf("fk%d", i),
			Edge:   "1",
			Style:  "endArrow=block;edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;",
			Parent: "1",
			Source: source,
			Target: target,
			Geometry: &Geometry{
				As: "geometry",
			},
		}

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, cell)

	}

	//--------------------------------
	// crear carpeta draw
	//--------------------------------

	folder := "draw"

	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}

	// generar nombre versionado
	finalName := fmt.Sprintf("%s.draw", filename)

	outputPath := filepath.Join(folder, finalName)

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	return encoder.Encode(mxfile)
}
