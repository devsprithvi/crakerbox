// Package api provides the HTTP API server for crackerbox-manager.
// This is the skeleton — handlers will be expanded as features are added.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/devsprithvi/crakerbox/apps/manager/pkg/cluster"
)

// Server is the HTTP API server for the manager.
type Server struct {
	cluster *cluster.Cluster
	mux     *http.ServeMux
}

// NewServer creates a new API server backed by the given cluster.
func NewServer(c *cluster.Cluster) *Server {
	s := &Server{
		cluster: c,
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler returns the http.Handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/v1/nodes", s.handleNodes)
	s.mux.HandleFunc("/api/v1/cluster", s.handleCluster)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.cluster.ListNodes()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"nodes": nodes,
		"total": len(nodes),
	})
}

func (s *Server) handleCluster(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":  "crackerbox",
		"nodes": s.cluster.NodeCount(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
