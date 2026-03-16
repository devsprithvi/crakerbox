# Crackerbox CLI

The command-line interface for the Crackerbox platform.

## Install

```bash
go install github.com/devsprithvi/crakerbox/apps/cli@latest
```

This installs the `cli` binary. You can rename it:

```bash
mv $(go env GOPATH)/bin/cli $(go env GOPATH)/bin/crackerbox
```

> **Note:** Proper binary naming will be handled in a future release using
> Go build flags or a Makefile.

## Usage

```bash
crackerbox version       # Print version info
crackerbox --help        # Show all commands
```

## Development

```bash
cd apps/cli
go run . version
go build -o crackerbox .
```

## Planned Commands

These are scaffolded but not yet implemented:

| Command | Description |
|---------|-------------|
| `init` | Initialize a new workspace |
| `status` | Show cluster status |
| `vm list` | List running microVMs |
| `vm create` | Create a new microVM |
| `vm stop` | Stop a running microVM |
| `config` | Manage configuration |
