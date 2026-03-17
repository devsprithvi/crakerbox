package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	listenAddr string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the cluster orchestrator",
	Long:  `Starts the Crackerbox Manager API server that orchestrates multiple crackerboxd nodes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔥 crackerbox-manager starting...")
		fmt.Printf("   Listening on %s\n", listenAddr)
		fmt.Println("   Press Ctrl+C to stop")
		fmt.Println()

		mux := http.NewServeMux()

		// Health check
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, Version)
		})

		// API stub — list nodes
		mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"nodes":[],"total":0}`)
		})

		// API stub — cluster info
		mux.HandleFunc("/api/v1/cluster", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"name":"crackerbox","version":"%s","nodes":0,"vms":0}`, Version)
		})

		log.Fatal(http.ListenAndServe(listenAddr, mux))
	},
}

func init() {
	serveCmd.Flags().StringVar(&listenAddr, "listen", ":9090", "Address to listen on")
	rootCmd.AddCommand(serveCmd)
}
