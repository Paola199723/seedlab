package generator

import (
	"encoding/xml"
	"fmt"
	"os"
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

func GenerateDraw(tables []domain.Table, fks []domain.ForeignKey, filename string) error {

	mxfile := Mxfile{
		Host: "app.diagrams.net",
		Diagram: Diagram{
			ID:   "schema",
			Name: "UML Schema",
			MxGraphModel: MxGraphModel{
				Dx:         "1200",
				Dy:         "800",
				Grid:       "1",
				GridSize:   "10",
				Page:       "1",
				PageScale:  "1",
				PageWidth:  "850",
				PageHeight: "1100",
				Root: Root{
					Cells: []Cell{
						{ID: "0"},
						{ID: "1", Parent: "0"},
					},
				},
			},
		},
	}

	tableIDs := map[string]string{}

	y := 40

	//--------------------------------
	// Crear tablas UML
	//--------------------------------

	for i, table := range tables {

		id := fmt.Sprintf("%d", i+2)
		tableIDs[table.Name] = id

		// texto UML
		value := fmt.Sprintf("<b>%s</b><br>", table.Name)

		for _, col := range table.Columns {
			value += fmt.Sprintf("+ %s : %s<br>", col.Name, col.Type)
		}

		// calcular altura dinámica
		rowHeight := 22
		headerHeight := 30
		height := headerHeight + (len(table.Columns) * rowHeight)

		// CUADRO de la tabla
		tableCell := Cell{
			ID:     id,
			Value:  table.Name,
			Style:  "shape=swimlane;fontStyle=1;startSize=28;html=1;",
			Vertex: "1",
			Parent: "1",
			Geometry: &Geometry{
				X:      "40",
				Y:      fmt.Sprintf("%d", y),
				Width:  "260",
				Height: fmt.Sprintf("%d", height),
				As:     "geometry",
			},
		}

		// TEXTO interno (columnas)
		textCell := Cell{
			ID:     id + "_text",
			Value:  value,
			Style:  "text;align=left;verticalAlign=top;spacingLeft=8;spacingTop=4;html=1;",
			Vertex: "1",
			Parent: id,
			Geometry: &Geometry{
				X:      "0",
				Y:      "28",
				Width:  "260",
				Height: fmt.Sprintf("%d", height-28),
				As:     "geometry",
			},
		}

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, tableCell)

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, textCell)

		y += height + 40
	}

	//--------------------------------
	// Crear relaciones UML
	//--------------------------------

	for i, fk := range fks {

		sourceID := tableIDs[fk.Table]
		targetID := tableIDs[fk.ReferencedTable]

		if sourceID == "" || targetID == "" {
			continue
		}

		cell := Cell{
			ID:     fmt.Sprintf("fk%d", i),
			Value:  "1..n",
			Edge:   "1",
			Style:  "endArrow=block;endFill=1;edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;",
			Parent: "1",
			Source: sourceID,
			Target: targetID,
			Geometry: &Geometry{
				As: "geometry",
			},
		}

		mxfile.Diagram.MxGraphModel.Root.Cells =
			append(mxfile.Diagram.MxGraphModel.Root.Cells, cell)
	}

	//--------------------------------
	// Guardar archivo
	//--------------------------------

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	return encoder.Encode(mxfile)
}
