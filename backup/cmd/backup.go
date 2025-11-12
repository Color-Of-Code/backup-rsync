package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Add the run and simulate verbs with empty implementations
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the backup jobs",
	Run: func(cmd *cobra.Command, args []string) {
		// Empty implementation for now
		fmt.Println("Run command executed.")
	},
}

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate the backup jobs",
	Run: func(cmd *cobra.Command, args []string) {
		// Empty implementation for now
		fmt.Println("Simulate command executed.")
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(simulateCmd)
}
