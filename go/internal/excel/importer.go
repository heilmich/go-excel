package excel

import (
	"fmt"
	"io"
	"strings"

	"chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

type Importer struct {
	fileReader io.Reader
	request    *models.Request
	file       *excelize.File
}

func NewImporter(reader io.Reader, request *models.Request) *Importer {
	return &Importer{
		fileReader: reader,
		request:    request,
	}
}

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

		for _, block := range sheetData.Blocks {
			blockData, err := i.readBlockData(sheetName, block)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading block on sheet %s: %v", sheetName, err))
				continue
			}
			blockIdentifier := block.Id
			if blockIdentifier == "" {
				blockIdentifier = fmt.Sprintf("block_%s", block.StartCell)
			}
			result.Data[sheetName][blockIdentifier] = blockData
		}
	}

	return result, nil
}

func (i *Importer) readBlockData(sheetName string, block models.Block) ([]map[string]interface{}, error) {
	startCol, startRow, err := getStartCoordinates(block)
	if err != nil { return nil, err }

	var allRowsData []map[string]interface{}

	for rowOffset := 0; ; rowOffset++ {
		isRowEmpty := true
		rowData := make(map[string]interface{})

		for colOffset, colDef := range block.Columns {
			var cellCoords string
			if block.Orientation == models.Horizontal {
				cellCoords, _ = excelize.CoordinatesToCellName(startCol+rowOffset, startRow+colOffset)
			} else {
				cellCoords, _ = excelize.CoordinatesToCellName(startCol+colOffset, startRow+rowOffset)
			}

			cellValue, _ := i.file.GetCellValue(sheetName, cellCoords)
			if strings.TrimSpace(cellValue) != "" {
				isRowEmpty = false
			}
			rowData[colDef.Name] = cellValue
		}

		if isRowEmpty {
			break
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
