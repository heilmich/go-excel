package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var logger service.Logger

type program struct {
	app *fiber.App
}

func (p *program) Start(s service.Service) error {
	p.app = createServer(os.Getenv("AUTH_TOKEN"))
	go p.run()
	return nil
}

func (p *program) run() {
	portVal, _ := rootCmd.Flags().GetInt("port")
	log.Printf("Service running on port %d", portVal)
	if err := p.app.Listen(fmt.Sprintf(":%d", portVal)); err != nil {
		_ = logger.Error(err)
	}
}

func (p *program) Stop(s service.Service) error {
	log.Println("Stopping service...")
	if err := p.app.Shutdown(); err != nil {
		return err
	}
	return nil
}

var (
	svc service.Service
)

func init() {
	prg := &program{}
	svcConfig := &service.Config{
		Name:        "ExcelMS",
		DisplayName: "Excel Microservice",
		Description: "Service for importing and exporting Excel files.",
		Arguments:   []string{"--serve"}, // When run as a service, always use the --serve flag
	}

	var err error
	svc, err = service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err = svc.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	var serviceCmd = &cobra.Command{
		Use:   "service [action]",
		Short: "Manage the application as a system service (install|uninstall|start|stop|restart).",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			action := args[0]
			err := service.Control(svc, action)
			if err != nil {
				log.Fatalf("Failed to %s service: %v", action, err)
			} else {
				fmt.Printf("Service %s successful.\n", action)
			}
		},
	}

	rootCmd.AddCommand(serviceCmd)
}
