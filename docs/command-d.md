USAGE:
  crackerbox [command] [flags]

--- CLI MANAGEMENT COMMANDS ---
  version     Print the current version of the CLI and daemon
  update      Self-update the CLI binary
  doctor      Check system readiness (Linux OS, KVM access, dependencies)
  help        Display detailed help for any command
  config      View or modify the global CLI settings

Commands:
  # PROCESS MANAGEMENT (The Daemon's Core Job)
  spawn --config <path>   # Starts the Firecracker process.
  terminate <id>          # Kills the Firecracker process.
  list                    # Shows all running Firecracker processes.
  inspect <id>            # Shows host-level details (PID, paths, uptime).

  # TELEMETRY (Reading Files from Disk)
  logs <id>               # Streams the process log file.
  metrics <id>            # Reads the latest process metrics.

  # THE BRIDGE (User-Initiated API Access)
  call <id> <method> <p>  # Forwards a request (e.g., GET /mmds) to the socket.



COMMAND: spawn

1. INPUT
   - STATIC: User-provided JSON Configuration File.
   - DYNAMIC: CLI Flags (CPU, RAM, Network, etc.) + Global Defaults.

2. PROCESSING (The Resolver)
   - CONFIGURATION: Resolve the final Machine Specification (SDK Object).
     - IF Static: Pass-through the provided JSON.
     - IF Dynamic: Merge flags into the default template.
   - PROVISIONING: Execute Host-Level Preparation.
     - Dynamically allocate Host Resources (ID, Network Interfaces, Workspace).
     - Map Host-Side artifacts to the SDK Configuration object.
   - INITIATION: 
     - Hand the resolved Configuration to the Go SDK.
     - Start the Firecracker process (utilizing the --config-file exception).

3. OUTPUT
   - STATE: A running Firecracker Process (OS PID).
   - REGISTRY: A tracked entry in the Daemon (Mapping ID to PID and Host Paths).
   - RESULT: Success/Failure status returned to the caller.

   # --- MANAGEMENT COMMANDS ---

COMMAND: terminate
- INPUT: Unique VM Identifier (ID).
- ACTION: Stops the Firecracker process and deallocates assigned host resources (TAP, Workspace).
- OUTPUT: Confirmation of process removal and host cleanup.

COMMAND: list
- INPUT: Optional status filters.
- ACTION: Scans the active registry and host process table.
- OUTPUT: A collection of managed IDs, their PIDs, and their API Socket paths.

COMMAND: inspect
- INPUT: Unique VM Identifier (ID).
- ACTION: Aggregates host-level metadata including uptime, configuration sources, and file paths.
- OUTPUT: A detailed manifest of the host-side state for that specific process.

# --- DATA COMMANDS ---

COMMAND: logs
- INPUT: Unique VM Identifier (ID).
- ACTION: Accesses and reads the associated log file from the host filesystem.
- OUTPUT: A stream of internal process events.

COMMAND: metrics
- INPUT: Unique VM Identifier (ID).
- ACTION: Accesses and reads the associated metrics file from the host filesystem.
- OUTPUT: A point-in-time snapshot of process performance data (JSON).

# --- INTERACTION COMMANDS ---

COMMAND: call
- INPUT: ID, HTTP Method, API Endpoint, and optional Request Body.
- ACTION: Forwards the raw request to the specific Unix Domain Socket.
- OUTPUT: The raw response returned directly from the Firecracker API.