package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "ensemble",
		Short: "Ensemble - Multi-Agent Coordination Tool",
		Long: `Ensemble is a multi-agent developer tool where a coordinating agent 
dynamically assembles teams of specialized agents from a pool to 
collaboratively accomplish software development tasks.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Ensemble CLI client - Coming soon")
			fmt.Printf("Version: %s\n", version)
			fmt.Println("\nThis is the client component of Ensemble.")
			fmt.Println("Use 'ensemble help' to see available commands.")
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
