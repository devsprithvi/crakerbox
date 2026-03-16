# Crackerboxd

The Crackerbox daemon — a Firecracker microVM control plane binary.

## Features

- **CLI** — Powered by [Cobra](https://github.com/spf13/cobra)
- **TUI** — Terminal dashboard via [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **GUI** — Native desktop UI via [Wails](https://wails.io)

## Commands

```bash
crackerboxd serve     # Start the daemon
crackerboxd status    # System status TUI
crackerboxd ui        # Launch desktop GUI
crackerboxd version   # Print version info
```

## Build

```bash
# CLI only
go build -o crackerboxd .

# With Wails GUI
wails build
```
