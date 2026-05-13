//go:build linux

package vm

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// setProcGroup sets process group isolation on the command (Linux only).
func setProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// isProcessAlive checks if a process with the given PID is still running.
func isProcessAliveOS(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Linux, sending signal 0 checks if process exists without killing it.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// createFifo creates a named pipe (FIFO) at the given path (Linux only).
func createFifo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	os.Remove(path)
	return syscall.Mkfifo(path, 0644)
}
