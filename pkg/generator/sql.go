package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const batchSize = 500
const deleteBatch = 500

func GenerateInsertRollbackFromExcel(version int, excelPath string) error {

	basePath := getBasePath()

	file, err := excelize.OpenFile(filepath.Join("excel", excelPath))
	if err != nil {
		return err
	}
	defer file.Close()

	sheets := file.GetSheetList()

	if len(sheets) == 0 {
		return fmt.Errorf("no sheets found in excel")
	}

	insertDir := filepath.Join(basePath, fmt.Sprintf("insert_version%04d", version))
	rollbackDir := filepath.Join(basePath, fmt.Sprintf("rollback_version%04d", version))

	err = os.MkdirAll(insertDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(rollbackDir, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println("Insert folder:", insertDir)
	fmt.Println("Rollback folder:", rollbackDir)

	var insertFiles []string
	var rollbackFiles []string

	fileIndex := 1

	for _, sheet := range sheets {

		fmt.Println("Processing table:", sheet)

		rows, err := file.GetRows(sheet)
		if err != nil {
			return err
		}

		if len(rows) < 2 {
			continue
		}

		columns := rows[0]

		pkIndex := detectPrimaryKey(columns)
		pkColumn := columns[pkIndex]

		insertFileName := fmt.Sprintf("%04d_%s.sql", fileIndex, strings.ToLower(sheet))
		rollbackFileName := fmt.Sprintf("%04d_%s.sql", fileIndex, strings.ToLower(sheet))

		insertFilePath := filepath.Join(insertDir, insertFileName)
		rollbackFilePath := filepath.Join(rollbackDir, rollbackFileName)

		insertFile, err := os.Create(insertFilePath)
		if err != nil {
			return err
		}

		rollbackFile, err := os.Create(rollbackFilePath)
		if err != nil {
			return err
		}

		fmt.Fprintln(insertFile, "BEGIN;\n")
		fmt.Fprintln(rollbackFile, "BEGIN;\n")

		var batch []string
		var ids []string

		for _, row := range rows[1:] {

			if isRowEmpty(row) {
				continue
			}

			values := make([]string, len(columns))

			for i := range columns {

				var raw string

				if i < len(row) {
					raw = row[i]
				}

				values[i] = sqlValue(raw)
			}

			batch = append(batch, fmt.Sprintf("(%s)", strings.Join(values, ",")))
			ids = append(ids, values[pkIndex])

			if len(batch) >= batchSize {

				writeInsertBatch(insertFile, sheet, columns, batch)
				batch = []string{}
			}
		}

		if len(batch) > 0 {
			writeInsertBatch(insertFile, sheet, columns, batch)
		}

		writeRollback(rollbackFile, sheet, pkColumn, ids)

		fmt.Fprintln(insertFile, "\nCOMMIT;")
		fmt.Fprintln(rollbackFile, "\nCOMMIT;")

		insertFile.Close()
		rollbackFile.Close()

		insertFiles = append(insertFiles, insertFileName)
		rollbackFiles = append(rollbackFiles, rollbackFileName)

		fileIndex++
	}

	reverseSlice(rollbackFiles)

	generateMasterScript(insertDir, "run_insert.sql", insertFiles)
	generateMasterScript(rollbackDir, "run_rollback.sql", rollbackFiles)

	fmt.Println("SQL generation completed successfully")

	return nil
}

func writeInsertBatch(file *os.File, table string, columns []string, batch []string) {

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s)\nVALUES\n%s;\n\n",
		table,
		strings.Join(columns, ","),
		strings.Join(batch, ",\n"),
	)

	file.WriteString(sql)
}

func writeRollback(file *os.File, table string, pk string, ids []string) {

	for i := 0; i < len(ids); i += deleteBatch {

		end := i + deleteBatch

		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[i:end]

		sql := fmt.Sprintf(
			"DELETE FROM %s WHERE %s IN (%s);\n",
			table,
			pk,
			strings.Join(chunk, ","),
		)

		file.WriteString(sql)
	}
}

func detectPrimaryKey(columns []string) int {

	for i, col := range columns {
		if strings.ToLower(col) == "id" {
			return i
		}
	}

	return 0
}

func sqlValue(raw string) string {

	raw = strings.TrimSpace(raw)

	if raw == "" {
		return "NULL"
	}

	if isNumber(raw) {
		return raw
	}

	lower := strings.ToLower(raw)

	if lower == "true" || lower == "false" {
		return lower
	}

	raw = strings.ReplaceAll(raw, "'", "''")

	return fmt.Sprintf("'%s'", raw)
}

func isNumber(s string) bool {

	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}

	return true
}

func isRowEmpty(row []string) bool {

	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}

	return true
}

func reverseSlice(slice []string) {

	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func generateMasterScript(dir string, name string, files []string) {

	path := filepath.Join(dir, name)

	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	for _, f := range files {
		file.WriteString(fmt.Sprintf("\\i %s\n", f))
	}
}

func getBasePath() string {

	dir, err := os.Getwd()
	if err != nil {
		return "sql"
	}

	return filepath.Join(dir, "sql")
}

func getNextVersion(basePath, prefix string) int {

	files, err := os.ReadDir(basePath)

	if err != nil {
		return 1
	}

	maxVersion := 0

	for _, file := range files {

		name := file.Name()

		if strings.HasPrefix(name, prefix) {

			parts := strings.Split(name, "version")

			if len(parts) == 2 {

				v, err := strconv.Atoi(parts[1])

				if err == nil && v > maxVersion {
					maxVersion = v
				}
			}
		}
	}

	return maxVersion + 1
}
