# 🔥 Crackerbox

> A lightweight control plane for [Firecracker](https://github.com/firecracker-microvm/firecracker) microVMs.

## Repository Structure

This repository uses **Git Submodules** — each app lives in its own repository and is linked here.

```
crackerbox/
├── apps/
│   ├── web/            # Landing page & install script (Cloudflare Pages)
│   ├── crackerboxd/    # Daemon binary (Go + Cobra + Wails)
│   └── cli/            # CLI package (Go, go install-able)
├── .github/
│   └── workflows/      # CI/CD pipelines
├── docs/               # Project documentation
└── README.md
```

## Quick Start

### Install (Linux)

```bash
curl -sL https://get.crackerbox.dev/install.sh | sudo bash
```

### Development

```bash
# Web (landing page)
cd apps/web && bun install && bun run dev

# Daemon
cd apps/crackerboxd && wails dev

# CLI (once implemented)
go install github.com/devsprithvi/crakerbox/apps/cli@latest
```

## Branching Strategy (Git Flow)

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready releases |
| `develop` | Integration branch for features |
| `feature/*` | New features (branch from `develop`) |
| `release/*` | Release prep (branch from `develop`) |
| `hotfix/*` | Emergency fixes (branch from `main`) |

## Apps

| App | Description | Deployment |
|-----|-------------|------------|
| **web** | Landing page + install script | Cloudflare Pages |
| **crackerboxd** | Daemon binary with TUI/GUI | GitHub Releases |
| **cli** | CLI tool (`go install`-able) | Go module registry |

## License

MIT
