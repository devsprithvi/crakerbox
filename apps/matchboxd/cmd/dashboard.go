package cmd

import (
	"context"
	"fmt"
)

// DashboardApp is the Wails application struct for the UI
type DashboardApp struct {
	ctx context.Context
}

// NewDashboardApp creates a new DashboardApp instance
func NewDashboardApp() *DashboardApp {
	return &DashboardApp{}
}

// startup is called when the Wails app starts
func (a *DashboardApp) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetStatus returns the current daemon status
func (a *DashboardApp) GetStatus() string {
	return "running"
}

// GetVersion returns the daemon version
func (a *DashboardApp) GetVersion() string {
	return fmt.Sprintf("%s (commit: %s)", Version, GitCommit)
}
