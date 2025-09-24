package excel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

type Exporter struct {
	request       *models.Request
	templateBytes []byte
	file          *excelize.File
}

func NewExporter(request *models.Request, templateBytes []byte) *Exporter {
	return &Exporter{
		request:       request,
		templateBytes: templateBytes,
	}
}

func (e *Exporter) Process() ([]byte, error) {
	var err error
	if len(e.templateBytes) > 0 {
		e.file, err = excelize.OpenReader(bytes.NewReader(e.templateBytes))
	} else {
		e.file = excelize.NewFile()
	}
	if err != nil { return nil, err }
	defer e.file.Close()

	for i, sheetData := range e.request.Sheets {
		sheetName, err := e.getTargetSheet(i, sheetData)
		if err != nil { return nil, err }

		styleCache, err := e.createStyleCache(sheetData)
		if err != nil { return nil, err }

		for _, block := range sheetData.Blocks {
			startCol, startRow, err := getStartCoordinates(block)
			if err != nil { return nil, err }
			err = e.writeBlock(sheetName, block, startCol, startRow, styleCache)
			if err != nil { return nil, err }
		}

		if err := e.applyPostWriteOptions(sheetName, sheetData); err != nil {
			return nil, err
		}
	}

	if len(e.request.Sheets) > 0 {
		if _, err := e.file.GetSheetIndex("Sheet1"); err == nil {
			isDefaultSheetUsed := false
			for _, s := range e.request.Sheets {
				if s.Name == "Sheet1" {
					isDefaultSheetUsed = true
					break
				}
			}
			if !isDefaultSheetUsed {
				e.file.DeleteSheet("Sheet1")
			}
		}
	}

	buf, err := e.file.WriteToBuffer()
	if err != nil { return nil, err }
	return buf.Bytes(), nil
}

func (e *Exporter) getTargetSheet(sheetIndex int, sheet models.Sheet) (string, error) {
	f := e.file
	name := sheet.Name

	if name != "" {
		if idx, err := f.GetSheetIndex(name); err == nil {
			f.SetActiveSheet(idx)
			return name, nil
		}
		idx, err := f.NewSheet(name)
		if err != nil { return "", err }
		f.SetActiveSheet(idx)
		return name, nil
	}

	if sheetIndex == 0 {
		return f.GetSheetName(0), nil
	}

	if sheet.Index > 0 {
		sheetList := f.GetSheetList()
		if sheet.Index < len(sheetList) { return sheetList[sheet.Index], nil }
		return "", fmt.Errorf("sheet index %d is out of bounds", sheet.Index)
	}

	return f.GetSheetName(f.GetActiveSheetIndex()), nil
}

func (e *Exporter) writeBlock(sheetName string, block models.Block, startCol, startRow int, styleCache map[string]int) error {
	f := e.file
	headerRow := startRow - 1
	if headerRow < 1 { headerRow = 1 }
	dataStartRow := startRow
	orientation := block.Orientation
	if orientation == "" { orientation = models.Vertical }

	if block.ShowHeaders {
		for i, colDef := range block.Columns {
			var cell string
			if orientation == models.Horizontal {
				cell, _ = excelize.CoordinatesToCellName(startCol, headerRow+i)
			} else {
				cell, _ = excelize.CoordinatesToCellName(startCol+i, headerRow)
			}
			if err := f.SetCellValue(sheetName, cell, colDef.Header); err != nil { return err }
		}
	}

	for rowIndex, dataRow := range block.Data {
		for colIndex, colDef := range block.Columns {
			value, _ := dataRow[colDef.Name]
			var cell string
			if orientation == models.Horizontal {
				cell, _ = excelize.CoordinatesToCellName(startCol+rowIndex, dataStartRow+colIndex)
			} else {
				cell, _ = excelize.CoordinatesToCellName(startCol+colIndex, dataStartRow+rowIndex)
			}
			if err := f.SetCellValue(sheetName, cell, value); err != nil { return err }
			if styleID, ok := styleCache[colDef.CustomFormat]; ok {
				if err := f.SetCellStyle(sheetName, cell, cell, styleID); err != nil { return err }
			}
		}
	}
	return nil
}

func (e *Exporter) createStyleCache(sheet models.Sheet) (map[string]int, error) {
	cache := make(map[string]int)
	for _, block := range sheet.Blocks {
		for _, col := range block.Columns {
			if col.CustomFormat != "" {
				if _, exists := cache[col.CustomFormat]; !exists {
					style, err := e.file.NewStyle(&excelize.Style{CustomNumFmt: &col.CustomFormat})
					if err != nil { return nil, err }
					cache[col.CustomFormat] = style
				}
			}
		}
	}
	return cache, nil
}

func (e *Exporter) applyPostWriteOptions(sheetName string, sheetData models.Sheet) error {
	f := e.file
	for _, block := range sheetData.Blocks {
		startCol, _, _ := getStartCoordinates(block)
		for i, col := range block.Columns {
			if col.Width > 0 {
				colName, _ := excelize.ColumnNumberToName(startCol + i)
				if err := f.SetColWidth(sheetName, colName, colName, col.Width); err != nil { return err }
			}
		}
	}

	for _, formula := range sheetData.Formulas {
		if formula.Column == "" || formula.Expr == "" || formula.StartRow <= 0 { continue }
		if formula.EndRow > 0 {
			for r := formula.StartRow; r <= formula.EndRow; r++ {
				cell := fmt.Sprintf("%s%d", formula.Column, r)
				formulaStr := strings.ReplaceAll(formula.Expr, "{row}", strconv.Itoa(r))
				if err := f.SetCellFormula(sheetName, cell, formulaStr); err != nil { return err }
			}
		}
	}

	if sheetData.Options.Freeze != "" {
		if err := f.SetPanes(sheetName, &excelize.Panes{Freeze: true, TopLeftCell: sheetData.Options.Freeze}); err != nil { return err }
	}
	if sheetData.Options.AutoFilter != "" {
		if err := f.AutoFilter(sheetName, sheetData.Options.AutoFilter, nil); err != nil { return err }
	}
	return nil
}

func getStartCoordinates(block models.Block) (int, int, error) {
	if block.StartCell != "" {
		col, row, err := excelize.CellNameToCoordinates(block.StartCell)
		if err != nil { return 0, 0, err }
		return col, row, nil
	}
	if block.StartRow > 0 && block.StartCol != "" {
		col, err := excelize.ColumnNameToNumber(block.StartCol)
		if err != nil { return 0, 0, err }
		return col, block.StartRow, nil
	}
	return 1, 1, nil
}
