package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "matchboxd",
	Short: "Matchbox Daemon — Firecracker VM Control Plane",
	Long: `matchboxd is the daemon binary for the Matchbox platform.
It manages Firecracker microVM lifecycles, handles orchestration,
and provides both a CLI and GUI interface for server management.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
