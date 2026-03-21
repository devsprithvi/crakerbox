//go:build !linux

package vm

import (
	"os"
	"os/exec"
	"path/filepath"
)

// setProcGroup is a no-op on non-Linux platforms.
func setProcGroup(cmd *exec.Cmd) {
	// Setpgid is only available on Linux.
}

// isProcessAliveOS is a basic check on non-Linux platforms.
func isProcessAliveOS(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On non-Linux, FindProcess always succeeds so this is a best-effort check.
	_ = process
	return true
}

// createFifo on non-Linux creates a regular file as a placeholder.
// Firecracker only runs on Linux; this stub enables cross-compilation.
func createFifo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	os.Remove(path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}
