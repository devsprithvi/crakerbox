package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	callDaemonAddr string
	callCertDir    string
	callBody       string
)

var callCmd = &cobra.Command{
	Use:   "call <id> <method> <path>",
	Short: "Forward a raw HTTP request to a VM's Firecracker API socket",
	Long: `Forwards a raw HTTP request (e.g., GET /machine-config) directly to the
specified VM's Unix Domain Socket and returns the raw Firecracker API response.

Examples:
  crackerboxd call abc123 GET /machine-config
  crackerboxd call abc123 GET /
  crackerboxd call abc123 PUT /actions --body '{"action_type":"SendCtrlAltDel"}'
  crackerboxd call abc123 PATCH /machine-config --body '{"vcpu_count":4}'
  crackerboxd call abc123 GET /mmds`,

	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		method := args[1]
		path := args[2]

		conn, err := dialDaemon(callDaemonAddr, callCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		req := &api.CallRequest{
			Id:     vmID,
			Method: method,
			Path:   path,
		}

		if callBody != "" {
			req.Body = []byte(callBody)
		}

		resp, err := client.Call(context.Background(), req)
		if err != nil {
			return fmt.Errorf("call failed: %w", err)
		}

		// Print status
		fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.StatusCode)

		// Print headers
		for k, v := range resp.Headers {
			fmt.Fprintf(os.Stderr, "%s: %s\n", k, v)
		}
		fmt.Fprintln(os.Stderr)

		// Pretty-print JSON body if possible
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(resp.Body, &prettyJSON); err == nil {
			out, _ := json.MarshalIndent(prettyJSON, "", "  ")
			fmt.Println(string(out))
		} else {
			// Print raw body
			os.Stdout.Write(resp.Body)
			fmt.Println()
		}

		return nil
	},
}

func init() {
	callCmd.Flags().StringVar(&callDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	callCmd.Flags().StringVar(&callCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")
	callCmd.Flags().StringVar(&callBody, "body", "", "Request body (JSON)")

	rootCmd.AddCommand(callCmd)
}
