USAGE:
  microd [command] [flags]

GLOBAL FLAGS:
  -c, --config      Path to daemon config file (Default: /etc/microd/config.yaml)
      --log-level   Set logging verbosity: debug, info, warn, error (Default: info)
  -h, --help        Show help for any command

AUTHENTICATION COMMANDS:
  login             Authenticate and register this host with the Central Manager.
                    (Saves the Manager URI and NetBird Setup Key to the local SQLite DB)
                    Flags:
                      --token     The registration token provided by the Manager
                      --manager   The URL of the Central Manager API

  logout            Disconnect from the mesh, deregister from the Manager, and clear auth state.

DAEMON LIFECYCLE COMMANDS (Host Level):
  server run        Run the daemon in the foreground (blocks terminal, logs to stdout).
                    Flags:
                      -p, --port  Override default API listening port
                      -H, --host  Override default bind address

  server start      Start the daemon intelligently in the background as a root process.
                    Flags:
                      -p, --port  Override default API listening port
                      -H, --host  Override default bind address

  server stop       Stop the background daemon gracefully (drains pending VM tasks).
                    Flags:
                      -f, --force    Hard kill (SIGKILL) immediately
                      -t, --timeout  Seconds to wait before forcing (Default: 30s)

  server reload     Reload config.yaml without dropping the running process (SIGHUP).

  server status     Check if the daemon is running, its PID, and NetBird connection state.
                    Flags:
                      --format    Output format (table, json, yaml. Default: table)

  doctor            Run pre-flight checks on the host OS (/dev/kvm, Firecracker binary, SQLite access).

  update            Pull the latest daemon binary from the Manager and restart safely.

  version           Print the daemon binary version and compiled Go version.

MICRO-VM CORE COMMANDS (Data Plane):
  spawn             Create and boot a new MicroVM, attaching it to the mesh network.
                    Flags:
                      --id        Unique name/ID for the VM (e.g., web-01)
                      --config    Path to a specific VM YAML config (overrides daemon defaults)
                      --vcpu      Number of virtual CPUs (overrides config)
                      --mem       Memory in MB (overrides config)
                      --kernel    Path to kernel binary (overrides config)
                      --rootfs    Path to root filesystem (overrides config)

  terminate         Gracefully shutdown a MicroVM and destroy its TAP interface.
                    Usage: microd terminate <vm-id>
                    Flags:
                      -f, --force Send hard SIGKILL to the Firecracker process

OBSERVABILITY & DEBUGGING COMMANDS:
  list              List all MicroVMs currently managed by this daemon (Queries SQLite).
                    Flags:
                      --limit     Max number of VMs to show (Default: 50)
                      --status    Filter by status (running, stopped, crashed)
                      --filter    Regex/wildcard to filter by VM ID (e.g., "web-*")

  inspect           Output detailed JSON state of a specific MicroVM (PID, IPs, MAC, uptime).
                    Usage: microd inspect <vm-id>

  logs              Stream the console output of the guest OS inside the MicroVM.
                    Usage: microd logs <vm-id>
                    Flags:
                      -f, --follow  Keep streaming logs live
                      --tail        Number of lines to show (Default: 50)

  metrics           Print the internal Firecracker performance metrics (CPU faults, API errors).
                    Usage: microd metrics <vm-id>

  api               Send raw HTTP requests directly to a specific Firecracker Unix socket.
                    Usage: microd api <vm-id> [GET|PUT|PATCH] <endpoint> [body]
                    Example: microd api web-01 GET /machine-config