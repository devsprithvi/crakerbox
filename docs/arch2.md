cbox: Decentralized MicroVM Orchestration Engine

Architectural Specification & Design Document
Status: Living Document - Core Architecture Established

1. Executive Summary

cbox is a hyper-isolated, completely decentralized orchestration platform designed to replace bloated, container-based orchestration systems (like Kubernetes). Workloads are executed in hardware-virtualized Firecracker microVMs (Pods) rather than shared-kernel containers.

Abandoning the legacy "Master/Worker" node paradigm, cbox has no master nodes. Every machine running the cboxd daemon is an equal peer. The system leverages peer-to-peer networking via embedded Netbird and distributed state consensus via dqlite (Raft).

2. Redefining "The Cluster"

In legacy systems, a "Cluster" is a physical group of servers tied to a central control plane. In cbox, a Cluster (or "Cell") is purely a logical construct.

It is defined as a cryptographic boundary (a Netbird Group) and a shared database quorum (dqlite).

A single physical bare-metal server can seamlessly participate in multiple isolated clusters simultaneously without the workloads ever being able to route to one another.

3. The Architectural Stack (Separation of Concerns)

To maintain absolute security and zero-bloat, the system is strictly divided into three distinct layers.

3.1 The Client Layer: cbox CLI & Pkl

A stateless binary used by developers and CI/CD pipelines.

Pkl-First Configuration: YAML fatigue is eliminated by using Apple's Pkl configuration language. Users write type-safe, programmable .pkl files.

Compilation: When a user runs cbox apply -f app.pkl, the CLI evaluates the Pkl file locally, compiles it down to a deterministic JSON manifest, and pushes it to the API Server. The core engine never has to parse complex logic or templates.

3.2 The Host Layer: cboxd (The Daemon)

The cboxd binary is a lightweight, background system service installed on bare-metal servers. It acts strictly as the "Muscle" and contains zero high-level business logic.

The Reconciler: A continuous background loop that reads the distributed database. It compares the Desired State (DB) against the Actual State (Hardware) and uses the Firecracker Go SDK to spawn, rotate, or destroy microVMs.

The Mesh Engine: Embeds the Netbird Go SDK. It manages WireGuard tunnels to join the physical host to the cluster's secure cryptographic group (Layer 3 routing).

3.3 The Control Plane: The API Server & Database

The brain of the cluster does not run on the host OS; it runs inside its own dedicated Firecracker Pods.

State (dqlite): Distributed SQLite using the Raft consensus algorithm. As new machines join the cluster, their daemon can spin up follower dqlite Pods to join the highly available quorum.

The API Server: A stateless Go application sitting in front of dqlite that receives compiled JSON manifests from the CLI, validates RBAC, and writes the desired state to the database.

4. Workloads & Advanced Capabilities

Unlike Docker containers, Firecracker requires a raw Linux kernel (vmlinux) and a root filesystem (.ext4) to boot. cbox embraces this to provide advanced workload capabilities:

Custom Kernel Registry: Power users can strip down kernels to the absolute minimum (e.g., using Unikraft) for 5-millisecond boot times and pull them directly from the cbox Kernel Registry.

Docker-to-Ext4 Translation Pipeline: To support existing OCI/Docker images, cbox provides a translation pipeline that pulls Docker layers and flattens them into a bootable .ext4 block device, seamlessly converting legacy containers into hyper-isolated microVMs.

5. Extensibility: Operators & CRDs

cbox is designed to be infinitely extensible without modifying the core API Server or Daemon.

The Concept: All intelligent logic is delegated to external "Operators." An Operator is a standalone application running in its own Pod.

The Flow: 1. A cluster admin applies a Custom Resource Definition (CRD) using Pkl imports from the cbox Operator Registry.
2. The custom Operator watches the API Server for these specific resources.
3. When an event occurs, the Operator executes its custom logic (e.g., dynamically provisioning cloud storage or acting as a GitOps agent).
4. The Operator translates its findings into standard lowest-level cbox Pod blueprints and writes them back to the API Server. The "dumb" cboxd daemon then physically boots the resulting Pods.

6. The Two-Phase "Join" Process

Joining a cluster requires both network connectivity and state awareness. When a new physical node joins the ecosystem, it executes a two-phase handshake:

Layer 3 (The Network Handshake): cboxd uses a specific Setup Key (cbox cluster join --token <key>). The embedded Netbird SDK authenticates and joins the isolated cryptographic group, obtaining a mesh IP.

Layer 7 (The Control Plane Registration): The daemon uses internal mesh DNS to locate the API Server Pod. It sends a secure JWT detailing its available hardware (CPU/RAM). The API Server registers the node in the dqlite database, making it officially available for the Reconciler to schedule workloads.

7. The Lifecycle Flow (CLI Reference)

Host Management (The Muscle)

cbox daemon start - Runs the background engine on the bare-metal host.

cbox cluster init --name prod - The daemon establishes a new Netbird group, initializes the mesh, and boots the Genesis API/dqlite Pod.

cbox cluster join --token <jwt> - The daemon executes the two-phase handshake to join an existing logical cluster's mesh and database pool.

Workload Management (The Brain)

cbox apply -f app.pkl - The stateless client compiles the Pkl configuration into JSON, sending it over the mesh to the API Server Pod to declare a desired state. The local Reconcilers across the mesh instantly react and boot the Firecracker Pods.