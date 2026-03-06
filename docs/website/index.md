# digital.vasic.memory

A Go module for memory management with Mem0-style capabilities, entity extraction, knowledge graph construction, and memory leak detection.

## Key Features

- **Memory Store** -- CRUD, search, and scoped listing of memory units with metadata and embedding vectors
- **Mem0-Style Manager** -- Importance scoring, exponential time decay, and automatic consolidation of similar memories
- **Entity Extraction** -- Regex-based extraction of entities (emails, URLs, noun phrases) and relations (is_a, has, uses) from text
- **Knowledge Graph** -- In-memory directed graph with BFS shortest path and depth-bounded subgraph extraction
- **Leak Detection** -- Runtime memory monitoring with heap/goroutine profiling and alert callbacks

## Installation

```bash
go get digital.vasic.memory
```

Requires Go 1.24+.

## Package Overview

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `store` | `digital.vasic.memory/pkg/store` | Core `MemoryStore` interface, `Memory` type, and in-memory implementation |
| `mem0` | `digital.vasic.memory/pkg/mem0` | Mem0-style manager with decay, importance, and consolidation |
| `entity` | `digital.vasic.memory/pkg/entity` | Entity and relation extraction via regex patterns |
| `graph` | `digital.vasic.memory/pkg/graph` | In-memory knowledge graph with BFS traversal |
| `memory` | `digital.vasic.memory/pkg/memory` | Leak detector, memory monitor, and profiling utilities |

## Quick Example

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.memory/pkg/mem0"
    "digital.vasic.memory/pkg/store"
)

func main() {
    backend := store.NewInMemoryStore()
    mgr := mem0.NewManager(backend, nil) // nil = default config

    ctx := context.Background()
    mgr.Add(ctx, &store.Memory{
        Content: "Go 1.24 introduced range-over-func iterators",
        Scope:   store.ScopeGlobal,
    })

    results, _ := mgr.Search(ctx, "iterators", nil)
    for _, m := range results {
        fmt.Printf("%.2f  %s\n", m.Score, m.Content)
    }
}
```
