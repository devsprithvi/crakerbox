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
It manages Firecracker microVM lifecycles on a single host, exposing
a mTLS-secured gRPC API for the cluster manager to connect to.

Process Management:
  crackerboxd serve                        Start the daemon (gRPC + mTLS)
  crackerboxd spawn --config <path>        Spawn a new Firecracker VM
  crackerboxd terminate <id>               Terminate a running VM
  crackerboxd list                         List all managed VMs
  crackerboxd inspect <id>                 Show host-level VM details

Telemetry:
  crackerboxd logs <id>                    Stream a VM's log file
  crackerboxd metrics <id>                 Read a VM's metrics

Bridge:
  crackerboxd call <id> <method> <path>    Forward a request to the Firecracker API socket

Certificate Management:
  crackerboxd certs generate               Generate mTLS PKI bundle
  crackerboxd certs info                   Show certificate status

Utilities:
  crackerboxd status                       Show daemon & system status
  crackerboxd version                      Print version information
  crackerboxd update                       Self-update to latest release`,

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
