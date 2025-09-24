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
	request := &models.Request{
		Sheets: []models.Sheet{
			{
				Name: "TestSheet",
				Blocks: []models.Block{
					{
						StartCell:   "A1",
						ShowHeaders: true, // This means data will start on row 2
						Columns: []models.Column{
							{Name: "name", Header: "Name", Type: models.TypeString},
							{Name: "value", Header: "Value", Type: models.TypeInt},
						},
						Data: []map[string]interface{}{
							{"name": "Alice", "value": 100},
							{"name": "Bob", "value": 200},
						},
					},
				},
				Options: models.Options{
					Freeze: "A2",
				},
			},
		},
	}

	exporter := NewExporter(request, nil)
	resultBytes, err := exporter.Process()

	require.NoError(t, err)
	require.NotEmpty(t, resultBytes)

	f, err := excelize.OpenReader(bytes.NewReader(resultBytes))
	require.NoError(t, err)
	defer f.Close()

	assert.Equal(t, "TestSheet", f.GetSheetName(0))

	// Headers are on row 1
	header1, _ := f.GetCellValue("TestSheet", "A1")
	assert.Equal(t, "Name", header1)

	// Data starts on row 2
	cellA2, _ := f.GetCellValue("TestSheet", "A2")
	assert.Equal(t, "Alice", cellA2)

	cellB3, _ := f.GetCellValue("TestSheet", "B3")
	assert.Equal(t, "200", cellB3) // No custom format, so it's a plain string representation

	panes, err := f.GetPanes("TestSheet")
	require.NoError(t, err)
	assert.True(t, panes.Freeze)
	assert.Equal(t, "A2", panes.TopLeftCell)
}

func TestExporter_Process_InvalidRequest(t *testing.T) {
	request := &models.Request{
		Sheets: []models.Sheet{
			{
				Name: "TestSheet",
				Blocks: []models.Block{
					{
						StartCell: "INVALID_CELL_123",
						Data:      []map[string]interface{}{{"name": "data"}},
						Columns:   []models.Column{{Name: "name"}},
					},
				},
			},
		},
	}

	exporter := NewExporter(request, nil)
	_, err := exporter.Process()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cell name")
}
