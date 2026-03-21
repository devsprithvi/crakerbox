package cmd

import (
	"context"
	"fmt"
	"os"

	"crackerboxd/pkg/daemon"
	"crackerboxd/pkg/pki"
	"crackerboxd/pkg/vm"

	"github.com/spf13/cobra"
)

var (
	serveListenAddr    string
	serveDataDir       string
	serveCertDir       string
	serveFirecrackerBin string
	serveJailerBin     string
	serveHostnames     []string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the crackerboxd daemon",
	Long: `Starts the Crackerbox daemon process that manages Firecracker microVMs.

The daemon exposes a mTLS-secured gRPC API for the manager and CLI to connect to.
On first run, it automatically generates the PKI certificates if they don't exist.

Examples:
  crackerboxd serve
  crackerboxd serve --listen :9090
  crackerboxd serve --cert-dir /opt/crackerbox/pki
  crackerboxd serve --firecracker-bin /usr/local/bin/firecracker`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := daemon.DefaultConfig()

		// Apply CLI overrides
		if serveListenAddr != "" {
			cfg.ListenAddr = serveListenAddr
		}
		if serveDataDir != "" {
			cfg.DataDir = serveDataDir
		}
		if serveCertDir != "" {
			cfg.CertDir = serveCertDir
		}
		if serveFirecrackerBin != "" {
			cfg.FirecrackerBin = serveFirecrackerBin
		}
		if serveJailerBin != "" {
			cfg.JailerBin = serveJailerBin
		}
		if len(serveHostnames) > 0 {
			cfg.ServerHostnames = serveHostnames
		}

		cfg.Version = Version

		d := daemon.New(cfg)
		if err := d.Start(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveListenAddr, "listen", ":8090", "gRPC listen address")
	serveCmd.Flags().StringVar(&serveDataDir, "data-dir", "/var/lib/crackerbox", "Data directory for VM workspaces")
	serveCmd.Flags().StringVar(&serveCertDir, "cert-dir", pki.DefaultCertDir, "Directory for mTLS certificates")
	serveCmd.Flags().StringVar(&serveFirecrackerBin, "firecracker-bin", vm.DefaultFirecrackerBin, "Path to firecracker binary")
	serveCmd.Flags().StringVar(&serveJailerBin, "jailer-bin", vm.DefaultJailerBin, "Path to jailer binary")
	serveCmd.Flags().StringSliceVar(&serveHostnames, "hostnames", []string{"localhost", "127.0.0.1"}, "Server certificate SANs (hostnames and IPs)")

	rootCmd.AddCommand(serveCmd)
}
