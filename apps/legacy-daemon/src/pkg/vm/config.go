// Package vm provides Firecracker microVM lifecycle management using the
// Firecracker Go SDK. It handles VM spawning, termination, inspection, and
// bridging to the Firecracker API socket.
package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the resolved configuration for a Firecracker microVM.
// It can be loaded from a static JSON file or constructed dynamically from CLI flags.
type Config struct {
	// Identity
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	// Machine
	VCPUs     int  `json:"vcpus"`
	MemoryMiB int  `json:"memory_mib"`
	HtEnabled bool `json:"ht_enabled"`
	SMT       bool `json:"smt"`

	// Boot
	KernelImagePath string `json:"kernel_image_path"`
	InitrdPath      string `json:"initrd_path,omitempty"`
	KernelArgs      string `json:"kernel_args,omitempty"`

	// Storage
	RootDrivePath string `json:"root_drive_path"`
	RootDriveID   string `json:"root_drive_id,omitempty"`
	ReadOnly      bool   `json:"read_only"`

	// Network
	NetworkInterfaces []NetworkInterface `json:"network_interfaces,omitempty"`

	// Paths (resolved at runtime by the daemon)
	SocketPath  string `json:"socket_path,omitempty"`
	LogPath     string `json:"log_path,omitempty"`
	MetricsPath string `json:"metrics_path,omitempty"`
	WorkDir     string `json:"work_dir,omitempty"`

	// Jailer
	UseJailer   bool   `json:"use_jailer"`
	JailerBin   string `json:"jailer_bin,omitempty"`
	JailerUID   int    `json:"jailer_uid,omitempty"`
	JailerGID   int    `json:"jailer_gid,omitempty"`
	ChrootBase  string `json:"chroot_base,omitempty"`

	// Metadata (MMDS)
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Raw: when the user provides a full config-file for Firecracker's --config-file
	RawConfigPath string `json:"raw_config_path,omitempty"`
}

// NetworkInterface represents a VM's network interface configuration.
type NetworkInterface struct {
	ID              string `json:"iface_id"`
	HostDevName     string `json:"host_dev_name"`
	GuestMAC        string `json:"guest_mac,omitempty"`
	AllowMMDSAccess bool   `json:"allow_mmds_requests"`
}

// DefaultConfig returns sensible defaults for a microVM.
func DefaultConfig() Config {
	return Config{
		VCPUs:       1,
		MemoryMiB:   128,
		HtEnabled:   false,
		SMT:         false,
		KernelArgs:  "console=ttyS0 reboot=k panic=1 pci=off",
		RootDriveID: "rootfs",
		ReadOnly:    false,
	}
}

// LoadConfigFromFile reads and parses a VM configuration JSON file.
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks the configuration for required fields and consistency.
func (c *Config) Validate() error {
	if c.KernelImagePath == "" && c.RawConfigPath == "" {
		return fmt.Errorf("kernel_image_path is required (or provide raw_config_path)")
	}
	if c.RootDrivePath == "" && c.RawConfigPath == "" {
		return fmt.Errorf("root_drive_path is required (or provide raw_config_path)")
	}
	if c.VCPUs < 1 {
		return fmt.Errorf("vcpus must be at least 1")
	}
	if c.MemoryMiB < 1 {
		return fmt.Errorf("memory_mib must be at least 1")
	}
	if c.KernelImagePath != "" {
		if _, err := os.Stat(c.KernelImagePath); os.IsNotExist(err) {
			return fmt.Errorf("kernel image not found: %s", c.KernelImagePath)
		}
	}
	if c.RootDrivePath != "" {
		if _, err := os.Stat(c.RootDrivePath); os.IsNotExist(err) {
			return fmt.Errorf("root drive not found: %s", c.RootDrivePath)
		}
	}
	if c.UseJailer {
		if c.JailerBin == "" {
			c.JailerBin = "/usr/bin/jailer"
		}
		if _, err := os.Stat(c.JailerBin); os.IsNotExist(err) {
			return fmt.Errorf("jailer binary not found: %s", c.JailerBin)
		}
	}
	return nil
}

// ResolveWorkspace sets up all runtime paths for this VM within the given base directory.
func (c *Config) ResolveWorkspace(baseDir string) error {
	if c.ID == "" {
		return fmt.Errorf("VM ID must be set before resolving workspace")
	}

	c.WorkDir = filepath.Join(baseDir, c.ID)
	if err := os.MkdirAll(c.WorkDir, 0755); err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}

	if c.SocketPath == "" {
		c.SocketPath = filepath.Join(c.WorkDir, "firecracker.sock")
	}
	if c.LogPath == "" {
		c.LogPath = filepath.Join(c.WorkDir, "firecracker.log")
	}
	if c.MetricsPath == "" {
		c.MetricsPath = filepath.Join(c.WorkDir, "firecracker-metrics.fifo")
	}

	return nil
}
