package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

Get started:
  crackerbox-manager serve       Start the orchestrator
  crackerbox-manager status      Show cluster status
  crackerbox-manager nodes       List registered nodes`,
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
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
