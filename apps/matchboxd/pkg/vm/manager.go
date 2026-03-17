// Package vm provides the Firecracker VM lifecycle management.
// This is the skeleton — actual Firecracker socket communication comes later.
package vm

import (
	"sync"
	"time"
)

// State represents the current lifecycle state of a VM.
type State string

const (
	StateCreating State = "creating"
	StateRunning  State = "running"
	StatePaused   State = "paused"
	StateStopped  State = "stopped"
	StateError    State = "error"
)

// VM represents a Firecracker microVM instance.
type VM struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	State     State     `json:"state"`
	VCPUs     int       `json:"vcpus"`
	MemoryMiB int       `json:"memory_mib"`
	KernelImg string    `json:"kernel_image"`
	RootFS    string    `json:"rootfs"`
	CreatedAt time.Time `json:"created_at"`
	SocketPath string   `json:"socket_path"`
}

// Manager handles the lifecycle of Firecracker VMs on this node.
type Manager struct {
	mu  sync.RWMutex
	vms map[string]*VM
}

// NewManager creates a new VM manager.
func NewManager() *Manager {
	return &Manager{
		vms: make(map[string]*VM),
	}
}

// List returns all VMs managed by this node.
func (m *Manager) List() []*VM {
	m.mu.RLock()
	defer m.mu.RUnlock()
	vms := make([]*VM, 0, len(m.vms))
	for _, vm := range m.vms {
		vms = append(vms, vm)
	}
	return vms
}

// Get returns a VM by ID.
func (m *Manager) Get(id string) (*VM, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	vm, ok := m.vms[id]
	return vm, ok
}

// Count returns the number of VMs.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.vms)
}

// TODO: Create, Start, Stop, Delete — will use Firecracker API via unix socket
