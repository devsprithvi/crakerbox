package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"crackerboxd/pkg/api"
	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	spawnConfigPath   string
	spawnName         string
	spawnVCPUs        int32
	spawnMemoryMiB    int32
	spawnKernelImage  string
	spawnRootDrive    string
	spawnKernelArgs   string
	spawnUseJailer    bool
	spawnDaemonAddr   string
	spawnCertDir      string
)

var spawnCmd = &cobra.Command{
	Use:   "spawn",
	Short: "Start a new Firecracker microVM",
	Long: `Spawns a new Firecracker microVM process.

Configuration can be provided via a JSON file (--config) or through
dynamic CLI flags. Dynamic flags override values from the config file.

Examples:
  crackerboxd spawn --config /path/to/vm.json
  crackerboxd spawn --kernel-image /boot/vmlinux --root-drive /images/rootfs.ext4
  crackerboxd spawn --kernel-image /boot/vmlinux --root-drive /images/rootfs.ext4 --vcpus 2 --memory 256
  crackerboxd spawn --config /path/to/vm.json --use-jailer`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to daemon via mTLS
		conn, err := dialDaemon(spawnDaemonAddr, spawnCertDir)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := api.NewDaemonServiceClient(conn)

		req := &api.SpawnRequest{
			ConfigPath:      spawnConfigPath,
			Name:            spawnName,
			Vcpus:           spawnVCPUs,
			MemoryMib:       spawnMemoryMiB,
			KernelImagePath: spawnKernelImage,
			RootDrivePath:   spawnRootDrive,
			KernelArgs:      spawnKernelArgs,
			UseJailer:       spawnUseJailer,
		}

		resp, err := client.Spawn(context.Background(), req)
		if err != nil {
			return fmt.Errorf("spawn failed: %w", err)
		}

		// Pretty-print the result
		out, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(out))

		return nil
	},
}

func init() {
	spawnCmd.Flags().StringVar(&spawnConfigPath, "config", "", "Path to VM configuration JSON file")
	spawnCmd.Flags().StringVar(&spawnName, "name", "", "VM name")
	spawnCmd.Flags().Int32Var(&spawnVCPUs, "vcpus", 0, "Number of vCPUs")
	spawnCmd.Flags().Int32Var(&spawnMemoryMiB, "memory", 0, "Memory in MiB")
	spawnCmd.Flags().StringVar(&spawnKernelImage, "kernel-image", "", "Path to kernel image")
	spawnCmd.Flags().StringVar(&spawnRootDrive, "root-drive", "", "Path to root filesystem")
	spawnCmd.Flags().StringVar(&spawnKernelArgs, "kernel-args", "", "Kernel boot arguments")
	spawnCmd.Flags().BoolVar(&spawnUseJailer, "use-jailer", false, "Run VM inside the Firecracker jailer")
	spawnCmd.Flags().StringVar(&spawnDaemonAddr, "daemon", "localhost:8090", "Daemon gRPC address")
	spawnCmd.Flags().StringVar(&spawnCertDir, "cert-dir", pki.DefaultCertDir, "mTLS certificate directory")

	rootCmd.AddCommand(spawnCmd)
}

// dialDaemon creates an mTLS-secured gRPC connection to the daemon.
func dialDaemon(addr, certDir string) (*grpc.ClientConn, error) {
	tlsConfig, err := pki.LoadClientTLSConfig(certDir)
	if err != nil {
		return nil, fmt.Errorf("load client TLS: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("connect to daemon at %s: %w", addr, err)
	}

	return conn, nil
}
