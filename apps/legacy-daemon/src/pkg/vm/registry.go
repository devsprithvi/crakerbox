package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ProcessState represents the lifecycle state of a managed Firecracker process.
type ProcessState string

const (
	ProcessStateStarting    ProcessState = "starting"
	ProcessStateRunning     ProcessState = "running"
	ProcessStateStopping    ProcessState = "stopping"
	ProcessStateStopped     ProcessState = "stopped"
	ProcessStateError       ProcessState = "error"
)

// ProcessEntry tracks a single Firecracker process managed by the daemon.
type ProcessEntry struct {
	// Identity
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`

	// Process
	PID   int          `json:"pid"`
	State ProcessState `json:"state"`

	// Configuration
	Config *Config `json:"config"`

	// Runtime Paths
	SocketPath  string `json:"socket_path"`
	LogPath     string `json:"log_path"`
	MetricsPath string `json:"metrics_path"`
	WorkDir     string `json:"work_dir"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`

	// Error tracking
	LastError string `json:"last_error,omitempty"`
}

// Uptime returns the duration since the process was started.
func (p *ProcessEntry) Uptime() time.Duration {
	if p.StartedAt == nil {
		return 0
	}
	if p.StoppedAt != nil {
		return p.StoppedAt.Sub(*p.StartedAt)
	}
	return time.Since(*p.StartedAt)
}

// Registry tracks all Firecracker processes managed by this daemon.
// It persists state to disk so processes can be recovered after daemon restart.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*ProcessEntry
	dataDir string
}

// NewRegistry creates a new process registry backed by the given data directory.
func NewRegistry(dataDir string) *Registry {
	return &Registry{
		entries: make(map[string]*ProcessEntry),
		dataDir: dataDir,
	}
}

// Register adds a new process entry to the registry.
func (r *Registry) Register(entry *ProcessEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[entry.ID]; exists {
		return fmt.Errorf("VM %s already registered", entry.ID)
	}

	entry.CreatedAt = time.Now()
	r.entries[entry.ID] = entry

	return r.persistLocked()
}

// Update modifies an existing process entry.
func (r *Registry) Update(id string, fn func(*ProcessEntry)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[id]
	if !ok {
		return fmt.Errorf("VM %s not found", id)
	}

	fn(entry)
	return r.persistLocked()
}

// Remove deletes a process entry from the registry.
func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.entries[id]; !ok {
		return fmt.Errorf("VM %s not found", id)
	}

	delete(r.entries, id)
	return r.persistLocked()
}

// Get returns a process entry by ID.
func (r *Registry) Get(id string) (*ProcessEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	return e, ok
}

// List returns all process entries.
func (r *Registry) List() []*ProcessEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*ProcessEntry, 0, len(r.entries))
	for _, e := range r.entries {
		entries = append(entries, e)
	}
	return entries
}

// Count returns the number of registered processes.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Load reads persisted registry state from disk.
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := r.registryPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // no persisted state yet
	}
	if err != nil {
		return fmt.Errorf("read registry: %w", err)
	}

	var entries map[string]*ProcessEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("parse registry: %w", err)
	}

	r.entries = entries
	return nil
}

// persistLocked writes current state to disk. Caller must hold r.mu.
func (r *Registry) persistLocked() error {
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	data, err := json.MarshalIndent(r.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	path := r.registryPath()
	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return os.Rename(tmpPath, path)
}

func (r *Registry) registryPath() string {
	return filepath.Join(r.dataDir, "registry.json")
}
