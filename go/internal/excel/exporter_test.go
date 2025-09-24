package excel

import (
	"bytes"
	"testing"

	"chelbit/excelms/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestExporter_Process_Valid(t *testing.T) {
	request := models.ExportRequest{
		Sheet: "TestSheet",
		Columns: []models.Column{
			{Name: "name", Header: "Name", Type: models.TypeString, Width: 20},
			{Name: "value", Header: "Value", Type: models.TypeInt, CustomFormat: "0.0"},
		},
		HeaderRow: 1,
		StartRow:  2,
		Rows: []map[string]interface{}{
			{"name": "Alice", "value": 100},
			{"name": "Bob", "value": 200},
		},
		Options: models.Options{
			Freeze: "A2",
		},
	}

	exporter := NewExporter(request, nil)
	resultBytes, err := exporter.Process()

	// 1. Check for processing errors
	require.NoError(t, err)
	require.NotEmpty(t, resultBytes)

	// 2. Read the generated file back to verify its contents
	f, err := excelize.OpenReader(bytes.NewReader(resultBytes))
	require.NoError(t, err)
	defer f.Close()

	// Check sheet name
	assert.Equal(t, "TestSheet", f.GetSheetName(0))

	// Check header values
	header1, _ := f.GetCellValue("TestSheet", "A1")
	header2, _ := f.GetCellValue("TestSheet", "B1")
	assert.Equal(t, "Name", header1)
	assert.Equal(t, "Value", header2)

	// Check data values
	cellA2, _ := f.GetCellValue("TestSheet", "A2")
	cellB3, _ := f.GetCellValue("TestSheet", "B3")
	assert.Equal(t, "Alice", cellA2)
	assert.Equal(t, "200.0", cellB3) // The custom format "0.0" makes it a float string.

	// Check freeze panes
	panes, err := f.GetPanes("TestSheet")
	require.NoError(t, err)
	assert.True(t, panes.Freeze)
	assert.Equal(t, "A2", panes.TopLeftCell)
}

func TestExporter_Process_InvalidRequest(t *testing.T) {
	// An invalid request, e.g., with a bad cell coordinate
	request := models.ExportRequest{
		Columns: []models.Column{
			{Name: "name", Header: "Name", Type: models.TypeString},
		},
		Blocks: []models.Block{
			{
				StartCell: "INVALID_CELL_123", // This should cause an error
				Data: []map[string]interface{}{
					{"name": "data"},
				},
			},
		},
	}

	exporter := NewExporter(request, nil)
	_, err := exporter.Process()

	// Assert that an error was returned and it's about the invalid cell
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid startCell")
}
