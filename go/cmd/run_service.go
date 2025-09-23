package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	var runServiceCmd = &cobra.Command{
		Use:   "run-as-service",
		Short: "Run the application as a service (for internal use by the service manager).",
		Hidden: true, // This hides the command from the help message.
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Run(); err != nil {
				log.Fatalf("Service run failed: %v", err)
			}
		},
	}

	rootCmd.AddCommand(runServiceCmd)
}
