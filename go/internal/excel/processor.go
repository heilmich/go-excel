package excel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

// #############################################################################
// # EXPORT LOGIC
// #############################################################################

func ExportProcessor(request models.ExportRequest, templateBytes []byte) ([]byte, error) {
	var f *excelize.File
	var err error

	if len(templateBytes) > 0 {
		f, err = excelize.OpenReader(bytes.NewReader(templateBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to open template: %w", err)
		}
	} else {
		f = excelize.NewFile()
	}
	defer f.Close()

	sheetName, err := getTargetSheet(f, request.Sheet, request.SheetIndex)
	if err != nil {
		return nil, err
	}

	styleCache, err := createStyleCache(f, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create style cache: %w", err)
	}

	if len(request.Rows) > 0 {
		headerRow := request.HeaderRow
		if headerRow == 0 {
			headerRow = 1
		}
		dataStartRow := request.StartRow
		if dataStartRow == 0 {
			dataStartRow = headerRow + 1
		}
		err = writeBlock(f, sheetName, request.Columns, request.Rows, 1, dataStartRow, headerRow, true, models.Vertical, styleCache)
		if err != nil {
			return nil, fmt.Errorf("failed to write legacy rows block: %w", err)
		}
	}

	for _, block := range request.Blocks {
		startCol, startRow, err := getStartCoordinates(block)
		if err != nil {
			return nil, fmt.Errorf("invalid block coordinates: %w", err)
		}
		headerRow := startRow - 1
		if headerRow < 1 {
			headerRow = 1
		}
		orientation := block.Orientation
		if orientation == "" {
			orientation = models.Vertical
		}
		err = writeBlock(f, sheetName, request.Columns, block.Data, startCol, startRow, headerRow, block.ShowHeaders, orientation, styleCache)
		if err != nil {
			return nil, fmt.Errorf("failed to write block starting at %s: %w", block.StartCell, err)
		}
	}

	for _, table := range request.Tables {
		startCol, startRow, err := getStartCoordinates(table.Block)
		if err != nil {
			return nil, fmt.Errorf("invalid table coordinates: %w", err)
		}
		headerRow := startRow - 1
		if headerRow < 1 {
			headerRow = 1
		}
		orientation := table.Orientation
		if orientation == "" {
			orientation = models.Vertical
		}
		err = writeBlock(f, sheetName, table.Columns, table.Data, startCol, startRow, headerRow, table.ShowHeaders, orientation, styleCache)
		if err != nil {
			return nil, fmt.Errorf("failed to write table starting at %s: %w", table.StartCell, err)
		}
	}

	if err := applyPostWriteOptions(f, sheetName, request); err != nil {
		return nil, fmt.Errorf("failed to apply post-write options: %w", err)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write to buffer: %w", err)
	}
	return buf.Bytes(), nil
}

// #############################################################################
// # IMPORT LOGIC
// #############################################################################

func ImportProcessor(fileReader io.Reader, schema models.ImportSchema, outputWriter io.Writer) error {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheetName, err := getTargetSheet(f, schema.Sheet, schema.SheetIndex)
	if err != nil {
		return err
	}

	colMapping, err := buildColumnMapping(f, sheetName, schema)
	if err != nil {
		return fmt.Errorf("failed to build column mapping: %w", err)
	}

	rows, err := f.Rows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows iterator for sheet '%s': %w", sheetName, err)
	}

	encoder := json.NewEncoder(outputWriter)
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		if rowIndex < schema.StartRow {
			continue
		}

		rowCells, err := rows.Columns()
		if err != nil {
			_ = encoder.Encode(models.ImportResultLine{OK: false, RowIndex: rowIndex, Errors: []models.ValidationError{{Msg: "Failed to read row cells"}}})
			continue
		}

		resultLine := models.ImportResultLine{
			RowIndex: rowIndex,
			Data:     make(map[string]interface{}),
			Errors:   make([]models.ValidationError, 0),
		}

		for colIdx, colDef := range colMapping {
			var rawValue string
			if colIdx-1 < len(rowCells) {
				rawValue = rowCells[colIdx-1]
			}
			convertedValue, validationErrors := validateAndConvert(rawValue, colDef)
			if len(validationErrors) > 0 {
				resultLine.Errors = append(resultLine.Errors, validationErrors...)
			}
			if convertedValue != nil {
				resultLine.Data[colDef.Name] = convertedValue
			}
		}

		resultLine.OK = len(resultLine.Errors) == 0
		if err := encoder.Encode(resultLine); err != nil {
			return fmt.Errorf("failed to write ndjson line for row %d: %w", rowIndex, err)
		}
	}
	return rows.Error()
}

