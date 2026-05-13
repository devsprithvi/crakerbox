# Matchbox Database

**Cluster-internal component** — automatically provisioned alongside the API server when a cluster is created.

## Role
- Persistent storage for cluster state
- Exclusively used by the API server — no other component touches it directly

## Lifecycle
- Automatically created when a cluster is provisioned
- Automatically destroyed when a cluster is deleted

## Access Policy
- **Only the API server may access the database** — this is a strict architectural boundary.
