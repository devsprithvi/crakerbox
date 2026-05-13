// Package api provides the HTTP API server for crackerbox-manager.
// It serves as a proxy layer, forwarding requests to daemon nodes via mTLS gRPC.
package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/devsprithvi/crakerbox/apps/manager/pkg/client"
	"github.com/devsprithvi/crakerbox/apps/manager/pkg/cluster"
)

// Server is the HTTP API server for the manager.
// It authenticates clients via mTLS and proxies requests to daemon nodes.
type Server struct {
	cluster    *cluster.Cluster
	clientPool *client.ClientPool
	mux        *http.ServeMux
	certDir    string
	version    string
	startedAt  time.Time
}

// NewServer creates a new API server backed by the given cluster.
func NewServer(c *cluster.Cluster, pool *client.ClientPool, certDir, version string) *Server {
	s := &Server{
		cluster:    c,
		clientPool: pool,
		mux:        http.NewServeMux(),
		certDir:    certDir,
		version:    version,
		startedAt:  time.Now(),
	}
	s.registerRoutes()
	return s
}

// ListenAndServeTLS starts the server with mTLS.
func (s *Server) ListenAndServeTLS(addr string) error {
	tlsConfig, err := s.loadServerTLS()
	if err != nil {
		return fmt.Errorf("load server TLS: %w", err)
	}

	server := &http.Server{
		Addr:      addr,
		Handler:   s.mux,
		TLSConfig: tlsConfig,
	}

	certFile := filepath.Join(s.certDir, "server.pem")
	keyFile := filepath.Join(s.certDir, "server-key.pem")

	log.Printf("Manager API listening on %s (mTLS enabled)", addr)
	return server.ListenAndServeTLS(certFile, keyFile)
}

// Handler returns the http.Handler for the server (for plain HTTP fallback during dev).
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	// Health
	s.mux.HandleFunc("/healthz", s.handleHealthz)

	// Cluster management
	s.mux.HandleFunc("/api/v1/nodes", s.handleNodes)
	s.mux.HandleFunc("/api/v1/nodes/register", s.handleRegisterNode)
	s.mux.HandleFunc("/api/v1/cluster", s.handleCluster)

	// VM proxy — forwards to daemon nodes
	s.mux.HandleFunc("/api/v1/vms", s.handleListVMs)
	s.mux.HandleFunc("/api/v1/vms/spawn", s.handleSpawnVM)
}

// --- Handlers ---

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": s.version,
		"uptime":  time.Since(s.startedAt).Round(time.Second).String(),
	})
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.cluster.ListNodes()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"nodes": nodes,
		"total": len(nodes),
	})
}

type registerNodeRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func (s *Server) handleRegisterNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.ID == "" || req.Address == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id and address are required"})
		return
	}

	// Register node in cluster
	s.cluster.RegisterNode(req.ID, req.Address)

	// Attempt to connect via mTLS
	if err := s.clientPool.Connect(req.ID, req.Address); err != nil {
		log.Printf("WARN: connected node %s but mTLS connection failed: %v", req.ID, err)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "registered",
			"connected": false,
			"error":     err.Error(),
		})
		return
	}

	log.Printf("Node %s registered and connected at %s", req.ID, req.Address)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "registered",
		"connected": true,
	})
}

func (s *Server) handleCluster(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":    "crackerbox",
		"version": s.version,
		"nodes":   s.cluster.NodeCount(),
		"uptime":  time.Since(s.startedAt).Round(time.Second).String(),
	})
}

func (s *Server) handleListVMs(w http.ResponseWriter, r *http.Request) {
	// For now, return the connected nodes info
	// In full implementation, this would fan out gRPC List calls to all daemons
	connectedNodes := s.clientPool.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"connected_daemons": len(connectedNodes),
		"message":           "VM listing requires daemon connectivity — use the daemon CLI directly for now",
	})
}

func (s *Server) handleSpawnVM(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return info about available nodes
	connectedNodes := s.clientPool.List()
	if len(connectedNodes) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "no daemon nodes connected — register a node first",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "VM spawn proxy — coming soon. Use daemon CLI directly for now.",
		"available_nodes": connectedNodes,
	})
}

// --- TLS ---

func (s *Server) loadServerTLS() (*tls.Config, error) {
	caFile := filepath.Join(s.certDir, "ca.pem")

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caPool,
		MinVersion: tls.VersionTLS13,
	}, nil
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
