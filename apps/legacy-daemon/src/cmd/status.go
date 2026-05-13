package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	statusDaemonAddr string
	statusCertDir    string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon and system status",
	Long:  `Displays the current status of the crackerboxd daemon and system capabilities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Println("  🔥 Crackerbox Daemon Status")
		fmt.Println("  ════════════════════════════════════")

		// System info
		fmt.Printf("  %-20s %s\n", "Version:", Version)
		fmt.Printf("  %-20s %s/%s\n", "Platform:", runtime.GOOS, runtime.GOARCH)

		// KVM check
		kvmStatus := "❌ Not available"
		if runtime.GOOS == "linux" {
			if _, err := os.Stat("/dev/kvm"); err == nil {
				kvmStatus = "✅ Available"
			}
		} else {
			kvmStatus = "⚠  Not on Linux"
		}
		fmt.Printf("  %-20s %s\n", "KVM:", kvmStatus)

		// Certs check
		certsStatus := "❌ Not found"
		if pki.CertsExist(statusCertDir) {
			certsStatus = "✅ Present"
		}
		fmt.Printf("  %-20s %s (%s)\n", "Certificates:", certsStatus, statusCertDir)

		// Try to connect to running daemon
		fmt.Printf("  %-20s ", "Daemon:")
		conn, err := dialDaemon(statusDaemonAddr, statusCertDir)
		if err != nil {
			fmt.Println("● Not reachable")
		} else {
			defer conn.Close()
			client := api.NewDaemonServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			resp, err := client.Health(ctx, &api.HealthRequest{})
			if err != nil {
				fmt.Println("● Not responding")
			} else {
				fmt.Printf("✅ Online — %d VM(s), uptime %s\n", resp.VmCount, resp.Uptime)
			}
		}

		fmt.Println("  ════════════════════════════════════")
		fmt.Println()
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	statusCmd.Flags().StringVar(&statusCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")

	rootCmd.AddCommand(statusCmd)
}
