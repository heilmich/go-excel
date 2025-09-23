package models

// Based on the user request and documentation.

// Orientation defines the direction of data processing in a block.
type Orientation string

const (
	Vertical   Orientation = "vertical"
	Horizontal Orientation = "horizontal"
)

// DataType defines the type of data in a column for validation.
type DataType string

const (
	TypeString   DataType = "string"
	TypeInt      DataType = "int"
	TypeFloat    DataType = "float"
	TypeBool     DataType = "bool"
	TypeDate     DataType = "date"
	TypeDateTime DataType = "datetime"
)

// Column defines a single column or row within a block/table.
type Column struct {
	Name         string   `json:"name"`                   // Key for data mapping, required.
	Header       string   `json:"header,omitempty"`       // Display text in the header row (for import/export).
	Column       string   `json:"column,omitempty"`       // Column letter, e.g. "A" (for import).
	Type         DataType `json:"type"`                   // Data type for validation, required.
	CustomFormat string   `json:"customFormat,omitempty"` // Excel custom format string.
	Width        float64  `json:"width,omitempty"`        // Column width.

	// Validation fields from docs
	Required bool     `json:"required,omitempty"`
	MinLen   int      `json:"minLen,omitempty"`
	MaxLen   int      `json:"maxLen,omitempty"`
	Regex    string   `json:"regex,omitempty"`
	Enum     []string `json:"enum,omitempty"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	Trim     bool     `json:"trim"` // Defaults to true in docs

	// Date-related fields
	DateFormat string `json:"dateFormat,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

// Block defines a rectangular area of cells for export.
type Block struct {
	StartCell   string                 `json:"startCell,omitempty"`   // e.g., "A2"
	StartRow    int                    `json:"startRow,omitempty"`    // e.g., 12 (1-based)
	StartCol    string                 `json:"startCol,omitempty"`    // e.g., "C"
	Orientation Orientation            `json:"orientation,omitempty"` // "vertical" or "horizontal", defaults to vertical.
	ShowHeaders bool                   `json:"showHeaders"`           // If true, headers from the main schema are printed.
	Data        []map[string]interface{} `json:"data"`                  // Array of data objects.
}

// Table is a user-friendly alias for a Block with a defined set of columns.
// The user mentioned "Table" is a wrapper for a block.
type Table struct {
	Block
	Columns []Column `json:"columns"`
}

// Formula defines a formula to be applied to a range of cells.
type Formula struct {
	Column   string `json:"column"`             // e.g., "C"
	StartRow int    `json:"startRow"`           // 1-based
	EndRow   int    `json:"endRow"`             // 1-based, inclusive. 0 means to the end of data.
	Expr     string `json:"expr"`               // Formula expression, e.g., "=B{row}*1.2"
}

// Options defines global options for the export.
type Options struct {
	Freeze     string `json:"freeze,omitempty"`     // e.g., "A2"
	AutoFilter string `json:"autofilter,omitempty"` // e.g., "A1:C100"
}

// ExportRequest is the main structure for an export request.
type ExportRequest struct {
	Sheet          string                 `json:"sheet,omitempty"`     // Target sheet name. If empty, first sheet.
	SheetIndex     int                    `json:"sheetIndex,omitempty"`  // Target sheet index (0-based).
	HeaderRow      int                    `json:"headerRow,omitempty"`   // 1-based, for legacy `rows` block.
	StartRow       int                    `json:"startRow,omitempty"`    // 1-based, for legacy `rows` block.
	Columns        []Column               `json:"columns"`             // Defines the columns for legacy `rows` and default for `blocks`.
	Rows           []map[string]interface{} `json:"rows,omitempty"`      // Legacy continuous data block.
	Blocks         []Block                `json:"blocks,omitempty"`    // Disjoint data blocks.
	Tables         []Table                `json:"tables,omitempty"`    // Tables (blocks with their own columns).
	Formulas       []Formula              `json:"formulas,omitempty"`
	Options        Options                `json:"options,omitempty"`
	TemplateBase64 string                 `json:"templateBase64,omitempty"`
}

// ImportSchema defines the structure for an import operation.
// Based on the docs, import is simpler and row-based.
type ImportSchema struct {
	Sheet       string   `json:"sheet,omitempty"`    // Name of the sheet to import.
	SheetIndex  int      `json:"sheetIndex,omitempty"` // Index of the sheet (0-based).
	HeaderRow   int      `json:"headerRow"`          // Row number of the header (1-based). 0 for no header.
	StartRow    int      `json:"startRow"`           // First row of data (1-based).
	Columns     []Column `json:"columns"`
}

// ImportResultLine represents a single line of output in NDJSON format for import.
type ImportResultLine struct {
	OK       bool              `json:"ok"`
	RowIndex int               `json:"rowIndex"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Errors   []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a single validation error for a field.
type ValidationError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
}
