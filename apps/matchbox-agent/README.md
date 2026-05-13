# Matchbox Agent

The **worker node** binary. Runs on each node in the cluster and executes assigned workloads.

## Role
- Receives instructions from the cluster (via the API server)
- Executes and manages workloads on the local node
- Reports status back to the cluster

## Interface
- CLI (primary)
