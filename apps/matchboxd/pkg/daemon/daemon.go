// Package daemon provides the core daemon lifecycle management for crackerboxd.
// This is the skeleton — actual Firecracker integration comes later.
package daemon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Config holds the daemon configuration.
type Config struct {
	ListenAddr string
	DataDir    string
	LogLevel   string
}

// DefaultConfig returns the default daemon configuration.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: ":8090",
		DataDir:    "/var/lib/crackerbox",
		LogLevel:   "info",
	}
}

// Daemon is the main crackerboxd process.
type Daemon struct {
	config *Config
	server *http.Server
	mu     sync.RWMutex

	startedAt time.Time
	running   bool
}

// New creates a new Daemon instance with the given config.
func New(cfg *Config) *Daemon {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Daemon{
		config: cfg,
	}
}

// Start initialises and starts the daemon.
func (d *Daemon) Start(ctx context.Context) error {
	d.mu.Lock()
	d.running = true
	d.startedAt = time.Now()
	d.mu.Unlock()

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","uptime":"%s"}`, time.Since(d.startedAt).Round(time.Second))
	})

	// VM list stub
	mux.HandleFunc("/api/v1/vms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"vms":[],"total":0}`)
	})

	d.server = &http.Server{
		Addr:    d.config.ListenAddr,
		Handler: mux,
	}

	log.Printf("crackerboxd listening on %s", d.config.ListenAddr)

	go func() {
		<-ctx.Done()
		d.Stop()
	}()

	return d.server.ListenAndServe()
}

// Stop gracefully shuts down the daemon.
func (d *Daemon) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.running = false
	if d.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return d.server.Shutdown(ctx)
	}
	return nil
}

// IsRunning reports whether the daemon is currently running.
func (d *Daemon) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}
