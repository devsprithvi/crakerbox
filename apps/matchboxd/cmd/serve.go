package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the matchboxd daemon",
	Long:  `Starts the Matchbox daemon process that manages Firecracker microVMs.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔥 matchboxd daemon starting...")
		fmt.Println("   Listening on :8090")
		fmt.Println("   Press Ctrl+C to stop")
		// TODO: Start actual daemon (gRPC server, VM manager, etc.)
		select {} // block forever for now
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
