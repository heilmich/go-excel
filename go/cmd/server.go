package cmd

import (
	"fmt"
	"log"

	"chelbit/excelms/internal/api"
	"github.com/gofiber/fiber/v2"
)

// runServer initializes and starts the Fiber web server.
// This is the single entry point for running the server logic.
func runServer(port int, authToken string) {
	fmt.Printf("Starting server on port %d...\n", port)
	if authToken == "" {
		fmt.Println("AUTH_TOKEN not set, running in insecure mode.")
	}

	app := api.NewServer(authToken)

	// The Listen function blocks, so we run it.
	// Graceful shutdown will be handled by the service manager.
	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createServer just creates the app without running Listen.
// This is needed for the service so we can get a reference to the app for shutdown.
func createServer(authToken string) *fiber.App {
	if authToken == "" {
		log.Println("AUTH_TOKEN not set, running in insecure mode.")
	}
	app := api.NewServer(authToken)
	return app
}
