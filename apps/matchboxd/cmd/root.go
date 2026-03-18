package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "crackerboxd",
	Short: "Crackerbox Daemon — Firecracker VM Control Plane",
	Long: `crackerboxd is the daemon binary for the Crackerbox platform.
It manages Firecracker microVM lifecycles, handles orchestration,
and provides both a CLI and GUI interface for server management.

Quick start:
  crackerboxd serve          Start the node daemon
  crackerboxd status         Show daemon & system status
  crackerboxd version        Print version information
  crackerboxd update         Self-update to latest release`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/crackerbox/crackerboxd.yaml", "Config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
