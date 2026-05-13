cbox: Decentralized MicroVM Orchestration Engine

Architectural Specification & Design Document
Status: Living Document - Open for Enhancement

1. Executive Summary

cbox is a decentralized, hyper-isolated orchestration platform designed to replace heavy, container-based orchestration systems (like Kubernetes) with lightweight, highly secure Firecracker microVMs (referred to within the ecosystem as "Pods").

Abandoning the rigid "Master/Worker" node paradigm, cbox leverages peer-to-peer networking (Netbird) and distributed state consensus (dqlite). A "cluster" in cbox is not a collection of physical servers; it is a logical, cryptographically secured mesh network and a shared database quorum. Physical machines can seamlessly participate in multiple isolated clusters simultaneously.

2. Core Principles

True Decentralization: There is no centralized control plane. Any machine running the cboxd daemon can initialize a cluster, join an existing one, and host the distributed state.

Hyper-Isolation (Pods over Containers): Workloads run in hardware-virtualized Firecracker microVMs, sharing no kernel space with the host or each other.

Low Rigidity & High Extensibility: The core engine is exceptionally "dumb" and robust. All advanced logic, custom resources, and operational intelligence are delegated to external, SDK-driven Operators.

Declarative State: Driven by an API Server and independent Reconciliation Loops. Direct imperative commands to the hardware are strictly forbidden.

3. The Architectural Stack (Separation of Concerns)

To maintain security and prevent bloat, the system is strictly divided between the Host Layer (The Muscle) and the Pod Layer (The Brain & Workloads).

3.1 The Host Layer: cboxd (The Daemon)

The cboxd binary is a lightweight, background system service installed on bare-metal servers.

Role: The "Muscle." It interacts directly with the host operating system (/dev/kvm, network interfaces).

Components:

The Reconciler: A continuous background loop that reads the database and uses the Firecracker Go SDK to spawn, rotate, or destroy Pods.

Mesh Engine: An embedded Netbird daemon that manages WireGuard tunnels for local clusters.

Boundary: cboxd does not expose a public REST/gRPC API for cluster management. It contains no high-level business logic. It blindly trusts the cluster's database.

3.2 The Pod Layer: The API Server & Database

When a cluster is initialized, cboxd spawns a dedicated Firecracker Pod to hold the cluster's brain.

Database (dqlite): Distributed SQLite using the Raft consensus algorithm. As new machines join the cluster, their database Pods form a highly available quorum over the Netbird mesh.

The API Server: A stateless Go application sitting in front of dqlite.

Authentication & Authorization: Validates JWTs and Casbin RBAC rules.

Translation: Converts high-level user configurations into low-level Firecracker primitives (vCPU, RAM, rootfs).

Logical Splitting: Internally divided into a Core Router (handling standard Pods) and an Extension Engine (handling Custom Resource Definitions and Operator webhooks).

3.3 The Client Layer: cbox CLI & SDK

A completely stateless binary used by developers and CI/CD pipelines. It connects securely to the API Server Pod over the network to submit declarative YAML/JSON manifests.

4. Component Deep Dive & Data Flow

4.1 The Reconciliation Loop

The Reconciler is the heart of the cboxd daemon.

Watch: The Reconciler maintains a constant read-only connection to the cluster's dqlite database.

Filter: It only looks for lowest-level Pod definitions assigned to its specific host machine.

Compare: It checks the local Firecracker sockets to compare Desired State (DB) vs. Actual State (Hardware).

Execute: If a mismatch exists, it uses the Firecracker Go SDK to correct reality (boot a microVM, allocate RAM, etc.).

4.2 Handling Bootstrapping & Startup Commands

Because Firecracker boots raw Linux environments (not containers), cbox uses MMDS (MicroVM Metadata Service) to bridge the gap.

The Reconciler configures the Firecracker MMDS before booting the Pod.

Upon boot, a tiny initialization script inside the .ext4 rootfs queries the local 169.254.169.254 MMDS endpoint.

It retrieves user-defined startup commands (e.g., curl -O my-app && ./my-app) and executes them, allowing K8s-style dynamic container/binary loading inside the microVM.

5. Extensibility: Custom Operators & CRDs

cbox is designed to be infinitely extensible without modifying the core API Server or Daemon. It achieves this via Custom Operators.

The Concept: An Operator is a standalone application (which can run in its own Cbox Pod) that acts as an intelligent client.

The Flow:

A cluster admin applies a Custom Resource Definition (CRD) to the API Server (e.g., MCPServer).

The Operator uses the cbox Client SDK to open a Watch stream on the API Server for MCPServer objects.

When a user requests an MCPServer, the API Server alerts the Operator.

Operator Autonomy: The Operator executes its own complex logic. It might spy on a GitHub repository (GitOps), provision external cloud resources, or serve its own custom UI dashboard.

Translation: Finally, the Operator translates its findings into standard Cbox Pod blueprints and uses the SDK to write them back to the API Server. The cboxd Daemon then physically boots the Pods.

6. Security & Secret Management

To prevent sensitive data (API Keys, Passwords) from being exposed in plaintext configurations or written to raw .ext4 disk files, cbox implements a Just-In-Time (JIT) secrets pipeline:

Storage: Secrets are submitted to the API Server and encrypted at rest (e.g., AES-256) inside the dqlite database.

Reference: Workload manifests only contain a reference to the secret (e.g., secretKeyRef: my-db-pass).

Delivery: When the cboxd Daemon prepares to boot a Pod, it securely fetches the decrypted secret from the API Server into host RAM.

Injection: The Daemon pushes the secret into the Firecracker MMDS. The Pod boots, retrieves the secret via local HTTP, and exports it as an in-memory Environment Variable.

7. The Two-Tier User Interface Architecture

To maintain true decentralization, the visual management of cbox is split into two tiers:

Tier 1: The Binary UI (Host Level)

Embedded: Compiled directly into the cboxd Go binary using go:embed.

Scope: Limited strictly to the physical host.

Function: A lightweight dashboard accessed via http://<host-ip>:<port>/ui. Used solely for host hardware metrics (CPU/RAM usage), bootstrapping new logical clusters (cluster init), and joining existing meshes.

Tier 2: The Cluster UI (Global Level)

Floating: Packaged as a standard workload inside a Firecracker Pod.

Scope: The entire logical cluster network.

Function: The heavy, feature-rich dashboard for deploying applications, managing RBAC, viewing logs, and visualizing the Netbird mesh.

Resilience: Because it is a Pod, if the underlying physical server dies, the Reconciler will automatically spin up the UI Pod on a surviving machine in the mesh.

8. The Lifecycle Flow (CLI Reference)

cbox daemon start: Runs the cboxd background engine on the host.

cbox cluster init --name prod: The Daemon creates a local namespace folder, initializes a Netbird mesh, and boots the Genesis API/dqlite Pod.

cbox cluster join --name prod --token <jwt>: The Daemon authenticates with a remote peer, joins the Netbird mesh, and boots a follower dqlite Pod to join the Raft quorum.

cbox apply -f app.yaml: The Client SDK sends a declarative manifest to the API Server Pod, which translates it and saves it to the database, instantly triggering local Daemon Reconcilers across the mesh.

End of Document. Architecture is subject to continuous enhancement as core binaries are developed.