// #############################################################################
// # HELPER FUNCTIONS
// #############################################################################

func getTargetSheet(f *excelize.File, name string, index int) (string, error) {
	// Priority 1: Use sheet name if provided
	if name != "" {
		sheetIndex, err := f.GetSheetIndex(name)
		// If sheet doesn't exist, create it
		if err != nil {
			newIndex, err := f.NewSheet(name)
			if err != nil {
				return "", fmt.Errorf("failed to create sheet '%s': %w", name, err)
			}
			f.SetActiveSheet(newIndex)

			// If it's a new file, the default "Sheet1" might need to be deleted.
			if len(f.GetSheetList()) > 1 {
				if _, err := f.GetSheetIndex("Sheet1"); err == nil && name != "Sheet1" {
					rows, err := f.GetRows("Sheet1")
					if err == nil && len(rows) == 0 {
						_ = f.DeleteSheet("Sheet1") // ignore error
					}
				}
			}
			return name, nil
		}
		// If sheet exists, set it as active
		f.SetActiveSheet(sheetIndex)
		return name, nil
	}

	// Priority 2: Use sheet index if provided
	// Note: excelize GetSheetName is 0-indexed. User's request might imply 1-based for sheetIndex.
	// We'll assume 0-based for consistency with the library.
	if index > 0 {
		sheetList := f.GetSheetList()
		if index < len(sheetList) {
			return sheetList[index], nil
		}
		return "", fmt.Errorf("sheet index %d is out of bounds", index)
	}

	// Priority 3: Use the default active sheet
	if len(f.GetSheetList()) > 0 {
		return f.GetSheetName(f.GetActiveSheetIndex()), nil
	}

	// Fallback for an entirely empty file (shouldn't happen with NewFile)
	return "Sheet1", nil
}

func getStartCoordinates(block models.Block) (int, int, error) {
	if block.StartCell != "" {
		col, row, err := excelize.CellNameToCoordinates(block.StartCell)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid startCell '%s': %w", block.StartCell, err)
		}
		return col, row, nil
	}
	if block.StartRow > 0 && block.StartCol != "" {
		col, err := excelize.ColumnNameToNumber(block.StartCol)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid startCol '%s': %w", block.StartCol, err)
		}
		return col, block.StartRow, nil
	}
	return 1, 1, nil
}

func writeBlock(f *excelize.File, sheet string, cols []models.Column, data []map[string]interface{}, startCol, dataStartRow, headerRow int, showHeaders bool, orientation models.Orientation, styleCache map[string]int) error {
	if showHeaders {
		for i, colDef := range cols {
			var cell string
			var err error
			if orientation == models.Horizontal {
				cell, err = excelize.CoordinatesToCellName(startCol, headerRow+i)
			} else {
				cell, err = excelize.CoordinatesToCellName(startCol+i, headerRow)
			}
			if err != nil {
				return fmt.Errorf("failed to get header cell for col %d: %w", i, err)
			}
			if err := f.SetCellValue(sheet, cell, colDef.Header); err != nil {
				return fmt.Errorf("failed to set header value for cell %s: %w", cell, err)
			}
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
			if err != nil {
				return fmt.Errorf("failed to get data cell for row %d, col %d: %w", rowIndex, colIndex, err)
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return fmt.Errorf("failed to set data value for cell %s: %w", cell, err)
			}
			if styleID, ok := styleCache[colDef.CustomFormat]; ok && colDef.CustomFormat != "" {
				if err := f.SetCellStyle(sheet, cell, cell, styleID); err != nil {
					return fmt.Errorf("failed to set style for cell %s: %w", cell, err)
				}
			}
		}
	}
	return nil
}

