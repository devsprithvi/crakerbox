package cmd

import (
	"context"
	"fmt"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	metricsDaemonAddr string
	metricsCertDir    string
)

var metricsCmd = &cobra.Command{
	Use:   "metrics <id>",
	Short: "Read the latest process metrics for a Firecracker VM",
	Long: `Accesses and reads the associated metrics file from the host filesystem
and returns a point-in-time snapshot of process performance data in JSON format.

Examples:
  crackerboxd metrics abc123`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		conn, err := dialDaemon(metricsDaemonAddr, metricsCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		resp, err := client.Metrics(context.Background(), &api.MetricsRequest{Id: vmID})
		if err != nil {
			return fmt.Errorf("metrics failed: %w", err)
		}

		fmt.Println(string(resp.Data))
		return nil
	},
}

func init() {
	metricsCmd.Flags().StringVar(&metricsDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	metricsCmd.Flags().StringVar(&metricsCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")

	rootCmd.AddCommand(metricsCmd)
}
