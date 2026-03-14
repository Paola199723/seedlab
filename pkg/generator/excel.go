package generator

import (
	"os"
	"path/filepath"
	"seedlab/internal/domain"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(tables []domain.Table, filename string) error {

	// crear carpeta excel si no existe
	excelDir := "excel"

	err := os.MkdirAll(excelDir, os.ModePerm)
	if err != nil {
		return err
	}

	// ruta final del archivo
	filePath := filepath.Join(excelDir, filename)

	f := excelize.NewFile()
	defer f.Close()

	for i, table := range tables {

		sheetName := table.Name

		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}

		index, err := f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		//--------------------------------
		// headers columnas
		//--------------------------------

		for j, col := range table.Columns {

			cell, _ := excelize.CoordinatesToCellName(j+1, 1)

			f.SetCellValue(sheetName, cell, col.Name)
		}

		if i == 0 {
			f.SetActiveSheet(index)
		}
	}

	//--------------------------------
	// eliminar hoja default
	//--------------------------------

	f.DeleteSheet("Sheet1")

	//--------------------------------
	// guardar archivo
	//--------------------------------

	return f.SaveAs(filePath)
}
