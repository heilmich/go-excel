package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	serve      bool
	port       int
	importCmd  *cobra.Command
	exportCmd  *cobra.Command
)

var rootCmd = &cobra.Command{
	Use:   "excelms",
	Short: "ExcelMS is a microservice for importing and exporting Excel files.",
	Long:  `A versatile tool that can run as a CLI or as a web service to handle complex Excel import and export tasks based on a JSON schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		if serve {
			// This will block and run the service.
			// The service library handles whether it's running interactively or as a managed service.
			if err := svc.Run(); err != nil {
				log.Fatal(err)
			}
		} else {
			// If not in serve mode and no command is given, show help.
			if err := cmd.Help(); err != nil {
				log.Fatalf("Failed to show help: %v", err)
			}
			os.Exit(0)
		}
	},
}

func init() {
	// Persistent flags, available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&serve, "serve", "s", false, "Run in server mode")

	// Get port from env, with a default
	defaultPort := 8080
	if envPort, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		defaultPort = envPort
	}
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", defaultPort, "Port to run the server on")
}

func Execute() error {
	return rootCmd.Execute()
}

func startServer() {
	authToken := os.Getenv("AUTH_TOKEN")
	runServer(port, authToken)
}
