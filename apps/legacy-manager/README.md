# Crackerbox Manager

The cluster orchestration layer for the Crackerbox platform.

## Architecture

```
crackerbox-manager (this)
├── manages multiple crackerboxd nodes
├── schedules VM workloads across the fleet
├── monitors node health
└── provides unified cluster API
```

## Relationship

| Component | Role |
|-----------|------|
| `crackerboxd` | **Node-level** daemon — manages Firecracker VMs on a single host |
| `crackerbox-manager` | **Cluster-level** orchestrator — manages multiple crackerboxd nodes |

## Commands

```bash
crackerbox-manager serve        # Start the orchestrator API
crackerbox-manager status       # Show cluster status
crackerbox-manager nodes        # List registered nodes
crackerbox-manager version      # Print version
```

## Build

```bash
go build -o crackerbox-manager .

# With version info
go build -ldflags "-X cmd.Version=0.1.0 -X cmd.GitCommit=$(git rev-parse --short HEAD) -X cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o crackerbox-manager .
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/healthz` | GET | Health check |
| `/api/v1/nodes` | GET | List registered nodes |
| `/api/v1/cluster` | GET | Cluster overview |

## Configuration

Default listen address: `:9090`

```bash
crackerbox-manager serve --listen :9090
```
