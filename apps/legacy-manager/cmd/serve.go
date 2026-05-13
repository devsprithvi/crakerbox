package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/devsprithvi/crakerbox/apps/manager/pkg/api"
	"github.com/devsprithvi/crakerbox/apps/manager/pkg/client"
	"github.com/devsprithvi/crakerbox/apps/manager/pkg/cluster"
	"github.com/spf13/cobra"
)

var (
	listenAddr string
	certDir    string
	useTLS     bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the cluster orchestrator",
	Long: `Starts the Crackerbox Manager API server that orchestrates multiple crackerboxd nodes.

The manager authenticates using mutual TLS and provides a REST API for cluster management.
It proxies requests to registered daemon nodes over mTLS-secured gRPC connections.

Examples:
  crackerbox-manager serve
  crackerbox-manager serve --listen :9090
  crackerbox-manager serve --cert-dir /opt/crackerbox/pki --tls
  crackerbox-manager serve --listen :9090 --tls`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Print banner
		fmt.Println()
		fmt.Println("  ╔══════════════════════════════════════════════════╗")
		fmt.Println("  ║       🔥 CRACKERBOX MANAGER — ONLINE            ║")
		fmt.Println("  ╠══════════════════════════════════════════════════╣")
		fmt.Printf("  ║  Version:    %-37s║\n", Version)
		fmt.Printf("  ║  Listen:     %-37s║\n", listenAddr)
		fmt.Printf("  ║  Cert Dir:   %-37s║\n", certDir)
		if useTLS {
			fmt.Println("  ║  Transport:  mTLS (TLS 1.3)                     ║")
		} else {
			fmt.Println("  ║  Transport:  HTTP (development mode)             ║")
		}
		fmt.Println("  ╚══════════════════════════════════════════════════╝")
		fmt.Println()

		// Initialize components
		c := cluster.New()
		pool := client.NewClientPool(certDir)
		defer pool.CloseAll()

		server := api.NewServer(c, pool, certDir, Version)

		// Handle shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigCh
			log.Printf("Received signal %v, shutting down...", sig)
			pool.CloseAll()
			os.Exit(0)
		}()

		// Start server
		if useTLS {
			if err := server.ListenAndServeTLS(listenAddr); err != nil {
				return fmt.Errorf("server error: %w", err)
			}
		} else {
			log.Printf("Manager API listening on %s (HTTP — development mode)", listenAddr)
			if err := http.ListenAndServe(listenAddr, server.Handler()); err != nil {
				return fmt.Errorf("server error: %w", err)
			}
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&listenAddr, "listen", ":9090", "Address to listen on")
	serveCmd.Flags().StringVar(&certDir, "cert-dir", "/etc/crackerbox/pki", "mTLS certificate directory")
	serveCmd.Flags().BoolVar(&useTLS, "tls", false, "Enable mTLS (requires certificates)")

	rootCmd.AddCommand(serveCmd)
}
