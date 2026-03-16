package generator

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func FakeValue(colName, colType string, row int) any {

	colType = strings.ToLower(colType)

	switch {

	case strings.Contains(colType, "integer"):
		return row

	case strings.Contains(colType, "numeric"):
		return rand.Intn(1000) + 1

	case strings.Contains(colType, "boolean"):
		return rand.Intn(2) == 1

	case strings.Contains(colType, "timestamp"):
		return time.Now().Format("2006-01-02 15:04:05")

	case strings.Contains(colType, "character varying"),
		strings.Contains(colType, "text"):

		return FakeText(colName, row)

	default:
		return ""
	}
}

func FakeText(colName string, row int) string {

	name := strings.ToLower(colName)

	switch {

	case strings.Contains(name, "email"):
		return fmt.Sprintf("user%d@test.com", row)

	case strings.Contains(name, "name"):
		return fmt.Sprintf("User %d", row)

	case strings.Contains(name, "slug"):
		return fmt.Sprintf("slug-%d", row)

	case strings.Contains(name, "password"):
		return "password123"

	case strings.Contains(name, "url"),
		strings.Contains(name, "image"),
		strings.Contains(name, "img"):
		return fmt.Sprintf("https://img/%d.jpg", row)

	case strings.Contains(name, "description"):
		return fmt.Sprintf("description %d", row)

	default:
		return fmt.Sprintf("value_%d", row)
	}
}

func fakeValue(colName string, colType string, row int) any {

	colType = strings.ToLower(colType)
	colName = strings.ToLower(colName)

	switch {

	case strings.Contains(colType, "integer"):
		return row

	case strings.Contains(colType, "numeric"):
		return rand.Intn(1000) + 1

	case strings.Contains(colType, "boolean"):
		return rand.Intn(2) == 1

	case strings.Contains(colType, "timestamp"):
		return time.Now().Format("2006-01-02 15:04:05")

	case strings.Contains(colType, "character varying"),
		strings.Contains(colType, "text"):

		return FakeText(colName, row)

	default:
		return ""
	}
}