func createStyleCache(f *excelize.File, request models.ExportRequest) (map[string]int, error) {
	cache := make(map[string]int)
	allCols := make([]models.Column, 0)
	allCols = append(allCols, request.Columns...)
	for _, table := range request.Tables {
		allCols = append(allCols, table.Columns...)
	}
	for _, col := range allCols {
		if col.CustomFormat != "" {
			if _, exists := cache[col.CustomFormat]; !exists {
				style, err := f.NewStyle(&excelize.Style{CustomNumFmt: &col.CustomFormat})
				if err != nil {
					return nil, fmt.Errorf("failed to create style for format '%s': %w", col.CustomFormat, err)
				}
				cache[col.CustomFormat] = style
			}
		}
	}
	return cache, nil
}

func applyPostWriteOptions(f *excelize.File, sheet string, request models.ExportRequest) error {
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
			if err := f.SetColWidth(sheet, colName, colName, col.Width); err != nil {
				return fmt.Errorf("failed to set width for column %s: %w", colName, err)
			}
		}
	}

	for _, formula := range request.Formulas {
		if formula.Column == "" || formula.Expr == "" || formula.StartRow <= 0 {
			continue
		}
		if formula.EndRow > 0 {
			for r := formula.StartRow; r <= formula.EndRow; r++ {
				cell := fmt.Sprintf("%s%d", formula.Column, r)
				formulaStr := strings.ReplaceAll(formula.Expr, "{row}", strconv.Itoa(r))
				if err := f.SetCellFormula(sheet, cell, formulaStr); err != nil {
					return fmt.Errorf("failed to set formula for cell %s: %w", cell, err)
				}
			}
		}
	}

	if request.Options.Freeze != "" {
		if err := f.SetPanes(sheet, &excelize.Panes{Freeze: true, TopLeftCell: request.Options.Freeze}); err != nil {
			return fmt.Errorf("failed to set freeze panes at %s: %w", request.Options.Freeze, err)
		}
	}
	if request.Options.AutoFilter != "" {
		if err := f.AutoFilter(sheet, request.Options.AutoFilter, nil); err != nil {
			return fmt.Errorf("failed to set autofilter on range %s: %w", request.Options.AutoFilter, err)
		}
	}
	return nil
}

func buildColumnMapping(f *excelize.File, sheet string, schema models.ImportSchema) (map[int]models.Column, error) {
	mapping := make(map[int]models.Column)
	mappedCols := make(map[string]bool)
	if schema.HeaderRow > 0 {
		rows, err := f.GetRows(sheet)
		if err != nil || schema.HeaderRow > len(rows) {
			return nil, fmt.Errorf("headerRow %d is out of bounds or sheet is unreadable", schema.HeaderRow)
		}
		headerCells := rows[schema.HeaderRow-1]
		for schemaColIdx, schemaCol := range schema.Columns {
			if schemaCol.Header == "" {
				continue
			}
			for cellIdx, cellVal := range headerCells {
				if strings.TrimSpace(cellVal) == strings.TrimSpace(schemaCol.Header) {
					mapping[cellIdx+1] = schema.Columns[schemaColIdx]
					mappedCols[schemaCol.Name] = true
					break
				}
			}
		}
	}

	for schemaColIdx, schemaCol := range schema.Columns {
		if _, isMapped := mappedCols[schemaCol.Name]; isMapped {
			continue
		}
		if schemaCol.Column != "" {
			colNum, err := excelize.ColumnNameToNumber(schemaCol.Column)
			if err != nil {
				return nil, fmt.Errorf("invalid column letter '%s' for field '%s'", schemaCol.Column, schemaCol.Name)
			}
			mapping[colNum] = schema.Columns[schemaColIdx]
		}
	}
	if len(mapping) == 0 {
		return nil, fmt.Errorf("could not map any columns from the schema; check headers or column letters")
	}
	return mapping, nil
}

