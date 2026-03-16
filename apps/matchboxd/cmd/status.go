package cmd

import (
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// --- Bubble Tea TUI Model ---

type statusModel struct {
	kvmAvailable bool
	platform     string
	version      string
}

func initialStatusModel() statusModel {
	kvmAvail := false
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/dev/kvm"); err == nil {
			kvmAvail = true
		}
	}

	return statusModel{
		kvmAvailable: kvmAvail,
		platform:     runtime.GOOS + "/" + runtime.GOARCH,
		version:      "0.1.0-dev",
	}
}

func (m statusModel) Init() tea.Cmd {
	return nil
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m statusModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6600")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	kvmStatus := "❌ Not available"
	if m.kvmAvailable {
		kvmStatus = "✅ Available"
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6600")).
		Padding(1, 2).
		MarginTop(1)

	content := titleStyle.Render("🔥 Matchbox Daemon Status") + "\n\n" +
		labelStyle.Render("Version:") + valueStyle.Render(m.version) + "\n" +
		labelStyle.Render("Platform:") + valueStyle.Render(m.platform) + "\n" +
		labelStyle.Render("KVM:") + valueStyle.Render(kvmStatus) + "\n" +
		labelStyle.Render("Daemon:") + valueStyle.Render("● Not Running") + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("Press q to quit")

	return boxStyle.Render(content)
}

// --- Cobra Command ---

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon and system status",
	Long:  `Displays the current status of the matchboxd daemon and system capabilities.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialStatusModel())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running status TUI:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
