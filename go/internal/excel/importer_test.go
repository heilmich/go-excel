package excel

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"chelbit/excelms/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func createTestExcelFile(t *testing.T) *bytes.Buffer {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	_ = f.SetCellValue(sheet, "A1", "Full Name")
	_ = f.SetCellValue(sheet, "B1", "Email")
	_ = f.SetCellValue(sheet, "C1", "Age")

	// Valid data
	_ = f.SetCellValue(sheet, "A2", "John Doe")
	_ = f.SetCellValue(sheet, "B2", "john@example.com")
	_ = f.SetCellValue(sheet, "C2", 30)

	// Invalid data for the second test case
	_ = f.SetCellValue(sheet, "A3", "") // Required field is empty
	_ = f.SetCellValue(sheet, "B3", "jane@example.com")
	_ = f.SetCellValue(sheet, "C3", "twenty") // Invalid integer

	buf, err := f.WriteToBuffer()
	require.NoError(t, err)
	return buf
}

func TestImporter_Process_Valid(t *testing.T) {
	excelFile := createTestExcelFile(t)

	schema := models.ImportSchema{
		Sheet:     "Sheet1",
		HeaderRow: 1,
		StartRow:  2,
		Columns: []models.Column{
			{Name: "name", Header: "Full Name", Type: models.TypeString, Required: true},
			{Name: "email", Header: "Email", Type: models.TypeString},
			{Name: "age", Header: "Age", Type: models.TypeInt},
		},
	}

	outputBuf := new(bytes.Buffer)
	importer := NewImporter(excelFile, schema, outputBuf)
	err := importer.Process()

	require.NoError(t, err)

	// Process the NDJSON output
	lines := strings.Split(strings.TrimSpace(outputBuf.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 1, "Should have at least one line of output")

	var firstResult models.ImportResultLine
	err = json.Unmarshal([]byte(lines[0]), &firstResult)
	require.NoError(t, err)

	// Check the first valid row
	assert.True(t, firstResult.OK)
	assert.Equal(t, 2, firstResult.RowIndex)
	assert.Empty(t, firstResult.Errors)
	assert.Equal(t, "John Doe", firstResult.Data["name"])
	assert.Equal(t, float64(30), firstResult.Data["age"]) // JSON unmarshals numbers to float64
}

func TestImporter_Process_Invalid(t *testing.T) {
	excelFile := createTestExcelFile(t)

	schema := models.ImportSchema{
		Sheet:     "Sheet1",
		HeaderRow: 1,
		StartRow:  2,
		Columns: []models.Column{
			{Name: "name", Header: "Full Name", Type: models.TypeString, Required: true},
			{Name: "age", Header: "Age", Type: models.TypeInt},
		},
	}

	outputBuf := new(bytes.Buffer)
	importer := NewImporter(excelFile, schema, outputBuf)
	err := importer.Process()

	require.NoError(t, err)

	// Process the NDJSON output
	lines := strings.Split(strings.TrimSpace(outputBuf.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 2, "Should have at least two lines of output")

	var secondResult models.ImportResultLine
	err = json.Unmarshal([]byte(lines[1]), &secondResult)
	require.NoError(t, err)

	// Check the second row with invalid data
	assert.False(t, secondResult.OK)
	assert.Equal(t, 3, secondResult.RowIndex)
	require.Len(t, secondResult.Errors, 2)

	// Check for specific errors
	assert.Equal(t, "name", secondResult.Errors[0].Field)
	assert.Equal(t, "required", secondResult.Errors[0].Code)

	assert.Equal(t, "age", secondResult.Errors[1].Field)
	assert.Equal(t, "type", secondResult.Errors[1].Code)
}
