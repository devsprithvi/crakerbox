// Package vm provides Firecracker microVM lifecycle management using the
// Firecracker Go SDK. It handles VM spawning, termination, inspection, and
// bridging to the Firecracker API socket.
package vm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/google/uuid"
)

const (
	// DefaultDataDir is the base directory for VM workspaces.
	DefaultDataDir = "/var/lib/crackerbox/vms"

	// DefaultFirecrackerBin is the default path to the firecracker binary.
	DefaultFirecrackerBin = "/usr/bin/firecracker"

	// DefaultJailerBin is the default path to the jailer binary.
	DefaultJailerBin = "/usr/bin/jailer"

	// shutdownTimeout is the grace period for VM shutdown before SIGKILL.
	shutdownTimeout = 10 * time.Second
)

// Manager handles the lifecycle of Firecracker microVMs on this node.
// It uses the Firecracker Go SDK for machine creation and management.
type Manager struct {
	mu sync.RWMutex

	// machines maps VM ID to the running Firecracker SDK machine instance.
	machines map[string]*firecracker.Machine

	// cancels maps VM ID to the context cancel function for that machine.
	cancels map[string]context.CancelFunc

	// Registry persists process metadata to disk.
	Registry *Registry

	// Configuration
	DataDir        string
	FirecrackerBin string
	JailerBin      string
}

// NewManager creates a new VM manager.
func NewManager(dataDir string) *Manager {
	if dataDir == "" {
		dataDir = DefaultDataDir
	}

	m := &Manager{
		machines:       make(map[string]*firecracker.Machine),
		cancels:        make(map[string]context.CancelFunc),
		Registry:       NewRegistry(dataDir),
		DataDir:        dataDir,
		FirecrackerBin: DefaultFirecrackerBin,
		JailerBin:      DefaultJailerBin,
	}

	// Load any persisted state.
	if err := m.Registry.Load(); err != nil {
		log.Printf("WARN: failed to load registry: %v", err)
	}

	return m
}

// Spawn creates and starts a new Firecracker microVM.
// This is the daemon's primary command — it resolves configuration,
// provisions host resources, and hands off to the Firecracker Go SDK.
func (m *Manager) Spawn(ctx context.Context, cfg *Config) (*ProcessEntry, error) {
	// 1. Assign identity
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()[:8]
	}
	if cfg.Name == "" {
		cfg.Name = "vm-" + cfg.ID
	}

	// 2. Resolve workspace paths
	if err := cfg.ResolveWorkspace(m.DataDir); err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}

	// 3. Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// 4. Create log file
	logFile, err := os.Create(cfg.LogPath)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}

	// 5. Create metrics FIFO
	if err := createFifo(cfg.MetricsPath); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("create metrics fifo: %w", err)
	}

	// 6. Build Firecracker SDK configuration
	fcCfg, err := m.buildFirecrackerConfig(cfg, logFile)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("build firecracker config: %w", err)
	}

	// 7. Create the machine context
	machineCtx, machineCancel := context.WithCancel(ctx)

	// 8. Determine command builder (direct or via jailer)
	var machineOpts []firecracker.Opt

	if cfg.UseJailer {
		jailerCfg := m.buildJailerConfig(cfg)
		machineOpts = append(machineOpts,
			firecracker.WithProcessRunner(jailerCommand(machineCtx, jailerCfg, cfg)),
		)
	} else {
		firecrackerBin := m.FirecrackerBin
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(firecrackerBin).
			WithSocketPath(cfg.SocketPath).
			WithStdout(logFile).
			WithStderr(logFile).
			Build(machineCtx)

		machineOpts = append(machineOpts,
			firecracker.WithProcessRunner(cmd),
		)
	}

	// 9. Create the Firecracker machine
	machine, err := firecracker.NewMachine(machineCtx, fcCfg, machineOpts...)
	if err != nil {
		machineCancel()
		logFile.Close()
		return nil, fmt.Errorf("create machine: %w", err)
	}

	// 10. Start the machine
	if err := machine.Start(machineCtx); err != nil {
		machineCancel()
		logFile.Close()
		return nil, fmt.Errorf("start machine: %w", err)
	}

	// 11. Register in the process registry
	now := time.Now()
	pid := 0
	if machine.Cfg.JailerCfg == nil {
		// Direct mode: get the PID from the command
		if p, err := machine.PID(); err == nil && p > 0 {
			pid = p
		}
	}

	entry := &ProcessEntry{
		ID:          cfg.ID,
		Name:        cfg.Name,
		PID:         pid,
		State:       ProcessStateRunning,
		Config:      cfg,
		SocketPath:  cfg.SocketPath,
		LogPath:     cfg.LogPath,
		MetricsPath: cfg.MetricsPath,
		WorkDir:     cfg.WorkDir,
		StartedAt:   &now,
	}

	if err := m.Registry.Register(entry); err != nil {
		// Non-fatal: the VM is running, we just can't track it
		log.Printf("WARN: failed to register VM %s: %v", cfg.ID, err)
	}

	// Store machine reference
	m.mu.Lock()
	m.machines[cfg.ID] = machine
	m.cancels[cfg.ID] = machineCancel
	m.mu.Unlock()

	// 12. Monitor the machine in the background
	go m.monitorMachine(cfg.ID, machine, machineCtx, logFile)

	log.Printf("VM %s (%s) started — PID %d, socket %s", cfg.ID, cfg.Name, pid, cfg.SocketPath)
	return entry, nil
}

