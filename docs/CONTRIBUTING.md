# Contributing

Thank you for your interest in contributing to `digital.vasic.memory`. This document covers the development workflow, coding standards, and submission process.

## Prerequisites

- **Go 1.24+** (module uses `go 1.24.0` in `go.mod`)
- **Git** with SSH access configured
- Familiarity with Go conventions and the project's architecture (see `docs/ARCHITECTURE.md`)

## Getting Started

1. Clone the repository (SSH only):

```bash
git clone <ssh-url> Memory
cd Memory
```

2. Verify the build and tests pass:

```bash
go build ./...
go test ./... -count=1 -race
go vet ./...
```

## Development Workflow

### Branch Naming

Use conventional prefixes:

| Prefix | Use Case |
|--------|----------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `chore/` | Maintenance, dependency updates |
| `docs/` | Documentation changes |
| `refactor/` | Code restructuring |
| `test/` | Test additions or improvements |

Example: `feat/postgres-store`, `fix/decay-negative-hours`, `test/concurrent-consolidation`

### Commit Messages

Follow Conventional Commits:

```
<type>(<scope>): <description>
```

- **Types**: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`
- **Scopes**: `store`, `mem0`, `entity`, `graph`, or omit for cross-cutting changes
- **Description**: Imperative mood, lowercase, no period

Examples:

```
feat(store): add vector similarity search method
fix(mem0): handle nil metadata in consolidation merge
test(graph): add benchmark for BFS shortest path
docs: update API reference for SearchOptions
```

### Pre-Commit Checklist

Before committing, always run:

```bash
gofmt -l .          # Should produce no output
go vet ./...        # Should produce no warnings
go build ./...      # Should compile cleanly
go test ./... -count=1 -race    # All tests must pass
```

## Code Style

### General

- Follow standard Go conventions per [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` formatting (no alternatives)
- Line length: 100 characters or less for readability
- Use `goimports` for import ordering

### Import Groups

Separate with blank lines: stdlib, third-party, internal.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"

    "digital.vasic.memory/pkg/store"
)
```

### Naming

| Kind | Convention | Example |
|------|-----------|---------|
| Private | `camelCase` | `adjEntry`, `extractEntities` |
| Exported | `PascalCase` | `MemoryStore`, `NewManager` |
| Constants | `PascalCase` or `UPPER_SNAKE_CASE` | `ScopeUser`, `ScopeGlobal` |
| Acronyms | All caps | `ID`, `URL`, `BFS` |
| Receivers | 1--2 letters | `s` for store, `m` for manager, `g` for graph, `pe` for extractor |

### Error Handling

- Always check errors
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Return errors rather than panicking (exception: `regexp.MustCompile` in pattern initialization)
- Use `defer` for cleanup

### Interfaces

- Keep interfaces small and focused
- Define interfaces in the package that uses them, or alongside the primary type
- Accept interfaces, return concrete types
- Verify implementation at compile time: `var _ Interface = (*Impl)(nil)`

## Testing Standards

### Requirements

- Every exported function and method must have test coverage
- Table-driven tests using `testify` (assert + require)
- Test naming: `Test<Struct>_<Method>_<Scenario>`
- Concurrency tests for all thread-safe types
- Interface compliance tests for all implementations

### Test Structure

```go
func TestInMemoryStore_Search_WithMinScore(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        minScore float64
        expected int
    }{
        {"HighThreshold", "Go", 1.0, 0},
        {"LowThreshold", "Go", 0.0, 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := setupStore(t)  // helper
            results, err := s.Search(ctx, tt.query, &store.SearchOptions{
                MinScore: tt.minScore,
            })
            require.NoError(t, err)
            assert.Len(t, results, tt.expected)
        })
    }
}
```

### Running Tests

```bash
# All tests with race detection
go test ./... -count=1 -race

# Single package
go test -v ./pkg/store/...

# Single test
go test -v -run TestManager_Consolidate_MergesSimilar ./pkg/mem0/...

# Benchmarks
go test -bench=. ./...

# Short mode (unit tests only)
go test ./... -short
```

## Adding New Features

### New MemoryStore Implementation

1. Create `pkg/store/<name>.go` implementing all 6 methods of `MemoryStore`
2. Create `pkg/store/<name>_test.go` with:
   - Full CRUD tests
   - Search with all option combinations
   - List with pagination and ordering
   - Concurrency safety test
   - Interface compliance: `var _ MemoryStore = (*YourStore)(nil)`
3. Verify it works with `mem0.Manager`:
   ```go
   manager := mem0.NewManager(yourStore, nil)
   ```

### New Graph Implementation

1. Create `pkg/graph/<name>.go` implementing all 8 methods of `Graph`
2. Create `pkg/graph/<name>_test.go` with full coverage
3. Include interface compliance: `var _ Graph = (*YourGraph)(nil)`

### New Entity Patterns

For built-in patterns, modify `defaultEntityPatterns()` or `defaultRelationPatterns()` in `pkg/entity/entity.go`. Add corresponding tests.

For user-facing patterns, use the builder API (no code changes needed):
```go
pe.WithEntityPattern("name", "type", `regex`)
```

### New Package

If adding an entirely new package:
1. Place it under `pkg/<name>/`
2. Add comprehensive tests in `pkg/<name>/<name>_test.go`
3. Update `docs/API_REFERENCE.md`
4. Update `docs/ARCHITECTURE.md`
5. Update `CLAUDE.md` and `AGENTS.md`

## Documentation

- Update `docs/API_REFERENCE.md` when adding or changing exported APIs
- Update `docs/USER_GUIDE.md` when adding user-facing features
- Update `docs/ARCHITECTURE.md` for design changes
- Add an entry to `docs/CHANGELOG.md` under `[Unreleased]`
- Update `CLAUDE.md` for agent-facing changes

## Pull Request Process

1. Create a branch from `main` with conventional naming
2. Make changes following the code style and testing standards
3. Run the full pre-commit checklist
4. Write a clear PR description explaining the "why"
5. Reference any related issues
6. Ensure CI passes (build, test, vet, fmt)
