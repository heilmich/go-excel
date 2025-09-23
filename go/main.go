package main

import (
	"fmt"
	"os"

	"chelbit/excelms/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// In a real app, you might not want to ignore this error.
	_ = godotenv.Load()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
