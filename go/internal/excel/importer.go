package excel

import (
	"fmt"
	"io"
	"strings"

	"chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

// Importer handles the logic of importing data from an Excel file.
type Importer struct {
	fileReader io.Reader
	request    *models.Request
	file       *excelize.File
}

// NewImporter creates a new instance of an Importer.
func NewImporter(reader io.Reader, request *models.Request) *Importer {
	return &Importer{
		fileReader: reader,
		request:    request,
	}
}

// Process executes the import operation and returns the extracted data.
func (i *Importer) Process() (*models.ImportResult, error) {
	var err error
	i.file, err = excelize.OpenReader(i.fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer i.file.Close()

	result := &models.ImportResult{
		Data:   make(map[string]map[string][]map[string]interface{}),
		Errors: make([]string, 0),
	}

	for _, sheetData := range i.request.Sheets {
		sheetName, err := i.getTargetSheet(sheetData)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.Data[sheetName] = make(map[string][]map[string]interface{})

		// Process Blocks, using the sheet's default columns
		for _, block := range sheetData.Blocks {
			blockData, err := i.readBlockData(sheetName, block.StartCell, block.Orientation, sheetData.DefaultColumns)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading block at %s on sheet %s: %v", block.StartCell, sheetName, err))
				continue
			}
			blockIdentifier := fmt.Sprintf("block_%s", block.StartCell)
			result.Data[sheetName][blockIdentifier] = blockData
		}

		// Process Tables, using the table's own columns
		for _, table := range sheetData.Tables {
			tableData, err := i.readBlockData(sheetName, table.StartCell, table.Orientation, table.Columns)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading table at %s on sheet %s: %v", table.StartCell, sheetName, err))
				continue
			}
			tableIdentifier := fmt.Sprintf("table_%s", table.StartCell)
			result.Data[sheetName][tableIdentifier] = tableData
		}
	}

	return result, nil
}

func (i *Importer) readBlockData(sheetName, startCell string, orientation models.Orientation, columns []models.Column) ([]map[string]interface{}, error) {
	col, row, err := excelize.CellNameToCoordinates(startCell)
	if err != nil { return nil, err }

	var allRowsData []map[string]interface{}

	// Convention: Read rows until we hit a completely empty row.
	for rowOffset := 0; ; rowOffset++ {
		isRowEmpty := true
		rowData := make(map[string]interface{})

		for colOffset, colDef := range columns {
			var cellCoords string
			if orientation == models.Horizontal {
				cellCoords, _ = excelize.CoordinatesToCellName(col+rowOffset, row+colOffset)
			} else {
				cellCoords, _ = excelize.CoordinatesToCellName(col+colOffset, row+rowOffset)
			}

			cellValue, err := i.file.GetCellValue(sheetName, cellCoords)
			if err != nil {
				// This might just mean the cell is out of bounds, which is fine.
				continue
			}
			if strings.TrimSpace(cellValue) != "" {
				isRowEmpty = false
			}

			// We could add validation here by calling a validateAndConvert function
			// and returning errors along with the data. For now, just raw data.
			rowData[colDef.Name] = cellValue
		}

		if isRowEmpty {
			break // Stop when we find an empty row.
		}
		allRowsData = append(allRowsData, rowData)
	}

	return allRowsData, nil
}

func (i *Importer) getTargetSheet(sheet models.Sheet) (string, error) {
	f := i.file
	name := sheet.Name
	index := sheet.Index

	if name != "" {
		if _, err := f.GetSheetIndex(name); err != nil {
			return "", fmt.Errorf("sheet '%s' not found", name)
		}
		return name, nil
	}
	if index > 0 {
		sheetList := f.GetSheetList()
		if index < len(sheetList) { return sheetList[index], nil }
		return "", fmt.Errorf("sheet index %d is out of bounds", index)
	}
	if len(f.GetSheetList()) > 0 {
		return f.GetSheetName(f.GetActiveSheetIndex()), nil
	}
	return "", fmt.Errorf("no sheets found in the workbook")
}
