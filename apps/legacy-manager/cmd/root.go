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
	Use:   "crackerbox-manager",
	Short: "Crackerbox Manager — Firecracker Cluster Orchestrator",
	Long: `crackerbox-manager is the cluster orchestration layer for the Crackerbox platform.

It manages multiple crackerboxd nodes, handles workload scheduling,
health monitoring, and provides a unified API for fleet management.

Architecture:
  crackerbox-manager (this) ──► crackerboxd (node 1)
                             ──► crackerboxd (node 2)
                             ──► crackerboxd (node N)

Quick start:
  crackerbox-manager serve       Start the orchestrator
  crackerbox-manager status      Show cluster status
  crackerbox-manager nodes       List registered nodes
  crackerbox-manager version     Print version information
  crackerbox-manager update      Self-update to latest release`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/crackerbox/manager.yaml", "Config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