// Terminate stops a running Firecracker microVM and cleans up host resources.
func (m *Manager) Terminate(ctx context.Context, id string) error {
	m.mu.Lock()
	machine, ok := m.machines[id]
	cancel, hasCancel := m.cancels[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("VM %s not found or not running", id)
	}

	// Update registry state
	_ = m.Registry.Update(id, func(e *ProcessEntry) {
		e.State = ProcessStateStopping
	})

	// Try graceful shutdown via the Firecracker API
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	err := machine.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("WARN: graceful shutdown of VM %s failed: %v, sending SIGKILL", id, err)
		// Force stop
		if stopErr := machine.StopVMM(); stopErr != nil {
			log.Printf("WARN: StopVMM for VM %s failed: %v", id, stopErr)
		}
	}

	// Wait for the process to exit
	if err := machine.Wait(shutdownCtx); err != nil {
		log.Printf("WARN: wait for VM %s: %v", id, err)
	}

	// Cancel the machine context
	if hasCancel {
		cancel()
	}

	// Update registry
	now := time.Now()
	_ = m.Registry.Update(id, func(e *ProcessEntry) {
		e.State = ProcessStateStopped
		e.StoppedAt = &now
	})

	// Cleanup machine references
	m.mu.Lock()
	delete(m.machines, id)
	delete(m.cancels, id)
	m.mu.Unlock()

	// Cleanup workspace
	entry, found := m.Registry.Get(id)
	if found && entry.WorkDir != "" {
		if err := os.RemoveAll(entry.WorkDir); err != nil {
			log.Printf("WARN: cleanup workspace for VM %s: %v", id, err)
		}
	}

	// Remove from registry
	_ = m.Registry.Remove(id)

	log.Printf("VM %s terminated and cleaned up", id)
	return nil
}

// List returns all tracked VM process entries.
func (m *Manager) List() []*ProcessEntry {
	entries := m.Registry.List()

	// Reconcile with actual process state
	for _, e := range entries {
		if e.State == ProcessStateRunning {
			if !m.isProcessAlive(e.PID) {
				e.State = ProcessStateStopped
			}
		}
	}

	return entries
}

// Get returns a single VM process entry by ID.
func (m *Manager) Get(id string) (*ProcessEntry, bool) {
	return m.Registry.Get(id)
}

