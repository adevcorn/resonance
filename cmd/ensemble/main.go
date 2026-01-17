package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	configPath string

	rootCmd = &cobra.Command{
		Use:   "ensemble",
		Short: "Ensemble - Multi-Agent Coordination Tool",
		Long: `Ensemble is a multi-agent developer tool where a coordinating agent 
dynamically assembles teams of specialized agents from a pool to 
collaboratively accomplish software development tasks.`,
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(agentsCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ensemble CLI version %s\n", version)
	},
}
