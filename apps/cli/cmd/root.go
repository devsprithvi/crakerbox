package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Set via ldflags at build time
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "crackerbox",
	Short: "Crackerbox CLI — manage Firecracker microVMs",
	Long: `crackerbox is the command-line interface for the Crackerbox platform.

Use it to interact with crackerboxd, manage microVMs,
and configure your Firecracker infrastructure.

Get started:
  crackerbox init         Initialize a new workspace
  crackerbox status       Show cluster status
  crackerbox vm list      List running microVMs`,
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
