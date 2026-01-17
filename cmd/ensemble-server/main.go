package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "ensemble-server",
		Short: "Ensemble Server - Multi-Agent Coordination Backend",
		Long: `Ensemble Server is the backend component of the Ensemble multi-agent 
coordination tool. It manages agent orchestration, LLM provider communication, 
and session management.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Ensemble Server - Coming soon")
			fmt.Printf("Version: %s\n", version)
			fmt.Println("\nThis is the server component of Ensemble.")
			fmt.Println("Use 'ensemble-server help' to see available commands.")
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
