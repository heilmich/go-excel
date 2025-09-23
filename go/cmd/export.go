package cmd

import (
	"fmt"

	"chelbit/excelms/internal/cli"
	"github.com/spf13/cobra"
)

var (
	requestPath  string
	outputPath   string
	templatePath string
)

func init() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export data to an Excel file based on a JSON request.",
		Long:  `Generates an Excel file from a JSON request file that defines the structure, data, layout, and formatting.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Executing export command...")
			cli.HandleExport(requestPath, outputPath, templatePath)
		},
	}

	exportCmd.Flags().StringVar(&requestPath, "request", "", "Path to the JSON request file (required)")
	exportCmd.Flags().StringVar(&outputPath, "output", "", "Path to save the generated Excel file (required)")
	exportCmd.Flags().StringVar(&templatePath, "template", "", "Optional path to an Excel template file")

	_ = exportCmd.MarkFlagRequired("request")
	_ = exportCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(exportCmd)
}
