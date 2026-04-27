# AGENTS.md - Memory Module Multi-Agent Coordination Guide

## Overview

This document provides guidance for AI agents (Claude Code, Copilot, Cursor, etc.) working on the `digital.vasic.memory` module. It defines responsibilities, boundaries, and coordination protocols to prevent conflicts when multiple agents operate concurrently.

## Module Identity

- **Module path**: `digital.vasic.memory`
- **Language**: Go 1.24+
- **Dependencies**: `github.com/google/uuid`, `github.com/stretchr/testify`
- **Packages**: `pkg/store`, `pkg/mem0`, `pkg/entity`, `pkg/graph`

## Package Ownership Boundaries

### `pkg/store` -- Core Memory Store

- **Scope**: Memory struct, Scope type, MemoryStore interface, SearchOptions, ListOptions, InMemoryStore implementation.
- **Owner concern**: Any agent modifying the `MemoryStore` interface MUST update all implementations (`InMemoryStore`, `mem0.Manager`) and their tests.
- **Thread safety**: `InMemoryStore` uses `sync.RWMutex`. All new methods MUST acquire appropriate locks. Returns copies, never internal pointers.

### `pkg/mem0` -- Mem0-Style Memory Manager

- **Scope**: Config, Manager (wraps `store.MemoryStore`), importance scoring, time-based decay, consolidation.
- **Owner concern**: `Manager` implements `store.MemoryStore`. Any interface changes in `pkg/store` require corresponding changes here.
- **Dependency**: Imports `pkg/store`. Changes to `store.Memory` fields affect importance calculation in `CalculateImportance`.

### `pkg/entity` -- Entity and Relation Extraction

- **Scope**: Entity/Relation types, Extractor interface, PatternExtractor with builder pattern.
- **Owner concern**: Self-contained. No dependencies on other Memory packages. Safe for independent modification.
- **Extension point**: Custom patterns via `WithEntityPattern` and `WithRelationPattern` builders.

### `pkg/graph` -- Knowledge Graph

- **Scope**: Node/Edge types, Graph interface, InMemoryGraph with BFS shortest path and subgraph extraction.
- **Owner concern**: Self-contained. No dependencies on other Memory packages. Safe for independent modification.
- **Thread safety**: `InMemoryGraph` uses `sync.RWMutex`. All new methods MUST acquire appropriate locks.

## Dependency Graph

```
pkg/mem0 --> pkg/store
pkg/entity (independent)
pkg/graph  (independent)
```

`pkg/entity` and `pkg/graph` have zero internal dependencies and can be modified in isolation. `pkg/mem0` depends on `pkg/store`; changes to the store interface or Memory struct propagate to mem0.

## Agent Coordination Rules

### 1. Interface Changes

If you modify `store.MemoryStore`:
- Update `InMemoryStore` in `pkg/store/inmemory.go`
- Update `Manager` in `pkg/mem0/mem0.go`
- Add tests for both implementations
- Verify interface compliance tests still pass (`TestManager_ImplementsMemoryStore`, `TestInMemoryGraph_ImplementsGraph`)

If you modify `entity.Extractor`:
- Update `PatternExtractor` in `pkg/entity/entity.go`
- Verify `TestPatternExtractor_ImplementsExtractor`

If you modify `graph.Graph`:
- Update `InMemoryGraph` in `pkg/graph/graph.go`
- Verify `TestInMemoryGraph_ImplementsGraph`

### 2. Struct Field Changes

Adding fields to `store.Memory`:
- Check JSON tags follow existing `snake_case` convention with `omitempty` where appropriate
- Update `CalculateImportance` in `pkg/mem0/mem0.go` if the field affects scoring
- Update `mergeMemories` in `pkg/mem0/mem0.go` if the field should be merged during consolidation
- Add corresponding test cases

Adding fields to `graph.Node` or `graph.Edge`:
- Self-contained; no cross-package impact
- Ensure JSON tags are consistent

### 3. Concurrency Safety

All four packages are designed for concurrent access:
- `store.InMemoryStore`: `sync.RWMutex` on all operations
- `graph.InMemoryGraph`: `sync.RWMutex` on all operations
- `mem0.Manager`: `sync.RWMutex` for Add and Consolidate; delegates to backend for reads

Rules:
- Read operations use `RLock`/`RUnlock`
- Write operations use `Lock`/`Unlock`
- Never hold a lock while calling an external function that might also lock
- Always return copies of internal data, never pointers to stored objects

### 4. Testing Standards

