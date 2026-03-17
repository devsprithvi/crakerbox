package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Set via ldflags at build time
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of crackerbox-manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("crackerbox-manager %s (commit: %s, built: %s)\n", Version, GitCommit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
