package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	inspectDaemonAddr string
	inspectCertDir    string
	inspectJSON       bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Show host-level details for a Firecracker process",
	Long: `Aggregates and displays host-level metadata for a specific Firecracker microVM,
including PID, configuration sources, file paths, uptime, and process status.

Examples:
  crackerboxd inspect abc123
  crackerboxd inspect abc123 --json`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		conn, err := dialDaemon(inspectDaemonAddr, inspectCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		resp, err := client.Inspect(context.Background(), &api.InspectRequest{Id: vmID})
		if err != nil {
			return fmt.Errorf("inspect failed: %w", err)
		}

		if inspectJSON {
			out, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		// Human-readable output
		fmt.Printf("🔥 VM %s (%s)\n", resp.Id, resp.Name)
		fmt.Println("════════════════════════════════════════")

		// Process
		aliveStr := "❌ Dead"
		if resp.ProcessAlive {
			aliveStr = "✅ Alive"
		}
		socketStr := "❌ Missing"
		if resp.SocketExists {
			socketStr = "✅ Available"
		}

		fmt.Printf("  PID:           %d (%s)\n", resp.Pid, aliveStr)
		fmt.Printf("  State:         %s\n", resp.State)
		fmt.Printf("  Uptime:        %s\n", resp.Uptime)
		fmt.Printf("  Created:       %s\n", resp.CreatedAt)

		// Machine
		fmt.Println()
		fmt.Println("  Machine:")
		fmt.Printf("    vCPUs:       %d\n", resp.Vcpus)
		fmt.Printf("    Memory:      %d MiB\n", resp.MemoryMib)
		fmt.Printf("    Kernel:      %s\n", resp.KernelImage)
		fmt.Printf("    Root Drive:  %s\n", resp.RootDrive)
		fmt.Printf("    Jailer:      %v\n", resp.UseJailer)

		// Paths
		fmt.Println()
		fmt.Println("  Paths:")
		fmt.Printf("    Socket:      %s (%s)\n", resp.SocketPath, socketStr)
		fmt.Printf("    Logs:        %s\n", resp.LogPath)
		fmt.Printf("    Metrics:     %s\n", resp.MetricsPath)
		fmt.Printf("    Workspace:   %s\n", resp.WorkDir)

		fmt.Println("════════════════════════════════════════")
		return nil
	},
}

func init() {
	inspectCmd.Flags().StringVar(&inspectDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	inspectCmd.Flags().StringVar(&inspectCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")
	inspectCmd.Flags().BoolVar(&inspectJSON, "json", false, "Output in JSON format")

	rootCmd.AddCommand(inspectCmd)
}
