package cmd

import (
	"context"
	"fmt"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	terminateDaemonAddr string
	terminateCertDir    string
)

var terminateCmd = &cobra.Command{
	Use:   "terminate <id>",
	Short: "Terminate a running Firecracker microVM",
	Long: `Stops a running Firecracker microVM process and deallocates assigned host resources
(TAP interfaces, workspace directory, socket file).

Examples:
  crackerboxd terminate abc123
  crackerboxd terminate abc123 --daemon 10.0.0.5:8090`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		conn, err := dialDaemon(terminateDaemonAddr, terminateCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		resp, err := client.Terminate(context.Background(), &api.TerminateRequest{Id: vmID})
		if err != nil {
			return fmt.Errorf("terminate failed: %w", err)
		}

		if resp.Success {
			fmt.Printf("✅ %s\n", resp.Message)
		} else {
			fmt.Printf("❌ %s\n", resp.Message)
		}

		return nil
	},
}

func init() {
	terminateCmd.Flags().StringVar(&terminateDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	terminateCmd.Flags().StringVar(&terminateCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")

	rootCmd.AddCommand(terminateCmd)
}