// Inspect returns detailed host-level information about a VM.
func (m *Manager) Inspect(id string) (*InspectResult, error) {
	entry, ok := m.Registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("VM %s not found", id)
	}

	result := &InspectResult{
		ID:          entry.ID,
		Name:        entry.Name,
		PID:         entry.PID,
		State:       entry.State,
		SocketPath:  entry.SocketPath,
		LogPath:     entry.LogPath,
		MetricsPath: entry.MetricsPath,
		WorkDir:     entry.WorkDir,
		CreatedAt:   entry.CreatedAt,
		Uptime:      entry.Uptime().String(),
	}

	if entry.Config != nil {
		result.VCPUs = entry.Config.VCPUs
		result.MemoryMiB = entry.Config.MemoryMiB
		result.KernelImage = entry.Config.KernelImagePath
		result.RootDrive = entry.Config.RootDrivePath
		result.UseJailer = entry.Config.UseJailer
	}

	// Check if process is alive
	result.ProcessAlive = m.isProcessAlive(entry.PID)

	// Check socket availability
	if _, err := os.Stat(entry.SocketPath); err == nil {
		result.SocketExists = true
	}

	return result, nil
}

// InspectResult holds detailed host-level information about a VM.
type InspectResult struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	PID         int          `json:"pid"`
	State       ProcessState `json:"state"`
	VCPUs       int          `json:"vcpus"`
	MemoryMiB   int          `json:"memory_mib"`
	KernelImage string       `json:"kernel_image"`
	RootDrive   string       `json:"root_drive"`
	UseJailer   bool         `json:"use_jailer"`

	SocketPath   string    `json:"socket_path"`
	LogPath      string    `json:"log_path"`
	MetricsPath  string    `json:"metrics_path"`
	WorkDir      string    `json:"work_dir"`
	CreatedAt    time.Time `json:"created_at"`
	Uptime       string    `json:"uptime"`
	ProcessAlive bool      `json:"process_alive"`
	SocketExists bool      `json:"socket_exists"`
}

// ReadLogs reads the log file for the given VM. Returns the last `lines` lines
// or all content if lines <= 0.
func (m *Manager) ReadLogs(id string, follow bool) (io.ReadCloser, error) {
	entry, ok := m.Registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("VM %s not found", id)
	}

	f, err := os.Open(entry.LogPath)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return f, nil
}

// ReadMetrics reads the latest metrics from the VM's metrics file.
func (m *Manager) ReadMetrics(id string) ([]byte, error) {
	entry, ok := m.Registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("VM %s not found", id)
	}

	data, err := os.ReadFile(entry.MetricsPath)
	if err != nil {
		return nil, fmt.Errorf("read metrics: %w", err)
	}

	return data, nil
}

// Call forwards a raw HTTP request to the Firecracker API via the VM's Unix socket.
func (m *Manager) Call(ctx context.Context, id, method, path string, body io.Reader) (*http.Response, error) {
	entry, ok := m.Registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("VM %s not found", id)
	}

	// Verify socket exists
	if _, err := os.Stat(entry.SocketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("socket not found: %s (VM may not be running)", entry.SocketPath)
	}

	// Create HTTP client that dials the Unix socket
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", entry.SocketPath)
			},
		},
	}

	// Build the request
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := "http://localhost" + path

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

// --- Firecracker SDK Configuration Builders ---