- **Framework**: `github.com/stretchr/testify` (assert + require)
- **Naming**: `Test<Struct>_<Method>_<Scenario>` (e.g., `TestInMemoryStore_Search_ScoreOrdering`)
- **Style**: Table-driven tests with `tests` slice and `t.Run` subtests
- **Concurrency**: Each package has a `ConcurrentAccess` test; maintain these
- **Interface compliance**: Each implementation has a compile-time check (e.g., `var _ Graph = (*InMemoryGraph)(nil)`)
- **Run all tests**: `go test ./... -count=1 -race`

### 5. Adding New Implementations

To add a new `MemoryStore` implementation (e.g., PostgreSQL, Redis):
1. Create `pkg/store/<name>.go` implementing `MemoryStore`
2. Create `pkg/store/<name>_test.go` with full test coverage
3. Include interface compliance test
4. Include concurrency test
5. Do NOT modify existing implementations

To add a new `Graph` implementation:
1. Create `pkg/graph/<name>.go` implementing `Graph`
2. Follow the same test pattern as `InMemoryGraph`

### 6. File Ownership

| File | Primary Concern | Cross-Package Impact |
|------|----------------|---------------------|
| `pkg/store/store.go` | Interfaces and types | HIGH -- affects mem0 |
| `pkg/store/inmemory.go` | InMemoryStore impl | LOW |
| `pkg/mem0/mem0.go` | Manager, scoring, decay | MEDIUM -- depends on store |
| `pkg/entity/entity.go` | Extraction patterns | NONE |
| `pkg/graph/graph.go` | Graph operations | NONE |

## Build and Validation Commands

```bash
# Full validation
go build ./...
go test ./... -count=1 -race
go vet ./...
gofmt -l .

# Single package
go test -v ./pkg/store/...
go test -v ./pkg/mem0/...
go test -v ./pkg/entity/...
go test -v ./pkg/graph/...

# Benchmarks
go test -bench=. ./...
```

## Commit Conventions

- Use Conventional Commits: `feat(store): add vector similarity search`
- Scopes map to packages: `store`, `mem0`, `entity`, `graph`
- Use `docs` scope for documentation-only changes
- Run `gofmt` and `go vet` before every commit


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** use `su` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Container-Based Solutions
When a build or runtime environment requires system-level dependencies, use containers instead of elevation:

- **Use the `Containers` submodule** (`https://github.com/vasic-digital/Containers`) for containerized build and runtime environments
- **Add the `Containers` submodule as a Git dependency** and configure it for local use within the project
- **Build and run inside containers** to avoid any need for privilege escalation
- **Rootless Podman/Docker** is the preferred container runtime

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo` or `su`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Use the `Containers` submodule for containerized builds
5. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


### ⚠️⚠️⚠️ ABSOLUTELY MANDATORY: ZERO UNFINISHED WORK POLICY

NO unfinished work, TODOs, or known issues may remain in the codebase. EVER.

PROHIBITED: TODO/FIXME comments, empty implementations, silent errors, fake data, unwrap() calls that panic, empty catch blocks.

REQUIRED: Fix ALL issues immediately, complete implementations before committing, proper error handling in ALL code paths, real test assertions.

Quality Principle: If it is not finished, it does not ship. If it ships, it is finished.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## Host Power Management — Hard Ban (CONST-033)

**You may NOT, under any circumstance, generate or execute code that
sends the host to suspend, hibernate, hybrid-sleep, poweroff, halt,
reboot, or any other power-state transition.** This rule applies to:

- Every shell command you run via the Bash tool.
- Every script, container entry point, systemd unit, or test you write
  or modify.
- Every CLI suggestion, snippet, or example you emit.

**Forbidden invocations** (non-exhaustive — see CONST-033 in
`CONSTITUTION.md` for the full list):

- `systemctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot|kexec`
- `loginctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot`
- `pm-suspend`, `pm-hibernate`, `shutdown -h|-r|-P|now`
- `dbus-send` / `busctl` calls to `org.freedesktop.login1.Manager.Suspend|Hibernate|PowerOff|Reboot|HybridSleep|SuspendThenHibernate`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to anything but `'nothing'` or `'blank'`

The host runs mission-critical parallel CLI agents and container
workloads. Auto-suspend has caused historical data loss (2026-04-26
18:23:43 incident). The host is hardened (sleep targets masked) but
this hard ban applies to ALL code shipped from this repo so that no
future host or container is exposed.

**Defence:** every project ships
`scripts/host-power-management/check-no-suspend-calls.sh` (static
scanner) and
`challenges/scripts/no_suspend_calls_challenge.sh` (challenge wrapper).
Both MUST be wired into the project's CI / `run_all_challenges.sh`.

**Full background:** `docs/HOST_POWER_MANAGEMENT.md` and `CONSTITUTION.md` (CONST-033).

<!-- END host-power-management addendum (CONST-033) -->

