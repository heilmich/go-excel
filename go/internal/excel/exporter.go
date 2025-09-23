package excel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

// Exporter handles the logic of exporting data to an Excel file.
type Exporter struct {
	request       *models.Request
	templateBytes []byte
	file          *excelize.File
}

// NewExporter creates a new instance of an Exporter.
func NewExporter(request *models.Request, templateBytes []byte) *Exporter {
	return &Exporter{
		request:       request,
		templateBytes: templateBytes,
	}
}

// Process executes the export operation and returns the generated file as a byte slice.
func (e *Exporter) Process() ([]byte, error) {
	var err error

	if len(e.templateBytes) > 0 {
		e.file, err = excelize.OpenReader(bytes.NewReader(e.templateBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to open template: %w", err)
		}
	} else {
		e.file = excelize.NewFile()
	}
	defer e.file.Close()

	// Process each sheet defined in the request
	for _, sheetData := range e.request.Sheets {
		sheetName, err := e.getTargetSheet(sheetData)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare sheet '%s': %w", sheetData.Name, err)
		}

		styleCache, err := e.createStyleCache(sheetData)
		if err != nil {
			return nil, fmt.Errorf("failed to create style cache for sheet '%s': %w", sheetName, err)
		}

		// Process Blocks
		for _, block := range sheetData.Blocks {
			startCol, startRow, err := getStartCoordinates(block)
			if err != nil { return nil, fmt.Errorf("sheet '%s': invalid block coordinates: %w", sheetName, err) }

			orientation := block.Orientation
			if orientation == "" { orientation = models.Vertical }

			err = e.writeBlock(sheetName, sheetData.DefaultColumns, block.Data, startCol, startRow, block.ShowHeaders, orientation, styleCache)
			if err != nil { return nil, fmt.Errorf("sheet '%s': failed to write block at %s: %w", sheetName, block.StartCell, err) }
		}

		// Process Tables
		for _, table := range sheetData.Tables {
		// Create a temporary Block struct from the Table to pass to getStartCoordinates
		blockForCoords := models.Block{StartCell: table.StartCell, StartRow: table.StartRow, StartCol: table.StartCol}
		startCol, startRow, err := getStartCoordinates(blockForCoords)
			if err != nil { return nil, fmt.Errorf("sheet '%s': invalid table coordinates: %w", sheetName, err) }

			orientation := table.Orientation
			if orientation == "" { orientation = models.Vertical }

			err = e.writeBlock(sheetName, table.Columns, table.Data, startCol, startRow, table.ShowHeaders, orientation, styleCache)
			if err != nil { return nil, fmt.Errorf("sheet '%s': failed to write table at %s: %w", sheetName, table.StartCell, err) }
		}

		if err := e.applyPostWriteOptions(sheetName, sheetData); err != nil {
			return nil, fmt.Errorf("failed to apply post-write options for sheet '%s': %w", sheetName, err)
		}
	}

	// If the default "Sheet1" was never used and still exists, delete it.
	if len(e.request.Sheets) > 0 && e.file.SheetCount > len(e.request.Sheets) {
		if _, err := e.file.GetSheetIndex("Sheet1"); err == nil {
			e.file.DeleteSheet("Sheet1")
		}
	}

	buf, err := e.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write to buffer: %w", err)
	}
	return buf.Bytes(), nil
}

// ### Helper methods for Exporter ###

func (e *Exporter) getTargetSheet(sheet models.Sheet) (string, error) {
	f := e.file
	name := sheet.Name
	index := sheet.Index

	if name != "" {
		if idx, err := f.GetSheetIndex(name); err == nil {
			f.SetActiveSheet(idx)
			return name, nil
		}
		idx, err := f.NewSheet(name)
		if err != nil {
			return "", fmt.Errorf("failed to create new sheet '%s': %w", name, err)
		}
		f.SetActiveSheet(idx)
		return name, nil
	}

	if index > 0 {
		sheetList := f.GetSheetList()
		if index < len(sheetList) { return sheetList[index], nil }
		return "", fmt.Errorf("sheet index %d is out of bounds", index)
	}

	// If no name or index, and it's the first sheet being processed, rename "Sheet1".
	if f.SheetCount == 1 && f.GetSheetName(0) == "Sheet1" {
		// This case is tricky in a loop. Let's require a name for now.
		return "", fmt.Errorf("first sheet in a multi-sheet request must have a name")
	}

	return f.GetSheetName(f.GetActiveSheetIndex()), nil
}

