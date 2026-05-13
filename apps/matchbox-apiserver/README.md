# Matchbox API Server

**Cluster-internal component** — automatically provisioned when a cluster is created via the Matchbox Manager.

## Role
- Primary access point for all cluster operations
- Runs on the master node
- Communicates with the Matchbox Database for persistent state

## Access
- The Matchbox Controller connects to this component
- No direct external access — all interactions go through the controller

## Lifecycle
- Automatically created when a cluster is provisioned
- Automatically destroyed when a cluster is deleted
