// Package api provides the mTLS-secured gRPC server for crackerboxd.
// It exposes the DaemonService RPC methods that the manager connects to.
package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"crackerboxd/pkg/pki"
	"crackerboxd/pkg/vm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Server is the mTLS-secured gRPC server for the daemon.
type Server struct {
	vmManager  *vm.Manager
	grpcServer *grpc.Server
	listener   net.Listener

	startedAt time.Time
	version   string
	certDir   string
}

// ServerConfig holds configuration for the gRPC server.
type ServerConfig struct {
	ListenAddr string
	CertDir    string
	Version    string
	VMManager  *vm.Manager
}

// NewServer creates a new gRPC server with mTLS authentication.
func NewServer(cfg *ServerConfig) (*Server, error) {
	// Load server-side TLS configuration with client cert verification
	tlsConfig, err := pki.LoadServerTLSConfig(cfg.CertDir)
	if err != nil {
		return nil, fmt.Errorf("load TLS config: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)

	// Create gRPC server with TLS credentials and interceptors
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
		grpc.StreamInterceptor(loggingStreamInterceptor),
	)

	s := &Server{
		vmManager:  cfg.VMManager,
		grpcServer: grpcServer,
		startedAt:  time.Now(),
		version:    cfg.Version,
		certDir:    cfg.CertDir,
	}

	// Register the daemon service
	RegisterDaemonServiceServer(grpcServer, s)

	return s, nil
}

// Start begins listening and serving gRPC requests.
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	s.listener = lis

	log.Printf("gRPC server listening on %s (mTLS enabled)", addr)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// --- DaemonService RPC Implementations ---

// Spawn starts a new Firecracker microVM.
func (s *Server) Spawn(ctx context.Context, req *SpawnRequest) (*SpawnResponse, error) {
	clientCN := extractClientCN(ctx)
	log.Printf("Spawn request from %s: name=%s", clientCN, req.Name)

	cfg := &vm.Config{}

	// Load from config file if path is provided
	if req.ConfigPath != "" {
		loaded, err := vm.LoadConfigFromFile(req.ConfigPath)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "load config: %v", err)
		}
		cfg = loaded
	}

	// Override with dynamic fields
	if req.Vcpus > 0 {
		cfg.VCPUs = int(req.Vcpus)
	}
	if req.MemoryMib > 0 {
		cfg.MemoryMiB = int(req.MemoryMib)
	}
	if req.KernelImagePath != "" {
		cfg.KernelImagePath = req.KernelImagePath
	}
	if req.RootDrivePath != "" {
		cfg.RootDrivePath = req.RootDrivePath
	}
	if req.KernelArgs != "" {
		cfg.KernelArgs = req.KernelArgs
	}
	if req.Name != "" {
		cfg.Name = req.Name
	}
	cfg.UseJailer = req.UseJailer

	// Network interfaces
	for _, ni := range req.NetworkInterfaces {
		cfg.NetworkInterfaces = append(cfg.NetworkInterfaces, vm.NetworkInterface{
			ID:              ni.IfaceId,
			HostDevName:     ni.HostDevName,
			GuestMAC:        ni.GuestMac,
			AllowMMDSAccess: ni.AllowMmdsRequests,
		})
	}

	entry, err := s.vmManager.Spawn(ctx, cfg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "spawn VM: %v", err)
	}

	return &SpawnResponse{
		Id:         entry.ID,
		Name:       entry.Name,
		Pid:        int32(entry.PID),
		SocketPath: entry.SocketPath,
		State:      string(entry.State),
	}, nil
}

// Terminate stops a running Firecracker microVM.
func (s *Server) Terminate(ctx context.Context, req *TerminateRequest) (*TerminateResponse, error) {
	clientCN := extractClientCN(ctx)
	log.Printf("Terminate request from %s: id=%s", clientCN, req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "VM ID is required")
	}

	if err := s.vmManager.Terminate(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "terminate VM: %v", err)
	}

	return &TerminateResponse{
		Success: true,
		Message: fmt.Sprintf("VM %s terminated successfully", req.Id),
	}, nil
}

// List returns all managed Firecracker processes.
func (s *Server) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	entries := s.vmManager.List()

	var vms []*VMEntry
	for _, e := range entries {
		// Apply state filter if provided
		if req.StateFilter != "" && string(e.State) != req.StateFilter {
			continue
		}

		vms = append(vms, &VMEntry{
			Id:         e.ID,
			Name:       e.Name,
			Pid:        int32(e.PID),
			State:      string(e.State),
			SocketPath: e.SocketPath,
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
			Uptime:     e.Uptime().Round(time.Second).String(),
		})
	}

	return &ListResponse{
		Vms:   vms,
		Total: int32(len(vms)),
	}, nil
}

