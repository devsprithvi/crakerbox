// Package daemon provides the core daemon lifecycle management for crackerboxd.
// It orchestrates the gRPC server, VM manager, and mTLS infrastructure.
package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"
	"crackerboxd/pkg/vm"
)

// Config holds the daemon configuration.
type Config struct {
	// Network
	ListenAddr string

	// Paths
	DataDir    string
	CertDir    string

	// Logging
	LogLevel   string

	// Firecracker
	FirecrackerBin string
	JailerBin      string

	// Version (injected at build time)
	Version string

	// TLS
	ServerHostnames []string
}

// DefaultConfig returns the default daemon configuration.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:      ":8090",
		DataDir:         "/var/lib/crackerbox",
		CertDir:         pki.DefaultCertDir,
		LogLevel:        "info",
		FirecrackerBin:  vm.DefaultFirecrackerBin,
		JailerBin:       vm.DefaultJailerBin,
		Version:         "0.1.0-dev",
		ServerHostnames: []string{"localhost", "127.0.0.1"},
	}
}

// Daemon is the main crackerboxd process.
type Daemon struct {
	config    *Config
	vmManager *vm.Manager
	server    *api.Server
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
// This is the main entry point that:
//   1. Ensures PKI certificates exist (generates if needed)
//   2. Initialises the VM manager
//   3. Starts the mTLS gRPC server
//   4. Blocks until shutdown signal
func (d *Daemon) Start(ctx context.Context) error {
	// 1. Ensure PKI is provisioned
	log.Println("Checking PKI certificates...")
	if err := pki.EnsurePKI(d.config.CertDir, d.config.ServerHostnames...); err != nil {
		return fmt.Errorf("provision PKI: %w", err)
	}
	if !pki.CertsExist(d.config.CertDir) {
		return fmt.Errorf("PKI certificates not found at %s — run 'crackerboxd certs generate'", d.config.CertDir)
	}
	log.Printf("PKI certificates loaded from %s", d.config.CertDir)

	// 2. Initialise VM manager
	vmDataDir := d.config.DataDir + "/vms"
	d.vmManager = vm.NewManager(vmDataDir)
	d.vmManager.FirecrackerBin = d.config.FirecrackerBin
	d.vmManager.JailerBin = d.config.JailerBin

	log.Printf("VM manager initialised — data dir: %s", vmDataDir)
	log.Printf("Firecracker binary: %s", d.config.FirecrackerBin)
	log.Printf("Jailer binary: %s", d.config.JailerBin)

	// Recover any previously tracked VMs
	entries := d.vmManager.List()
	if len(entries) > 0 {
		log.Printf("Recovered %d VM(s) from registry", len(entries))
	}

	// 3. Create and start the gRPC server
	var err error
	d.server, err = api.NewServer(&api.ServerConfig{
		ListenAddr: d.config.ListenAddr,
		CertDir:    d.config.CertDir,
		Version:    d.config.Version,
		VMManager:  d.vmManager,
	})
	if err != nil {
		return fmt.Errorf("create gRPC server: %w", err)
	}

	// 4. Handle shutdown signals
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			log.Printf("Received signal %v, shutting down...", sig)
			cancel()
			d.Stop()
		case <-ctx.Done():
		}
	}()

	// Print banner
	d.printBanner()

	// 5. Start serving (blocks)
	return d.server.Start(d.config.ListenAddr)
}

// Stop gracefully shuts down the daemon.
func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")

	if d.server != nil {
		d.server.Stop()
	}

	log.Println("Daemon stopped")
}

// VMManager returns the daemon's VM manager.
func (d *Daemon) VMManager() *vm.Manager {
	return d.vmManager
}

// printBanner prints the daemon startup banner.
func (d *Daemon) printBanner() {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════╗")
	fmt.Println("  ║          🔥 CRACKERBOXD — DAEMON ONLINE         ║")
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Printf("  ║  Version:    %-37s║\n", d.config.Version)
	fmt.Printf("  ║  Listen:     %-37s║\n", d.config.ListenAddr)
	fmt.Printf("  ║  Data Dir:   %-37s║\n", d.config.DataDir)
	fmt.Printf("  ║  Cert Dir:   %-37s║\n", d.config.CertDir)
	fmt.Println("  ║  Transport:  mTLS (TLS 1.3)                     ║")
	fmt.Println("  ║  Protocol:   gRPC                                ║")
	fmt.Println("  ╚══════════════════════════════════════════════════╝")
	fmt.Println()
}
