# Memory

Generic, reusable Go module for memory management with Mem0-style capabilities, entity extraction, and knowledge graph construction.

**Module**: `digital.vasic.memory`

## Packages

- **pkg/store** -- Core memory store interfaces (`MemoryStore`), types (`Memory`, `Scope`, `SearchOptions`, `ListOptions`), and thread-safe in-memory implementation.
- **pkg/mem0** -- Mem0-style memory manager with automatic importance scoring, time-based decay, and memory consolidation (merging similar memories).
- **pkg/entity** -- Pattern-based entity and relation extraction from text. Ships with default patterns for emails, URLs, capitalized phrases, and common relations (is_a, has, uses). Extensible with custom patterns.
- **pkg/graph** -- In-memory knowledge graph with directed weighted edges, BFS shortest path, neighbor traversal, and subgraph extraction.

## Quick Start

```go
import (
    "digital.vasic.memory/pkg/store"
    "digital.vasic.memory/pkg/mem0"
)

backend := store.NewInMemoryStore()
manager := mem0.NewManager(backend, nil)

_ = manager.Add(ctx, &store.Memory{
    Content: "User prefers concise responses",
    Scope:   store.ScopeUser,
})

results, _ := manager.Search(ctx, "preferences", nil)
```

## Testing

```bash
go test ./... -count=1 -race
```