// Inspect returns detailed host-level information about a specific VM.
func (s *Server) Inspect(ctx context.Context, req *InspectRequest) (*InspectResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "VM ID is required")
	}

	result, err := s.vmManager.Inspect(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "inspect VM: %v", err)
	}

	return &InspectResponse{
		Id:           result.ID,
		Name:         result.Name,
		Pid:          int32(result.PID),
		State:        string(result.State),
		Vcpus:        int32(result.VCPUs),
		MemoryMib:    int32(result.MemoryMiB),
		KernelImage:  result.KernelImage,
		RootDrive:    result.RootDrive,
		UseJailer:    result.UseJailer,
		SocketPath:   result.SocketPath,
		LogPath:      result.LogPath,
		MetricsPath:  result.MetricsPath,
		WorkDir:      result.WorkDir,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339),
		Uptime:       result.Uptime,
		ProcessAlive: result.ProcessAlive,
		SocketExists: result.SocketExists,
	}, nil
}

// Logs streams the process log file for a VM.
func (s *Server) Logs(req *LogsRequest, stream DaemonService_LogsServer) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "VM ID is required")
	}

	reader, err := s.vmManager.ReadLogs(req.Id, req.Follow)
	if err != nil {
		return status.Errorf(codes.NotFound, "read logs: %v", err)
	}
	defer reader.Close()

	buf := make([]byte, 4096)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if err := stream.Send(&LogsResponse{Data: buf[:n]}); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			if !req.Follow {
				return nil
			}
			// In follow mode, sleep and try again
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if readErr != nil {
			return status.Errorf(codes.Internal, "read logs: %v", readErr)
		}
	}
}

// Metrics reads the latest process metrics for a VM.
func (s *Server) Metrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "VM ID is required")
	}

	data, err := s.vmManager.ReadMetrics(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "read metrics: %v", err)
	}

	return &MetricsResponse{Data: data}, nil
}

// Call forwards a raw HTTP request to the VM's Firecracker API socket.
func (s *Server) Call(ctx context.Context, req *CallRequest) (*CallResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "VM ID is required")
	}
	if req.Method == "" {
		return nil, status.Error(codes.InvalidArgument, "HTTP method is required")
	}
	if req.Path == "" {
		return nil, status.Error(codes.InvalidArgument, "API path is required")
	}

	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = io.NopCloser(io.LimitReader(
			io.NopCloser(bytesReader(req.Body)), 10*1024*1024, // 10MB limit
		))
	}

	resp, err := s.vmManager.Call(ctx, req.Id, req.Method, req.Path, bodyReader)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "call API: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "read response: %v", err)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &CallResponse{
		StatusCode: int32(resp.StatusCode),
		Body:       respBody,
		Headers:    headers,
	}, nil
}

// Health returns the daemon's health status.
func (s *Server) Health(ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "ok",
		Version: s.version,
		Uptime:  time.Since(s.startedAt).Round(time.Second).String(),
		VmCount: int32(s.vmManager.Registry.Count()),
	}, nil
}

// --- Interceptors ---

func loggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	clientCN := extractClientCN(ctx)
	if err != nil {
		log.Printf("RPC %s from %s — ERROR (%s): %v", info.FullMethod, clientCN, duration, err)
	} else {
		log.Printf("RPC %s from %s — OK (%s)", info.FullMethod, clientCN, duration)
	}

	return resp, err
}

func loggingStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)
	duration := time.Since(start)

	clientCN := extractClientCN(ss.Context())
	if err != nil {
		log.Printf("Stream %s from %s — ERROR (%s): %v", info.FullMethod, clientCN, duration, err)
	} else {
		log.Printf("Stream %s from %s — OK (%s)", info.FullMethod, clientCN, duration)
	}

	return err
}

// extractClientCN extracts the Common Name from the client's TLS certificate.
func extractClientCN(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "unknown"
	}

	if len(tlsInfo.State.PeerCertificates) > 0 {
		return tlsInfo.State.PeerCertificates[0].Subject.CommonName
	}

	return "unknown"
}

// bytesReader creates an io.Reader from a byte slice.
type bytesReaderImpl struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) io.Reader {
	return &bytesReaderImpl{data: data}
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// mustEmbedUnimplementedDaemonServiceServer is a marker for forward compatibility.
func (s *Server) mustEmbedUnimplementedDaemonServiceServer() {}

// --- TLS Helper for External Use ---

// LoadServerTLS is a convenience wrapper around pki.LoadServerTLSConfig.
func LoadServerTLS(certDir string) (*tls.Config, error) {
	return pki.LoadServerTLSConfig(certDir)
}
