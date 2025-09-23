package cmd

import (
	"fmt"

	"github.com/chelbit/excelms/internal/cli"
	"github.com/spf13/cobra"
)

var (
	importSchemaPath string
	importFilePath   string
)

func init() {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import data from an Excel file based on a schema.",
		Long:  `Reads an Excel file, validates it against a provided JSON schema, and outputs the data as NDJSON.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Executing import command...")
			cli.HandleImport(importSchemaPath, importFilePath)
		},
	}

	importCmd.Flags().StringVar(&importSchemaPath, "schema", "", "Path to the JSON schema file (required)")
	importCmd.Flags().StringVar(&importFilePath, "file", "", "Path to the Excel file to import (required)")
	_ = importCmd.MarkFlagRequired("schema")
	_ = importCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(importCmd)
}
