# 🔥 Crackerbox

> A lightweight control plane for [Firecracker](https://github.com/firecracker-microvm/firecracker) microVMs.

## Architecture

```
                  ┌─────────────────────────┐
                  │   crackerbox-manager     │  ← Cluster Orchestrator (:9090)
                  │   Fleet API, Scheduling  │     Built-in CLI
                  └────────┬────────────────┘
                           │
              ┌────────────┼────────────────┐
              │            │                │
        ┌─────┴─────┐ ┌───┴───────┐ ┌──────┴────┐
        │crackerboxd│ │crackerboxd│ │crackerboxd │  ← Node Daemons (:8090)
        │  Node 1   │ │  Node 2   │ │  Node N    │     Built-in CLI + TUI
        └───────────┘ └───────────┘ └────────────┘
              │            │                │
        ┌─────┴─────┐ ┌───┴───────┐ ┌──────┴────┐
        │Firecracker│ │Firecracker│ │Firecracker │  ← microVMs
        │  microVMs │ │  microVMs │ │  microVMs  │
        └───────────┘ └───────────┘ └────────────┘
```

Two binaries, each with a built-in CLI:

- **`crackerboxd`** — Node daemon. Manages Firecracker VMs on a single host. Includes TUI dashboard and GUI via Wails.
- **`crackerbox-manager`** — Cluster orchestrator. Manages multiple crackerboxd nodes, schedules workloads, monitors health.

## Repository Structure

```
crackerbox/
├── apps/
│   ├── web/              # Landing page & install script (Cloudflare Pages)
│   ├── matchboxd/        # Daemon binary (Go + Cobra + Wails)
│   └── manager/          # Cluster orchestrator (Go + Cobra)
├── .github/
│   └── workflows/
│       ├── build.yml     # CI — build & test all apps
│       ├── release.yml   # CD — cross-compile & publish GitHub Releases
│       └── deploy-web.yml# CD — deploy website to Cloudflare Pages
├── docs/                 # Project documentation
├── Makefile              # Unified build system
└── README.md
```

## Quick Start

### Install (Linux)

```bash
curl -sL https://get.crackerbox.dev/install.sh | sudo bash
```

### Development

```bash
# Build everything
make build

# Individual binaries
make build-daemon     # → bin/crackerboxd
make build-manager    # → bin/crackerbox-manager

# Run tests
make test

# Web (landing page)
cd apps/web && bun install && bun run dev
```

## Binaries

| App | Binary | Role | Port |
|-----|--------|------|------|
| **matchboxd** | `crackerboxd` | Node-level daemon with CLI + TUI — manages Firecracker VMs on a single host | `:8090` |
| **manager** | `crackerbox-manager` | Cluster-level orchestrator with CLI — manages multiple nodes | `:9090` |
| **web** | — | Landing page + install script | Cloudflare Pages |

## CI/CD

| Workflow | Trigger | Action |
|----------|---------|--------|
| `build.yml` | Push / PR to `main`, `develop` | Build & test all Go apps |
| `release.yml` | Push tag `v*` | Cross-compile linux/amd64+arm64, create GitHub Release |
| `deploy-web.yml` | Push to `develop` (apps/web changes) | Deploy to Cloudflare Pages |

### Creating a Release

```bash
git tag v0.1.0
git push origin v0.1.0
# → GitHub Actions builds binaries and creates a release
```

## Branching Strategy (Git Flow)

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready releases |
| `develop` | Integration branch for features |
| `feature/*` | New features (branch from `develop`) |
| `release/*` | Release prep (branch from `develop`) |
| `hotfix/*` | Emergency fixes (branch from `main`) |

## License

MIT
