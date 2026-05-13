Architecture Specification: Custom Firecracker Control Plane

1. Core Philosophy

This architecture strips away enterprise orchestration bloat (e.g., complex API aggregators, admission webhooks) in favor of predictability, anatomy order, and strict symmetry.
The system operates on a pure client-server model where every component—whether a high-level scheduler or a low-level worker node—is simply an authenticated client talking to a central brain.

2. The Core Components (The Brain)

The entire control plane is defined by just two tightly coupled components. There is no other source of truth.

The API Server: The strict gatekeeper. It handles API Key authentication, globalized schema validation (for predictable resource definitions), and exposes both declarative state endpoints and imperative command tunnels.

The Database: The single source of truth holding both intent and reality. High-level resources (e.g., App Configurations) and low-level resources (e.g., Firecracker MicroVMs) live side-by-side to maintain absolute architectural symmetry.

3. Binaries and Cluster Design

The system is deployed using two purpose-built binaries:

The Super Binary (Manager): The top-level tool used by administrators to instantiate and manage multiple clusters.

The Daemon Binary (Worker): Installed on the bare-metal servers. It acts as the lowest-level hardware operator, managing the Firecracker processes.

Cluster High-Availability (HA)

When the Super Binary creates a cluster, it self-hosts the control plane:

The API Server and Database are spun up inside isolated microVMs for that specific cluster.

To ensure high availability and failure tolerance, the database is replicated across multiple microVMs connecting back to the master.

4. Networking Infrastructure

To maintain a clean connection between the control plane and distributed worker nodes without relying on heavy external proxies, the system utilizes NetBird.

NetBird provides a zero-configuration, WireGuard-based peer-to-peer overlay network.

It ensures secure, direct, proxy-free communication channels between the API Server and all Daemon binaries across any physical location.

5. The Twin-Track Communication Model

To achieve the "Best of Both Worlds" (reliable state reconciliation + full API potential for immediate actions), communication is split into two tracks:

Track A: Declarative State (The Database)

Used for the baseline truth (Spec vs Status).

Flow: The user submits a configuration -> API Server validates -> Saved to Database. Worker nodes reconcile reality to match the database and update their Status continuously.

Track B: Imperative Commands (NATS JetStream)

Used for point-in-time, full-potential Firecracker actions (e.g., Snapshots, Memory Hotplugging).

Flow: The user requests an immediate action -> API Server pushes the command JSON directly into a NATS JetStream message queue -> Worker node executes the Firecracker API -> Upon success, the API Server updates the Database state.

6. The Operator Symmetry

In this architecture, the concept of an "Operator" is beautifully simplified. Every active component is essentially an operator:

High-Level Operators: External clients using SDKs that read high-level custom resources from the API Server and write low-level Firecracker MicroVM resources back to it.

Schedulers: External clients that read unassigned MicroVM resources and update them with a chosen Worker Node ID.

The Worker Node (Daemon): The ultimate, lowest-level operator. It reads assigned MicroVM resources from the API Server, executes the Linux binaries to make them real, and updates the Status in the database.

Symmetry is maintained because all operators, from abstraction to bare metal, are simply authenticated loops acting on the exact same API Server.