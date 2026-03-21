// Package api provides the Go types for the gRPC DaemonService.
//
// NOTE: These types mirror the protobuf definitions in proto/daemon.proto.
// In a production setup, you would run `protoc` to generate these.
// For now, we define them manually to avoid a protoc dependency during development.
//
// To generate from proto:
//   protoc --go_out=. --go-grpc_out=. pkg/api/proto/daemon.proto
package api

import (
	"context"

	"google.golang.org/grpc"
)

// --- Request/Response Types ---

type SpawnRequest struct {
	ConfigPath        string                  `json:"config_path"`
	Vcpus             int32                   `json:"vcpus"`
	MemoryMib         int32                   `json:"memory_mib"`
	KernelImagePath   string                  `json:"kernel_image_path"`
	RootDrivePath     string                  `json:"root_drive_path"`
	KernelArgs        string                  `json:"kernel_args"`
	Name              string                  `json:"name"`
	UseJailer         bool                    `json:"use_jailer"`
	NetworkInterfaces []*NetworkInterfaceSpec  `json:"network_interfaces"`
}

type NetworkInterfaceSpec struct {
	IfaceId           string `json:"iface_id"`
	HostDevName       string `json:"host_dev_name"`
	GuestMac          string `json:"guest_mac"`
	AllowMmdsRequests bool   `json:"allow_mmds_requests"`
}

type SpawnResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Pid        int32  `json:"pid"`
	SocketPath string `json:"socket_path"`
	State      string `json:"state"`
}

type TerminateRequest struct {
	Id string `json:"id"`
}

type TerminateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListRequest struct {
	StateFilter string `json:"state_filter"`
}

type ListResponse struct {
	Vms   []*VMEntry `json:"vms"`
	Total int32      `json:"total"`
}

type VMEntry struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Pid        int32  `json:"pid"`
	State      string `json:"state"`
	SocketPath string `json:"socket_path"`
	CreatedAt  string `json:"created_at"`
	Uptime     string `json:"uptime"`
}

type InspectRequest struct {
	Id string `json:"id"`
}

type InspectResponse struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Pid          int32  `json:"pid"`
	State        string `json:"state"`
	Vcpus        int32  `json:"vcpus"`
	MemoryMib    int32  `json:"memory_mib"`
	KernelImage  string `json:"kernel_image"`
	RootDrive    string `json:"root_drive"`
	UseJailer    bool   `json:"use_jailer"`
	SocketPath   string `json:"socket_path"`
	LogPath      string `json:"log_path"`
	MetricsPath  string `json:"metrics_path"`
	WorkDir      string `json:"work_dir"`
	CreatedAt    string `json:"created_at"`
	Uptime       string `json:"uptime"`
	ProcessAlive bool   `json:"process_alive"`
	SocketExists bool   `json:"socket_exists"`
}

type LogsRequest struct {
	Id     string `json:"id"`
	Follow bool   `json:"follow"`
}

type LogsResponse struct {
	Data []byte `json:"data"`
}

type MetricsRequest struct {
	Id string `json:"id"`
}

type MetricsResponse struct {
	Data []byte `json:"data"`
}

type CallRequest struct {
	Id     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   []byte `json:"body"`
}

type CallResponse struct {
	StatusCode int32             `json:"status_code"`
	Body       []byte            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

type HealthRequest struct{}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	VmCount int32  `json:"vm_count"`
}

// --- Service Interface ---

// DaemonServiceServer is the interface that the server must implement.
type DaemonServiceServer interface {
	Spawn(context.Context, *SpawnRequest) (*SpawnResponse, error)
	Terminate(context.Context, *TerminateRequest) (*TerminateResponse, error)
	List(context.Context, *ListRequest) (*ListResponse, error)
	Inspect(context.Context, *InspectRequest) (*InspectResponse, error)
	Logs(*LogsRequest, DaemonService_LogsServer) error
	Metrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
	Call(context.Context, *CallRequest) (*CallResponse, error)
	Health(context.Context, *HealthRequest) (*HealthResponse, error)
	mustEmbedUnimplementedDaemonServiceServer()
}

// DaemonService_LogsServer is the server-side streaming interface for Logs.
type DaemonService_LogsServer interface {
	Send(*LogsResponse) error
	grpc.ServerStream
}

type daemonServiceLogsServer struct {
	grpc.ServerStream
}

