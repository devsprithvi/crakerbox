// Package client provides a mTLS-secured gRPC client for communicating
// with crackerboxd daemon instances. This is used by the manager to
// proxy requests to any registered daemon node.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// DefaultCertDir is the default directory for client certificates.
	DefaultCertDir = "/etc/crackerbox/pki"

	// Connection timeout for dialing a daemon.
	dialTimeout = 10 * time.Second
)

// DaemonClient wraps a gRPC connection to a single crackerboxd instance.
type DaemonClient struct {
	mu   sync.RWMutex
	conn *grpc.ClientConn
	addr string
}

// NewDaemonClient creates a new mTLS-secured client connected to a daemon.
func NewDaemonClient(addr, certDir string) (*DaemonClient, error) {
	tlsConfig, err := loadClientTLS(certDir)
	if err != nil {
		return nil, fmt.Errorf("load client TLS: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to daemon at %s: %w", addr, err)
	}

	return &DaemonClient{
		conn: conn,
		addr: addr,
	}, nil
}

// Conn returns the underlying gRPC connection for making RPC calls.
func (c *DaemonClient) Conn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// Close closes the gRPC connection.
func (c *DaemonClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Address returns the daemon address this client is connected to.
func (c *DaemonClient) Address() string {
	return c.addr
}

// Health checks if the daemon is reachable and healthy.
func (c *DaemonClient) Health(ctx context.Context) error {
	// Simple connectivity check via gRPC
	state := c.conn.GetState()
	if state.String() == "TRANSIENT_FAILURE" || state.String() == "SHUTDOWN" {
		return fmt.Errorf("connection in state: %s", state)
	}
	return nil
}

// --- Pool of Daemon Clients ---

// ClientPool manages connections to multiple daemon instances.
type ClientPool struct {
	mu      sync.RWMutex
	clients map[string]*DaemonClient
	certDir string
}

// NewClientPool creates a new pool for daemon connections.
func NewClientPool(certDir string) *ClientPool {
	return &ClientPool{
		clients: make(map[string]*DaemonClient),
		certDir: certDir,
	}
}

// Connect establishes a connection to a daemon and adds it to the pool.
func (p *ClientPool) Connect(nodeID, addr string) error {
	client, err := NewDaemonClient(addr, p.certDir)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Close existing connection if any
	if existing, ok := p.clients[nodeID]; ok {
		existing.Close()
	}

	p.clients[nodeID] = client
	return nil
}

// Get returns a client for the given node ID.
func (p *ClientPool) Get(nodeID string) (*DaemonClient, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	c, ok := p.clients[nodeID]
	return c, ok
}

// Remove disconnects and removes a client from the pool.
func (p *ClientPool) Remove(nodeID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.clients[nodeID]; ok {
		c.Close()
		delete(p.clients, nodeID)
	}
}

// List returns all connected node IDs.
func (p *ClientPool) List() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]string, 0, len(p.clients))
	for id := range p.clients {
		ids = append(ids, id)
	}
	return ids
}

// CloseAll disconnects all clients.
func (p *ClientPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, c := range p.clients {
		c.Close()
		delete(p.clients, id)
	}
}

// --- Streaming Helpers ---

// LogStream wraps the server streaming of log data.
type LogStream interface {
	Recv() ([]byte, error)
}

// StreamToWriter copies streaming data to a writer (e.g., os.Stdout).
func StreamToWriter(stream LogStream, w io.Writer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
}

// --- TLS Helpers ---

func loadClientTLS(certDir string) (*tls.Config, error) {
	certFile := filepath.Join(certDir, "client.pem")
	keyFile := filepath.Join(certDir, "client-key.pem")
	caFile := filepath.Join(certDir, "ca.pem")

	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load client certificate from %s: %w", certDir, err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