func validateAndConvert(value string, col models.Column) (interface{}, []models.ValidationError) {
	errors := make([]models.ValidationError, 0)
	if col.Trim {
		value = strings.TrimSpace(value)
	}
	if col.Required && value == "" {
		errors = append(errors, models.ValidationError{Field: col.Name, Code: "required", Msg: "Value is required"})
		return nil, errors
	}
	if value == "" {
		return nil, nil
	}

	if col.MinLen > 0 && len(value) < col.MinLen {
		errors = append(errors, models.ValidationError{Field: col.Name, Code: "minLen", Msg: fmt.Sprintf("Length must be at least %d", col.MinLen)})
	}
	if col.MaxLen > 0 && len(value) > col.MaxLen {
		errors = append(errors, models.ValidationError{Field: col.Name, Code: "maxLen", Msg: fmt.Sprintf("Length must be at most %d", col.MaxLen)})
	}
	if col.Regex != "" {
		if matched, _ := regexp.MatchString(col.Regex, value); !matched {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "regex", Msg: "Value does not match pattern"})
		}
	}
	if len(col.Enum) > 0 {
		inEnum := false
		for _, enumVal := range col.Enum {
			if value == enumVal {
				inEnum = true
				break
			}
		}
		if !inEnum {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "enum", Msg: "Value is not in the allowed list"})
		}
	}

	var convertedValue interface{}
	switch col.Type {
	case models.TypeString:
		convertedValue = value
	case models.TypeInt:
		if f, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64); err == nil {
			i := int64(f)
			convertedValue = i
			if col.Min != nil && float64(i) < *col.Min {
				errors = append(errors, models.ValidationError{Field: col.Name, Code: "min", Msg: fmt.Sprintf("Value must be at least %v", *col.Min)})
			}
			if col.Max != nil && float64(i) > *col.Max {
				errors = append(errors, models.ValidationError{Field: col.Name, Code: "max", Msg: fmt.Sprintf("Value must be at most %v", *col.Max)})
			}
		} else {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid integer format"})
		}
	case models.TypeFloat:
		if f, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64); err == nil {
			convertedValue = f
			if col.Min != nil && f < *col.Min {
				errors = append(errors, models.ValidationError{Field: col.Name, Code: "min", Msg: fmt.Sprintf("Value must be at least %v", *col.Min)})
			}
			if col.Max != nil && f > *col.Max {
				errors = append(errors, models.ValidationError{Field: col.Name, Code: "max", Msg: fmt.Sprintf("Value must be at most %v", *col.Max)})
			}
		} else {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid float format"})
		}
	case models.TypeBool:
		lowerVal := strings.ToLower(value)
		switch lowerVal {
		case "1", "true", "yes", "y", "да":
			convertedValue = true
		case "0", "false", "no", "n", "нет":
			convertedValue = false
		default:
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid boolean format"})
		}
	case models.TypeDate, models.TypeDateTime:
		layouts := []string{
			time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "02.01.2006", "01/02/2006", "2006-01-02 15:04:05",
		}
		var parsedTime time.Time
		var parseErr error
		for _, layout := range layouts {
			parsedTime, parseErr = time.Parse(layout, value)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid date/datetime format"})
		} else {
			outputFormat := col.DateFormat
			if outputFormat == "" {
				if col.Type == models.TypeDate {
					outputFormat = "2006-01-02"
				} else {
					outputFormat = time.RFC3339
				}
			}
			convertedValue = parsedTime.Format(outputFormat)
		}
	}
	return convertedValue, errors
}
