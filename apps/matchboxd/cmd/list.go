package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	listDaemonAddr  string
	listCertDir     string
	listStateFilter string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all running Firecracker processes",
	Long: `Shows all Firecracker microVM processes managed by this daemon,
including their IDs, PIDs, states, and API socket paths.

Examples:
  crackerboxd list
  crackerboxd list --state running
  crackerboxd list --daemon 10.0.0.5:8090`,

	Aliases: []string{"ls", "ps"},
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialDaemon(listDaemonAddr, listCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		resp, err := client.List(context.Background(), &api.ListRequest{
			StateFilter: listStateFilter,
		})
		if err != nil {
			return fmt.Errorf("list failed: %w", err)
		}

		if resp.Total == 0 {
			fmt.Println("No Firecracker processes running.")
			return nil
		}

		// Print as formatted table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPID\tSTATE\tUPTIME\tSOCKET")
		fmt.Fprintln(w, "──\t────\t───\t─────\t──────\t──────")

		for _, vm := range resp.Vms {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
				vm.Id, vm.Name, vm.Pid, vm.State, vm.Uptime, vm.SocketPath)
		}

		w.Flush()
		fmt.Printf("\nTotal: %d\n", resp.Total)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	listCmd.Flags().StringVar(&listCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")
	listCmd.Flags().StringVar(&listStateFilter, "state", "", "Filter by state (running, stopped, error)")

	rootCmd.AddCommand(listCmd)
}
