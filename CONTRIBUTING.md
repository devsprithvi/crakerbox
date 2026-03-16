# Contributing to Crackerbox

## Git Flow Branching Strategy

We follow **Git Flow**. Here's the workflow:

### Branches

| Branch | Purpose | Merges Into |
|--------|---------|-------------|
| `main` | Production releases only | — |
| `develop` | Integration branch | `main` (via release) |
| `feature/<name>` | New features | `develop` |
| `release/<version>` | Release prep & QA | `main` + `develop` |
| `hotfix/<name>` | Emergency production fixes | `main` + `develop` |

### Workflow

#### Starting a New Feature
```bash
git checkout develop
git pull origin develop
git checkout -b feature/my-feature
# ... work ...
git push -u origin feature/my-feature
# Open PR → develop
```

#### Creating a Release
```bash
git checkout develop
git checkout -b release/0.2.0
# Bump versions, final QA
# Open PR → main
# After merge, tag: git tag v0.2.0
# Also merge back → develop
```

#### Hotfix
```bash
git checkout main
git checkout -b hotfix/critical-fix
# Fix the issue
# Open PR → main AND develop
```

### Commit Messages

Use conventional commits:

```
feat: add VM creation command
fix: correct KVM detection on ARM
docs: update install instructions
chore: update CI workflow
refactor: extract VM manager into package
```

## Submodules

Each app under `apps/` is intended to be a Git submodule. When cloning:

```bash
git clone --recurse-submodules git@github.com:devsprithvi/crakerbox.git
```

To update submodules:

```bash
git submodule update --remote --merge
```
