package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	serve     bool
	port      int
)

var rootCmd = &cobra.Command{
	Use:   "excelms",
	Short: "ExcelMS is a microservice for importing and exporting Excel files.",
	Long:  `A versatile tool that can run as a CLI or as a web service to handle complex Excel import and export tasks based on a JSON schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This logic only runs if no subcommand is specified.
		if serve {
			// Run the server interactively in the foreground.
			startServer()
		} else {
			// Default behavior: show help.
			_ = cmd.Help()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&serve, "serve", "s", false, "Run server interactively in the foreground")

	defaultPort := 8080
	if envPort, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		defaultPort = envPort
	}
	// Port needs to be persistent so other commands (like service) can see it.
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", defaultPort, "Port to run the server on")
}

func Execute() error {
	return rootCmd.Execute()
}

func startServer() {
	authToken := os.Getenv("AUTH_TOKEN")
	runServer(port, authToken)
}