func (e *Exporter) writeBlock(sheetName string, cols []models.Column, data []map[string]interface{}, startCol, startRow int, showHeaders bool, orientation models.Orientation, styleCache map[string]int) error {
	f := e.file
	headerRow := startRow - 1
	if headerRow < 1 { headerRow = 1 }
	dataStartRow := startRow

	if showHeaders {
		for i, colDef := range cols {
			var cell string
			var err error
			if orientation == models.Horizontal {
				cell, err = excelize.CoordinatesToCellName(startCol, headerRow+i)
			} else {
				cell, err = excelize.CoordinatesToCellName(startCol+i, headerRow)
			}
			if err != nil { return fmt.Errorf("failed to get header cell for col %d: %w", i, err) }
			if err := f.SetCellValue(sheetName, cell, colDef.Header); err != nil { return fmt.Errorf("failed to set header value for cell %s: %w", cell, err) }
		}
	}

	for rowIndex, dataRow := range data {
		for colIndex, colDef := range cols {
			value, _ := dataRow[colDef.Name]
			var cell string
			var err error
			if orientation == models.Horizontal {
				cell, err = excelize.CoordinatesToCellName(startCol+rowIndex, dataStartRow+colIndex)
			} else {
				cell, err = excelize.CoordinatesToCellName(startCol+colIndex, dataStartRow+rowIndex)
			}
			if err != nil { return fmt.Errorf("failed to get data cell for row %d, col %d: %w", rowIndex, colIndex, err) }
			if err := f.SetCellValue(sheetName, cell, value); err != nil { return fmt.Errorf("failed to set data value for cell %s: %w", cell, err) }
			if styleID, ok := styleCache[colDef.CustomFormat]; ok && colDef.CustomFormat != "" {
				if err := f.SetCellStyle(sheetName, cell, cell, styleID); err != nil { return fmt.Errorf("failed to set style for cell %s: %w", cell, err) }
			}
		}
	}
	return nil
}

func (e *Exporter) createStyleCache(sheet models.Sheet) (map[string]int, error) {
	cache := make(map[string]int)
	allCols := make([]models.Column, 0)
	allCols = append(allCols, sheet.DefaultColumns...)
	for _, table := range sheet.Tables {
		allCols = append(allCols, table.Columns...)
	}
	for _, col := range allCols {
		if col.CustomFormat != "" {
			if _, exists := cache[col.CustomFormat]; !exists {
				style, err := e.file.NewStyle(&excelize.Style{CustomNumFmt: &col.CustomFormat})
				if err != nil { return nil, fmt.Errorf("failed to create style for format '%s': %w", col.CustomFormat, err) }
				cache[col.CustomFormat] = style
			}
		}
	}
	return cache, nil
}

func (e *Exporter) applyPostWriteOptions(sheetName string, sheetData models.Sheet) error {
	f := e.file
	colMap := make(map[string]models.Column)
	for i, col := range sheetData.DefaultColumns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		colMap[colName] = col
	}
	for _, table := range sheetData.Tables {
		blockForCoords := models.Block{StartCell: table.StartCell, StartRow: table.StartRow, StartCol: table.StartCol}
		startCol, _, _ := getStartCoordinates(blockForCoords)
		for i, col := range table.Columns {
			colName, _ := excelize.ColumnNumberToName(startCol + i)
			colMap[colName] = col
		}
	}

	for colName, col := range colMap {
		if col.Width > 0 {
			if err := f.SetColWidth(sheetName, colName, colName, col.Width); err != nil { return fmt.Errorf("failed to set width for column %s: %w", colName, err) }
		}
	}

	for _, formula := range sheetData.Formulas {
		if formula.Column == "" || formula.Expr == "" || formula.StartRow <= 0 { continue }
		if formula.EndRow > 0 {
			for r := formula.StartRow; r <= formula.EndRow; r++ {
				cell := fmt.Sprintf("%s%d", formula.Column, r)
				formulaStr := strings.ReplaceAll(formula.Expr, "{row}", strconv.Itoa(r))
				if err := f.SetCellFormula(sheetName, cell, formulaStr); err != nil { return fmt.Errorf("failed to set formula for cell %s: %w", cell, err) }
			}
		}
	}

	if sheetData.Options.Freeze != "" {
		if err := f.SetPanes(sheetName, &excelize.Panes{Freeze: true, TopLeftCell: sheetData.Options.Freeze}); err != nil { return fmt.Errorf("failed to set freeze panes at %s: %w", sheetData.Options.Freeze, err) }
	}
	if sheetData.Options.AutoFilter != "" {
		if err := f.AutoFilter(sheetName, sheetData.Options.AutoFilter, nil); err != nil { return fmt.Errorf("failed to set autofilter on range %s: %w", sheetData.Options.AutoFilter, err) }
	}
	return nil
}

func getStartCoordinates(block models.Block) (int, int, error) {
	if block.StartCell != "" {
		col, row, err := excelize.CellNameToCoordinates(block.StartCell)
		if err != nil { return 0, 0, fmt.Errorf("invalid startCell '%s': %w", block.StartCell, err) }
		return col, row, nil
	}
	if block.StartRow > 0 && block.StartCol != "" {
		col, err := excelize.ColumnNameToNumber(block.StartCol)
		if err != nil { return 0, 0, fmt.Errorf("invalid startCol '%s': %w", block.StartCol, err) }
		return col, block.StartRow, nil
	}
	return 1, 1, nil
}
