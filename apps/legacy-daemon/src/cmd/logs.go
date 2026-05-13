package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	logsDaemonAddr string
	logsCertDir    string
	logsFollow     bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Stream the process log file for a Firecracker VM",
	Long: `Accesses and streams the internal log file for a specific Firecracker microVM.

Examples:
  crackerboxd logs abc123
  crackerboxd logs abc123 --follow`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		conn, err := dialDaemon(logsDaemonAddr, logsCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		stream, err := client.Logs(context.Background(), &api.LogsRequest{
			Id:     vmID,
			Follow: logsFollow,
		})
		if err != nil {
			return fmt.Errorf("logs failed: %w", err)
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("stream error: %w", err)
			}

			os.Stdout.Write(resp.Data)
		}
	},
}

func init() {
	logsCmd.Flags().StringVar(&logsDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	logsCmd.Flags().StringVar(&logsCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (like tail -f)")

	rootCmd.AddCommand(logsCmd)
}
