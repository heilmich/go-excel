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
	request       models.ExportRequest
	templateBytes []byte
	file          *excelize.File
	sheetName     string
}

// NewExporter creates a new instance of an Exporter.
func NewExporter(request models.ExportRequest, templateBytes []byte) *Exporter {
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

	e.sheetName, err = e.getTargetSheet()
	if err != nil {
		return nil, err
	}

	styleCache, err := e.createStyleCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create style cache: %w", err)
	}

	// Process legacy `rows`
	if len(e.request.Rows) > 0 {
		headerRow := e.request.HeaderRow
		if headerRow == 0 {	headerRow = 1 }
		dataStartRow := e.request.StartRow
		if dataStartRow == 0 { dataStartRow = headerRow + 1 }
		err = e.writeBlock(e.request.Columns, e.request.Rows, 1, dataStartRow, headerRow, true, models.Vertical, styleCache)
		if err != nil {
			return nil, fmt.Errorf("failed to write legacy rows block: %w", err)
		}
	}

	// Process `blocks`
	for _, block := range e.request.Blocks {
		startCol, startRow, err := getStartCoordinates(block)
		if err != nil { return nil, fmt.Errorf("invalid block coordinates: %w", err) }
		headerRow := startRow - 1
		if headerRow < 1 { headerRow = 1 }
		orientation := block.Orientation
		if orientation == "" { orientation = models.Vertical }
		err = e.writeBlock(e.request.Columns, block.Data, startCol, startRow, headerRow, block.ShowHeaders, orientation, styleCache)
		if err != nil { return nil, fmt.Errorf("failed to write block starting at %s: %w", block.StartCell, err) }
	}

	// Process `tables`
	for _, table := range e.request.Tables {
		startCol, startRow, err := getStartCoordinates(table.Block)
		if err != nil { return nil, fmt.Errorf("invalid table coordinates: %w", err) }
		headerRow := startRow - 1
		if headerRow < 1 { headerRow = 1 }
		orientation := table.Orientation
		if orientation == "" { orientation = models.Vertical }
		err = e.writeBlock(table.Columns, table.Data, startCol, startRow, headerRow, table.ShowHeaders, orientation, styleCache)
		if err != nil { return nil, fmt.Errorf("failed to write table starting at %s: %w", table.StartCell, err) }
	}

	if err := e.applyPostWriteOptions(); err != nil {
		return nil, fmt.Errorf("failed to apply post-write options: %w", err)
	}

	buf, err := e.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write to buffer: %w", err)
	}
	return buf.Bytes(), nil
}

// ### Helper methods for Exporter ###

func (e *Exporter) getTargetSheet() (string, error) {
	name := e.request.Sheet
	index := e.request.SheetIndex
	f := e.file

	if name != "" {
		sheetIndex, err := f.GetSheetIndex(name)
		if err != nil {
			newIndex, err := f.NewSheet(name)
			if err != nil { return "", fmt.Errorf("failed to create sheet '%s': %w", name, err) }
			f.SetActiveSheet(newIndex)
			if len(f.GetSheetList()) > 1 {
				if _, err := f.GetSheetIndex("Sheet1"); err == nil && name != "Sheet1" {
					if rows, err := f.GetRows("Sheet1"); err == nil && len(rows) == 0 {
						_ = f.DeleteSheet("Sheet1")
					}
				}
			}
			return name, nil
		}
		f.SetActiveSheet(sheetIndex)
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
	return "Sheet1", nil
}

func (e *Exporter) writeBlock(cols []models.Column, data []map[string]interface{}, startCol, dataStartRow, headerRow int, showHeaders bool, orientation models.Orientation, styleCache map[string]int) error {
	f := e.file
	sheet := e.sheetName

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
			if err := f.SetCellValue(sheet, cell, colDef.Header); err != nil { return fmt.Errorf("failed to set header value for cell %s: %w", cell, err) }
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
			if err := f.SetCellValue(sheet, cell, value); err != nil { return fmt.Errorf("failed to set data value for cell %s: %w", cell, err) }
			if styleID, ok := styleCache[colDef.CustomFormat]; ok && colDef.CustomFormat != "" {
				if err := f.SetCellStyle(sheet, cell, cell, styleID); err != nil { return fmt.Errorf("failed to set style for cell %s: %w", cell, err) }
			}
		}
	}
	return nil
}

func (e *Exporter) createStyleCache() (map[string]int, error) {
	cache := make(map[string]int)
	allCols := make([]models.Column, 0)
	allCols = append(allCols, e.request.Columns...)
	for _, table := range e.request.Tables {
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

func (e *Exporter) applyPostWriteOptions() error {
	f := e.file
	sheet := e.sheetName
	request := e.request

	colMap := make(map[string]models.Column)
	for i, col := range request.Columns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		colMap[colName] = col
	}
	for _, table := range request.Tables {
		startCol, _, _ := getStartCoordinates(table.Block)
		for i, col := range table.Columns {
			colName, _ := excelize.ColumnNumberToName(startCol + i)
			colMap[colName] = col
		}
	}

	for colName, col := range colMap {
		if col.Width > 0 {
			if err := f.SetColWidth(sheet, colName, colName, col.Width); err != nil { return fmt.Errorf("failed to set width for column %s: %w", colName, err) }
		}
	}

	for _, formula := range request.Formulas {
		if formula.Column == "" || formula.Expr == "" || formula.StartRow <= 0 { continue }
		if formula.EndRow > 0 {
			for r := formula.StartRow; r <= formula.EndRow; r++ {
				cell := fmt.Sprintf("%s%d", formula.Column, r)
				formulaStr := strings.ReplaceAll(formula.Expr, "{row}", strconv.Itoa(r))
				if err := f.SetCellFormula(sheet, cell, formulaStr); err != nil { return fmt.Errorf("failed to set formula for cell %s: %w", cell, err) }
			}
		}
	}

	if request.Options.Freeze != "" {
		if err := f.SetPanes(sheet, &excelize.Panes{Freeze: true, TopLeftCell: request.Options.Freeze}); err != nil { return fmt.Errorf("failed to set freeze panes at %s: %w", request.Options.Freeze, err) }
	}
	if request.Options.AutoFilter != "" {
		if err := f.AutoFilter(sheet, request.Options.AutoFilter, nil); err != nil { return fmt.Errorf("failed to set autofilter on range %s: %w", request.Options.AutoFilter, err) }
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
