// Package cluster provides the core cluster management logic for crackerbox-manager.
// This is the skeleton — actual implementation will be added incrementally.
package cluster

import (
	"sync"
	"time"
)

// NodeState represents the health state of a daemon node.
type NodeState string

const (
	NodeStateUnknown  NodeState = "unknown"
	NodeStateHealthy  NodeState = "healthy"
	NodeStateDegraded NodeState = "degraded"
	NodeStateOffline  NodeState = "offline"
)

// Node represents a registered crackerboxd instance.
type Node struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	State     NodeState `json:"state"`
	VMCount   int       `json:"vm_count"`
	JoinedAt  time.Time `json:"joined_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// Cluster holds the state of all registered nodes.
type Cluster struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

// New creates a new empty Cluster.
func New() *Cluster {
	return &Cluster{
		nodes: make(map[string]*Node),
	}
}

// RegisterNode adds a node to the cluster.
func (c *Cluster) RegisterNode(id, address string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[id] = &Node{
		ID:       id,
		Address:  address,
		State:    NodeStateUnknown,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
	}
}

// RemoveNode removes a node from the cluster.
func (c *Cluster) RemoveNode(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nodes, id)
}

// ListNodes returns all registered nodes.
func (c *Cluster) ListNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nodes := make([]*Node, 0, len(c.nodes))
	for _, n := range c.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// NodeCount returns the number of registered nodes.
func (c *Cluster) NodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}
