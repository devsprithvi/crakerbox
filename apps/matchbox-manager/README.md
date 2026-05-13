# Matchbox Manager

The **cluster lifecycle management** binary. Responsible for creating, deleting, and managing clusters.

## Role
- Creates clusters (provisions API server + database automatically)
- Deletes clusters
- Manages cluster lifecycle and state

## Interface
- CLI (primary)

## Cluster Provisioning
When a cluster is created via the manager, the following components are automatically provisioned on the master node:
1. **API Server** — The primary access point for cluster operations
2. **Database** — Internal storage used exclusively by the API server
