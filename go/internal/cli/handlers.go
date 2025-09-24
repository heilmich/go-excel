package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"chelbit/excelms/internal/excel"
	"chelbit/excelms/internal/models"
)

// HandleImport is the CLI handler for the import command.
func HandleImport(schemaPath, filePath string) {
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		log.Fatalf("Failed to open schema file: %v", err)
	}
	defer schemaFile.Close()

	var request models.Request
	if err := json.NewDecoder(schemaFile).Decode(&request); err != nil {
		log.Fatalf("Failed to parse schema JSON: %v", err)
	}

	excelFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open Excel file: %v", err)
	}
	defer excelFile.Close()

	fmt.Fprintf(os.Stderr, "Importing data...\n")
	importer := excel.NewImporter(excelFile, &request)
	result, err := importer.Process()
	if err != nil {
		log.Fatalf("Error during import: %v", err)
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize result to JSON: %v", err)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "Import complete.\n")
}

// HandleExport is the CLI handler for the export command.
func HandleExport(requestPath, outputPath, templatePath string) {
	requestBytes, err := ioutil.ReadFile(requestPath)
	if err != nil {
		log.Fatalf("Failed to read request file: %v", err)
	}

	var request models.Request
	if err := json.Unmarshal(requestBytes, &request); err != nil {
		log.Fatalf("Failed to parse request JSON: %v", err)
	}

	var templateBytes []byte
	if templatePath != "" {
		templateBytes, err = ioutil.ReadFile(templatePath)
		if err != nil {
			log.Fatalf("Failed to read template file: %v", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Generating Excel file...\n")
	exporter := excel.NewExporter(&request, templateBytes)
	resultBytes, err := exporter.Process()
	if err != nil {
		log.Fatalf("Failed to generate Excel file: %v", err)
	}

	if err := ioutil.WriteFile(outputPath, resultBytes, 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Successfully saved export to %s\n", outputPath)
}
