package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/chelbit/excelms/internal/api"
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
			startServer()
		} else {
			// If no subcommand is given and not in serve mode, show help.
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
	fmt.Printf("Starting server on port %d...\n", port)

	// Get auth token from env
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		fmt.Println("AUTH_TOKEN not set, running in insecure mode.")
	}

	// Initialize and run the Fiber app
	app := api.NewServer(authToken)

	// The Listen function blocks until the server is stopped.
	// We need to handle the potential error it returns.
	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