// buildFirecrackerConfig translates our Config into the Firecracker SDK's Config type.
func (m *Manager) buildFirecrackerConfig(cfg *Config, logFile *os.File) (firecracker.Config, error) {
	drives := []models.Drive{
		{
			DriveID:      firecracker.String(cfg.RootDriveID),
			PathOnHost:   firecracker.String(cfg.RootDrivePath),
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(cfg.ReadOnly),
		},
	}

	var networkInterfaces firecracker.NetworkInterfaces
	for _, ni := range cfg.NetworkInterfaces {
		networkInterfaces = append(networkInterfaces, firecracker.NetworkInterface{
			StaticConfiguration: &firecracker.StaticNetworkConfiguration{
				MacAddress:  ni.GuestMAC,
				HostDevName: ni.HostDevName,
			},
			AllowMMDS: ni.AllowMMDSAccess,
		})
	}

	fcCfg := firecracker.Config{
		SocketPath:      cfg.SocketPath,
		KernelImagePath: cfg.KernelImagePath,
		InitrdPath:      cfg.InitrdPath,
		KernelArgs:      cfg.KernelArgs,
		Drives:          drives,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(int64(cfg.VCPUs)),
			MemSizeMib: firecracker.Int64(int64(cfg.MemoryMiB)),
			Smt:        firecracker.Bool(cfg.SMT),
		},
		NetworkInterfaces: networkInterfaces,
		LogPath:           cfg.LogPath,
		MetricsPath:       cfg.MetricsPath,
	}

	// If using jailer, set the jailer config
	if cfg.UseJailer {
		fcCfg.JailerCfg = &firecracker.JailerConfig{
			ID:             cfg.ID,
			UID:            firecracker.Int(cfg.JailerUID),
			GID:            firecracker.Int(cfg.JailerGID),
			NumaNode:       firecracker.Int(0),
			ExecFile:       m.FirecrackerBin,
			JailerBinary:   cfg.JailerBin,
			ChrootBaseDir:  cfg.ChrootBase,
		}
	}

	return fcCfg, nil
}

// JailerConfig holds jailer-specific configuration.
type JailerConfig struct {
	ID        string
	UID       int
	GID       int
	NumaNode  int
	ExecFile  string
	JailerBin string
	ChrootDir string
}

// buildJailerConfig creates a JailerConfig from the VM Config.
func (m *Manager) buildJailerConfig(cfg *Config) *JailerConfig {
	chrootBase := cfg.ChrootBase
	if chrootBase == "" {
		chrootBase = "/srv/jailer"
	}

	return &JailerConfig{
		ID:        cfg.ID,
		UID:       cfg.JailerUID,
		GID:       cfg.JailerGID,
		NumaNode:  0,
		ExecFile:  m.FirecrackerBin,
		JailerBin: cfg.JailerBin,
		ChrootDir: chrootBase,
	}
}

// jailerCommand builds an exec.Cmd for running Firecracker via the jailer.
func jailerCommand(ctx context.Context, jcfg *JailerConfig, vmCfg *Config) *exec.Cmd {
	args := []string{
		"--id", jcfg.ID,
		"--exec-file", jcfg.ExecFile,
		"--uid", fmt.Sprintf("%d", jcfg.UID),
		"--gid", fmt.Sprintf("%d", jcfg.GID),
		"--chroot-base-dir", jcfg.ChrootDir,
		"--daemonize",
	}

	cmd := exec.CommandContext(ctx, jcfg.JailerBin, args...)
	setProcGroup(cmd)
	return cmd
}

// --- Background Monitoring ---

// monitorMachine watches a running machine and updates the registry on exit.
func (m *Manager) monitorMachine(id string, machine *firecracker.Machine, ctx context.Context, logFile *os.File) {
	defer logFile.Close()

	// Wait for the machine to exit
	if err := machine.Wait(ctx); err != nil && ctx.Err() == nil {
		log.Printf("VM %s exited with error: %v", id, err)
		_ = m.Registry.Update(id, func(e *ProcessEntry) {
			e.State = ProcessStateError
			e.LastError = err.Error()
			now := time.Now()
			e.StoppedAt = &now
		})
	} else {
		now := time.Now()
		_ = m.Registry.Update(id, func(e *ProcessEntry) {
			e.State = ProcessStateStopped
			e.StoppedAt = &now
		})
	}

	// Remove machine references
	m.mu.Lock()
	delete(m.machines, id)
	delete(m.cancels, id)
	m.mu.Unlock()

	log.Printf("VM %s monitor exited", id)
}

// isProcessAlive checks if a process with the given PID is still running.
func (m *Manager) isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return isProcessAliveOS(pid)
}

// createFifo and other platform-specific helpers are in platform_linux.go / platform_other.go
