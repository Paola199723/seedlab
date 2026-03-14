package generator

import (
	"fmt"
	"seedlab/internal/domain"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(tables []domain.Table, filename string) error {
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

		// Column names in row 1 horizontally
		for j, col := range table.Columns {
			colLetter := string(rune('A' + j))
			f.SetCellValue(sheetName, fmt.Sprintf("%s1", colLetter), col.Name)
		}

		if i == 0 {
			f.SetActiveSheet(index)
		}
	}

	// Remove default sheet
	f.DeleteSheet("Sheet1")

	return f.SaveAs(filename)
}