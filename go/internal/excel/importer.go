package excel

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"chelbit/excelms/internal/models"
	"github.com/xuri/excelize/v2"
)

// Importer handles the logic of importing data from an Excel file.
type Importer struct {
	fileReader   io.Reader
	schema       models.ImportSchema
	outputWriter io.Writer
	file         *excelize.File
}

// NewImporter creates a new instance of an Importer.
func NewImporter(reader io.Reader, schema models.ImportSchema, writer io.Writer) *Importer {
	return &Importer{
		fileReader:   reader,
		schema:       schema,
		outputWriter: writer,
	}
}

// Process executes the import operation.
func (i *Importer) Process() error {
	var err error
	i.file, err = excelize.OpenReader(i.fileReader)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %w", err)
	}
	defer i.file.Close()

	sheetName, err := i.getTargetSheet()
	if err != nil {
		return err
	}

	colMapping, err := i.buildColumnMapping(sheetName)
	if err != nil {
		return fmt.Errorf("failed to build column mapping: %w", err)
	}

	rows, err := i.file.Rows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows iterator for sheet '%s': %w", sheetName, err)
	}

	encoder := json.NewEncoder(i.outputWriter)
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		if rowIndex < i.schema.StartRow {
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

// ### Helper methods for Importer ###

func (i *Importer) getTargetSheet() (string, error) {
	// This is a simplified version for the importer's context.
	// It doesn't create sheets.
	f := i.file
	name := i.schema.Sheet
	index := i.schema.SheetIndex

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

func (i *Importer) buildColumnMapping(sheet string) (map[int]models.Column, error) {
	mapping := make(map[int]models.Column)
	mappedCols := make(map[string]bool)

	// PERFORMANCE FIX: Use iterator instead of GetRows()
	if i.schema.HeaderRow > 0 {
		rows, err := i.file.Rows(sheet)
		if err != nil {
			return nil, fmt.Errorf("could not open rows iterator for mapping: %w", err)
		}

		currentRow := 0
		var headerCells []string
		for rows.Next() {
			currentRow++
			if currentRow == i.schema.HeaderRow {
				headerCells, err = rows.Columns()
				if err != nil {
					return nil, fmt.Errorf("could not read header row %d: %w", i.schema.HeaderRow, err)
				}
				break
			}
		}
		rows.Close() // Close the iterator, a new one will be opened for processing data.

		if len(headerCells) == 0 {
			return nil, fmt.Errorf("headerRow %d not found or is empty", i.schema.HeaderRow)
		}

		for schemaColIdx, schemaCol := range i.schema.Columns {
			if schemaCol.Header == "" { continue }
			for cellIdx, cellVal := range headerCells {
				if strings.TrimSpace(cellVal) == strings.TrimSpace(schemaCol.Header) {
					mapping[cellIdx+1] = i.schema.Columns[schemaColIdx]
					mappedCols[schemaCol.Name] = true
					break
				}
			}
		}
	}

	// Map remaining columns by explicit column letter
	for schemaColIdx, schemaCol := range i.schema.Columns {
		if _, isMapped := mappedCols[schemaCol.Name]; isMapped { continue }
		if schemaCol.Column != "" {
			colNum, err := excelize.ColumnNameToNumber(schemaCol.Column)
			if err != nil {
				return nil, fmt.Errorf("invalid column letter '%s' for field '%s'", schemaCol.Column, schemaCol.Name)
			}
			mapping[colNum] = i.schema.Columns[schemaColIdx]
		}
	}

	if len(mapping) == 0 {
		return nil, fmt.Errorf("could not map any columns from the schema")
	}
	return mapping, nil
}

func validateAndConvert(value string, col models.Column) (interface{}, []models.ValidationError) {
	errors := make([]models.ValidationError, 0)
	if col.Trim { value = strings.TrimSpace(value) }

	if col.Required && value == "" {
		errors = append(errors, models.ValidationError{Field: col.Name, Code: "required", Msg: "Value is required"})
		return nil, errors
	}
	if value == "" { return nil, nil }

	if col.MinLen > 0 && len(value) < col.MinLen { errors = append(errors, models.ValidationError{Field: col.Name, Code: "minLen", Msg: fmt.Sprintf("Length must be at least %d", col.MinLen)}) }
	if col.MaxLen > 0 && len(value) > col.MaxLen { errors = append(errors, models.ValidationError{Field: col.Name, Code: "maxLen", Msg: fmt.Sprintf("Length must be at most %d", col.MaxLen)}) }
	if col.Regex != "" {
		if matched, _ := regexp.MatchString(col.Regex, value); !matched {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "regex", Msg: "Value does not match pattern"})
		}
	}
	if len(col.Enum) > 0 {
		inEnum := false
		for _, enumVal := range col.Enum {
			if value == enumVal { inEnum = true; break }
		}
		if !inEnum { errors = append(errors, models.ValidationError{Field: col.Name, Code: "enum", Msg: "Value is not in the allowed list"}) }
	}

	var convertedValue interface{}
	switch col.Type {
	case models.TypeString:
		convertedValue = value
	case models.TypeInt:
		if f, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64); err == nil {
			i := int64(f)
			convertedValue = i
			if col.Min != nil && float64(i) < *col.Min { errors = append(errors, models.ValidationError{Field: col.Name, Code: "min", Msg: fmt.Sprintf("Value must be at least %v", *col.Min)}) }
			if col.Max != nil && float64(i) > *col.Max { errors = append(errors, models.ValidationError{Field: col.Name, Code: "max", Msg: fmt.Sprintf("Value must be at most %v", *col.Max)}) }
		} else {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid integer format"})
		}
	case models.TypeFloat:
		if f, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64); err == nil {
			convertedValue = f
			if col.Min != nil && f < *col.Min { errors = append(errors, models.ValidationError{Field: col.Name, Code: "min", Msg: fmt.Sprintf("Value must be at least %v", *col.Min)}) }
			if col.Max != nil && f > *col.Max { errors = append(errors, models.ValidationError{Field: col.Name, Code: "max", Msg: fmt.Sprintf("Value must be at most %v", *col.Max)}) }
		} else {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid float format"})
		}
	case models.TypeBool:
		lowerVal := strings.ToLower(value)
		switch lowerVal {
		case "1", "true", "yes", "y", "да": convertedValue = true
		case "0", "false", "no", "n", "нет": convertedValue = false
		default: errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid boolean format"})
		}
	case models.TypeDate, models.TypeDateTime:
		layouts := []string{ time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "02.01.2006", "01/02/2006", "2006-01-02 15:04:05" }
		var parsedTime time.Time
		var parseErr error
		for _, layout := range layouts {
			parsedTime, parseErr = time.Parse(layout, value)
			if parseErr == nil { break }
		}
		if parseErr != nil {
			errors = append(errors, models.ValidationError{Field: col.Name, Code: "type", Msg: "Invalid date/datetime format"})
		} else {
			outputFormat := col.DateFormat
			if outputFormat == "" {
				if col.Type == models.TypeDate { outputFormat = "2006-01-02" } else { outputFormat = time.RFC3339 }
			}
			convertedValue = parsedTime.Format(outputFormat)
		}
	}
	return convertedValue, errors
}
