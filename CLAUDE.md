# CLAUDE.md - Memory Module

## Overview

`digital.vasic.memory` is a generic, reusable Go module for memory management with Mem0-style capabilities, entity extraction, and knowledge graph construction.

**Module**: `digital.vasic.memory` (Go 1.24+)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
go test ./... -short              # Unit tests only
go test -bench=. ./...            # Benchmarks
```

## Code Style

- Standard Go conventions, `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line separated)
- Line length <= 100 chars
- Naming: `camelCase` private, `PascalCase` exported, acronyms all-caps
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`
- Tests: table-driven, `testify`, naming `Test<Struct>_<Method>_<Scenario>`

## Package Structure

| Package | Purpose |
|---------|---------|
| `pkg/store` | Core memory store interfaces, types, and in-memory implementation |
| `pkg/mem0` | Mem0-style memory management with consolidation, decay, importance |
| `pkg/entity` | Entity and relation extraction using regex patterns |
| `pkg/graph` | In-memory knowledge graph with BFS shortest path and subgraph |

## Key Interfaces

- `store.MemoryStore` -- Memory CRUD, search, and list operations
- `entity.Extractor` -- Entity and relation extraction from text
- `graph.Graph` -- Knowledge graph with nodes, edges, traversal, shortest path

## Design Patterns

- **Strategy**: MemoryStore (in-memory, extensible to PostgreSQL/Redis/vector DBs)
- **Decorator**: Manager wraps MemoryStore adding decay, importance, consolidation
- **Factory**: `NewInMemoryStore()`, `NewManager()`, `NewPatternExtractor()`
- **Builder**: `PatternExtractor.WithEntityPattern().WithRelationPattern()`

## Commit Style

Conventional Commits: `feat(store): add vector similarity search`


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**

