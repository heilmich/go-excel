package models

// This file defines the shared structures for API requests and internal processing.

// --- Enums ---

type Orientation string
const (
	Vertical   Orientation = "vertical"
	Horizontal Orientation = "horizontal"
)

type DataType string
const (
	TypeString   DataType = "string"
	TypeInt      DataType = "int"
	TypeFloat    DataType = "float"
	TypeBool     DataType = "bool"
	TypeDate     DataType = "date"
	TypeDateTime DataType = "datetime"
)

// --- Core Data Structures ---

type Column struct {
	Name         string   `json:"name"`
	Header       string   `json:"header,omitempty"`
	Column       string   `json:"column,omitempty"`
	Type         DataType `json:"type"`
	CustomFormat string   `json:"customFormat,omitempty"`
	Width        float64  `json:"width,omitempty"`
	Required     bool     `json:"required,omitempty"`
	MinLen       int      `json:"minLen,omitempty"`
	MaxLen       int      `json:"maxLen,omitempty"`
	Regex        string   `json:"regex,omitempty"`
	Enum         []string `json:"enum,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Trim         bool     `json:"trim"`
	DateFormat   string   `json:"dateFormat,omitempty"`
	Timezone     string   `json:"timezone,omitempty"`
}

// Block is a generic container for data at a specific location.
// It uses the default columns defined on its parent Sheet.
type Block struct {
	StartCell   string                 `json:"startCell,omitempty"`
	StartRow    int                    `json:"startRow,omitempty"`
	StartCol    string                 `json:"startCol,omitempty"`
	Orientation Orientation            `json:"orientation,omitempty"`
	ShowHeaders bool                   `json:"showHeaders"` // Defaults to false for blocks
	Data        []map[string]interface{} `json:"data"`
}

// Table is a self-contained block that has its own column definitions.
type Table struct {
	StartCell   string                 `json:"startCell,omitempty"`
	StartRow    int                    `json:"startRow,omitempty"`
	StartCol    string                 `json:"startCol,omitempty"`
	Orientation Orientation            `json:"orientation,omitempty"`
	ShowHeaders bool                   `json:"showHeaders"` // Defaults to true for tables
	Data        []map[string]interface{} `json:"data"`
	Columns     []Column               `json:"columns"`
}

type Formula struct {
	Column   string `json:"column"`
	StartRow int    `json:"startRow"`
	EndRow   int    `json:"endRow"`
	Expr     string `json:"expr"`
}

type Options struct {
	Freeze     string `json:"freeze,omitempty"`
	AutoFilter string `json:"autofilter,omitempty"`
}

type Sheet struct {
	Name           string    `json:"sheet,omitempty"`
	Index          int       `json:"sheetIndex,omitempty"`
	HeaderRow      int       `json:"headerRow,omitempty"` // For import
	StartRow       int       `json:"startRow,omitempty"`  // For import
	DefaultColumns []Column  `json:"columns,omitempty"` // For Blocks
	Blocks         []Block   `json:"blocks,omitempty"`
	Tables         []Table   `json:"tables,omitempty"`
	Formulas       []Formula `json:"formulas,omitempty"`
	Options        Options   `json:"options,omitempty"`
}

// --- API Request/Response Structures ---

// Request is the main structure for all API requests.
type Request struct {
	Sheets         []Sheet `json:"sheets"`
	TemplateBase64 string  `json:"templateBase64,omitempty"`
}

// ImportResult is the new structure for the entire import response.
type ImportResult struct {
	Data   map[string]map[string][]map[string]interface{} `json:"data"` // SheetName -> BlockIdentifier -> []RowData
	Errors []string                                     `json:"errors,omitempty"`
}