func (x *daemonServiceLogsServer) Send(m *LogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

// DaemonServiceClient is the client interface for the daemon service.
type DaemonServiceClient interface {
	Spawn(ctx context.Context, in *SpawnRequest, opts ...grpc.CallOption) (*SpawnResponse, error)
	Terminate(ctx context.Context, in *TerminateRequest, opts ...grpc.CallOption) (*TerminateResponse, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	Inspect(ctx context.Context, in *InspectRequest, opts ...grpc.CallOption) (*InspectResponse, error)
	Logs(ctx context.Context, in *LogsRequest, opts ...grpc.CallOption) (DaemonService_LogsClient, error)
	Metrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error)
	Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error)
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error)
}

// DaemonService_LogsClient is the client-side streaming interface for Logs.
type DaemonService_LogsClient interface {
	Recv() (*LogsResponse, error)
	grpc.ClientStream
}

type daemonServiceLogsClient struct {
	grpc.ClientStream
}

func (x *daemonServiceLogsClient) Recv() (*LogsResponse, error) {
	m := new(LogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// --- Service Registration ---

const daemonServiceName = "crackerbox.daemon.v1.DaemonService"

// daemonServiceDesc is the gRPC service descriptor for DaemonService.
var daemonServiceDesc = grpc.ServiceDesc{
	ServiceName: daemonServiceName,
	HandlerType: (*DaemonServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Spawn", Handler: _DaemonService_Spawn_Handler},
		{MethodName: "Terminate", Handler: _DaemonService_Terminate_Handler},
		{MethodName: "List", Handler: _DaemonService_List_Handler},
		{MethodName: "Inspect", Handler: _DaemonService_Inspect_Handler},
		{MethodName: "Metrics", Handler: _DaemonService_Metrics_Handler},
		{MethodName: "Call", Handler: _DaemonService_Call_Handler},
		{MethodName: "Health", Handler: _DaemonService_Health_Handler},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Logs",
			Handler:       _DaemonService_Logs_Handler,
			ServerStreams:  true,
		},
	},
	Metadata: "proto/daemon.proto",
}

// RegisterDaemonServiceServer registers the service with the gRPC server.
func RegisterDaemonServiceServer(s *grpc.Server, srv DaemonServiceServer) {
	s.RegisterService(&daemonServiceDesc, srv)
}

// --- Method Handlers ---

func _DaemonService_Spawn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SpawnRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Spawn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Spawn"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Spawn(ctx, req.(*SpawnRequest))
	})
}

func _DaemonService_Terminate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TerminateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Terminate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Terminate"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Terminate(ctx, req.(*TerminateRequest))
	})
}

func _DaemonService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/List"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).List(ctx, req.(*ListRequest))
	})
}

func _DaemonService_Inspect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InspectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Inspect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Inspect"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Inspect(ctx, req.(*InspectRequest))
	})
}

func _DaemonService_Metrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Metrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Metrics"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Metrics(ctx, req.(*MetricsRequest))
	})
}

func _DaemonService_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Call"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Call(ctx, req.(*CallRequest))
	})
}

func _DaemonService_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServiceServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + daemonServiceName + "/Health"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServiceServer).Health(ctx, req.(*HealthRequest))
	})
}

func _DaemonService_Logs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServiceServer).Logs(m, &daemonServiceLogsServer{stream})
}

// --- Client Implementation ---

type daemonServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewDaemonServiceClient creates a new client for the DaemonService.
func NewDaemonServiceClient(cc grpc.ClientConnInterface) DaemonServiceClient {
	return &daemonServiceClient{cc}
}

func (c *daemonServiceClient) Spawn(ctx context.Context, in *SpawnRequest, opts ...grpc.CallOption) (*SpawnResponse, error) {
	out := new(SpawnResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Spawn", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) Terminate(ctx context.Context, in *TerminateRequest, opts ...grpc.CallOption) (*TerminateResponse, error) {
	out := new(TerminateResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Terminate", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/List", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) Inspect(ctx context.Context, in *InspectRequest, opts ...grpc.CallOption) (*InspectResponse, error) {
	out := new(InspectResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Inspect", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) Logs(ctx context.Context, in *LogsRequest, opts ...grpc.CallOption) (DaemonService_LogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &daemonServiceDesc.Streams[0], "/"+daemonServiceName+"/Logs", opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonServiceLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

func (c *daemonServiceClient) Metrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error) {
	out := new(MetricsResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Metrics", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error) {
	out := new(CallResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Call", in, out, opts...)
	return out, err
}

func (c *daemonServiceClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error) {
	out := new(HealthResponse)
	err := c.cc.Invoke(ctx, "/"+daemonServiceName+"/Health", in, out, opts...)
	return out, err
}
