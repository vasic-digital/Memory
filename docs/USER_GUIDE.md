# User Guide

## Introduction

`digital.vasic.memory` is a generic, reusable Go module for memory management with Mem0-style capabilities, entity extraction, and knowledge graph construction. It provides four packages that can be used independently or composed together.

## Installation

```bash
go get digital.vasic.memory@latest
```

**Requirements**: Go 1.24 or later.

## Package Overview

| Package | Import Path | Purpose |
|---------|-------------|---------|
| store | `digital.vasic.memory/pkg/store` | Core memory CRUD, search, scoping, in-memory backend |
| mem0 | `digital.vasic.memory/pkg/mem0` | Mem0-style manager with decay, importance, consolidation |
| entity | `digital.vasic.memory/pkg/entity` | Pattern-based entity and relation extraction |
| graph | `digital.vasic.memory/pkg/graph` | In-memory knowledge graph with BFS traversal |

## Quick Start

### Basic Memory Store

The `store` package provides the `MemoryStore` interface and a thread-safe in-memory implementation.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "digital.vasic.memory/pkg/store"
)

func main() {
    ctx := context.Background()
    s := store.NewInMemoryStore()

    // Add a memory (ID auto-generated if empty)
    mem := &store.Memory{
        Content:  "User prefers dark mode and concise responses",
        Scope:    store.ScopeUser,
        Metadata: map[string]any{"source": "settings"},
    }
    if err := s.Add(ctx, mem); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created memory:", mem.ID)

    // Search for memories
    results, err := s.Search(ctx, "dark mode", nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range results {
        fmt.Printf("  [%.2f] %s\n", r.Score, r.Content)
    }
}
```

### Mem0-Style Memory Manager

The `mem0` package wraps any `MemoryStore` with automatic importance scoring, time-based exponential decay, and memory consolidation.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "digital.vasic.memory/pkg/mem0"
    "digital.vasic.memory/pkg/store"
)

func main() {
    ctx := context.Background()

    // Create a backend store
    backend := store.NewInMemoryStore()

    // Wrap with Mem0-style manager (nil config uses defaults)
    manager := mem0.NewManager(backend, nil)

    // Add memories -- importance is auto-calculated
    _ = manager.Add(ctx, &store.Memory{
        Content: "User prefers concise responses",
        Scope:   store.ScopeUser,
    })
    _ = manager.Add(ctx, &store.Memory{
        Content: "User prefers brief answers",
        Scope:   store.ScopeUser,
    })

    // Search with automatic decay applied to scores
    results, err := manager.Search(ctx, "preferences", nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range results {
        fmt.Printf("  [%.4f] %s\n", r.Score, r.Content)
    }

    // Consolidate similar memories
    merged, err := manager.Consolidate(ctx, store.ScopeUser)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Consolidated %d memories\n", merged)
}
```

### Custom Configuration

```go
cfg := &mem0.Config{
    DefaultScope:          store.ScopeSession,
    MaxMemories:           5000,
    ConsolidationInterval: 10 * time.Minute,
    DecayRate:             0.05,          // Faster decay
    SimilarityThreshold:   0.8,           // Stricter consolidation
}
manager := mem0.NewManager(backend, cfg)
```

**Config fields**:

| Field | Default | Description |
|-------|---------|-------------|
| `DefaultScope` | `ScopeUser` | Scope assigned to memories without an explicit scope |
| `MaxMemories` | 10000 | Maximum memories to retain per scope (0 = unlimited) |
| `ConsolidationInterval` | 5 minutes | Minimum time between consolidation runs |
| `DecayRate` | 0.01 | Exponential decay rate (0 disables decay) |
| `SimilarityThreshold` | 0.7 | Minimum Jaccard similarity for consolidation merging |

## Memory Scopes

Scopes control memory visibility boundaries:

```go
store.ScopeUser          // "user" -- visible to a specific user
store.ScopeSession       // "session" -- visible within a session
store.ScopeConversation  // "conversation" -- visible within a conversation
store.ScopeGlobal        // "global" -- visible to all users
```

Scopes are used for filtering in `Search`, `List`, and `Consolidate` operations:

```go
// Search only within session scope
results, _ := s.Search(ctx, "query", &store.SearchOptions{
    Scope: store.ScopeSession,
})

// List all global memories
globals, _ := s.List(ctx, store.ScopeGlobal, nil)

// Consolidate user-scoped memories
merged, _ := manager.Consolidate(ctx, store.ScopeUser)
```

## Search and Filtering

### Search Options

```go
opts := &store.SearchOptions{
    TopK:     5,                      // Return at most 5 results
    MinScore: 0.5,                    // Minimum match score threshold
    Scope:    store.ScopeUser,        // Filter by scope
    TimeRange: &store.TimeRange{      // Filter by creation time
        Start: time.Now().Add(-24 * time.Hour),
        End:   time.Now(),
    },
    Filter: map[string]any{           // Metadata filters (reserved)
        "category": "preferences",
    },
}
results, err := s.Search(ctx, "dark mode preferences", opts)
```

The in-memory implementation uses word-overlap scoring: the fraction of query words found in the memory content. Results are sorted by descending score.

### List Options

```go
opts := &store.ListOptions{
    Offset:  20,             // Skip first 20 results
    Limit:   10,             // Return at most 10
    OrderBy: "updated_at",   // Sort by: "created_at" (default), "updated_at", "score"
    Scope:   store.ScopeUser,
}
results, err := s.List(ctx, store.ScopeUser, opts)
```

Pass an empty scope (`""`) to `List` to retrieve memories across all scopes.

## Importance Scoring

The `mem0.CalculateImportance` function assigns an importance score (0.0 to 1.0) based on:

| Factor | Boost | Condition |
|--------|-------|-----------|
| Base | +0.50 | Always |
| Content length | +0.10 | Content > 100 characters |
| Content length | +0.10 | Content > 500 characters |
| Metadata | +0.10 | Has at least one metadata entry |
| Embedding | +0.10 | Has an embedding vector |
| Global scope | +0.10 | Scope is `ScopeGlobal` |

The score is capped at 1.0. When adding memories via `Manager.Add`, importance is auto-calculated if the memory's score is zero. Existing non-zero scores are preserved.

## Time-Based Decay

Decay reduces memory relevance over time using exponential decay:

```
decayed_score = score * exp(-rate * hours_since_creation)
```

- `rate = 0.01` (default): after 24 hours, score retains ~78.7% of original value
- `rate = 0.05`: after 24 hours, score retains ~30.1% of original value
- `rate = 0`: decay is disabled

Decay is applied on read (`Search` and `List`), not on write. The stored score remains unchanged.

## Memory Consolidation

Consolidation merges similar memories to reduce redundancy:

```go
merged, err := manager.Consolidate(ctx, store.ScopeUser)
```

The algorithm:
1. Lists all memories in the given scope
2. Computes pairwise Jaccard word-overlap similarity
3. When similarity exceeds `SimilarityThreshold` (default 0.7), merges the pair:
   - Keeps the longer content
   - Merges metadata (target keys take precedence)
   - Keeps the higher score
   - Uses the earlier creation time
4. Deletes the absorbed memory from the backend
5. Respects `ConsolidationInterval` cooldown between runs

## Entity Extraction

The `entity` package extracts named entities and relations from text using regex patterns.

```go
package main

import (
    "fmt"
    "digital.vasic.memory/pkg/entity"
)

func main() {
    pe := entity.NewPatternExtractor()

    text := "Alice Smith works at https://acme.com and her email is alice@acme.com. Go is a programming language."

    entities, relations, _ := pe.Extract(text)

    fmt.Println("Entities:")
    for _, e := range entities {
        fmt.Printf("  %s (%s)\n", e.Name, e.Type)
    }

    fmt.Println("Relations:")
    for _, r := range relations {
        fmt.Printf("  %s -[%s]-> %s\n", r.Subject, r.Predicate, r.Object)
    }
}
```

### Default Patterns

**Entity patterns** (built-in):

| Name | Type | Matches |
|------|------|---------|
| email | `email` | Email addresses |
| url | `url` | HTTP/HTTPS URLs |
| capitalized_phrase | `noun_phrase` | Multi-word capitalized phrases (e.g., "John Smith") |

**Relation patterns** (built-in):

| Name | Predicate | Example Match |
|------|-----------|---------------|
| is_a | `is_a` | "Go is a language" |
| has | `has` | "system has components" |
| uses | `uses` | "project uses Docker" |

### Custom Patterns

Use the builder pattern to add domain-specific patterns:

```go
pe := entity.NewPatternExtractor().
    WithEntityPattern("version", "version", `v(\d+\.\d+\.\d+)`).
    WithEntityPattern("ticket", "ticket", `(JIRA-\d+)`).
    WithRelationPattern("depends", "depends_on", `(\w+)\s+depends\s+on\s+(\w+)`)

entities, relations, _ := pe.Extract("Module v1.2.3 depends on core (JIRA-1234)")
```

Entity patterns must have one capture group for the entity name. Relation patterns must have two capture groups: subject and object.

## Knowledge Graph

The `graph` package provides an in-memory directed graph with weighted edges.

```go
package main

import (
    "fmt"
    "digital.vasic.memory/pkg/graph"
)

func main() {
    g := graph.NewInMemoryGraph()

    // Add nodes
    _ = g.AddNode(graph.Node{ID: "go", Type: "language", Properties: map[string]any{"year": 2009}})
    _ = g.AddNode(graph.Node{ID: "goroutine", Type: "feature"})
    _ = g.AddNode(graph.Node{ID: "channel", Type: "feature"})
    _ = g.AddNode(graph.Node{ID: "concurrency", Type: "concept"})

    // Add edges
    _ = g.AddEdge(graph.Edge{Source: "go", Target: "goroutine", Relation: "has", Weight: 1.0})
    _ = g.AddEdge(graph.Edge{Source: "go", Target: "channel", Relation: "has", Weight: 1.0})
    _ = g.AddEdge(graph.Edge{Source: "goroutine", Target: "concurrency", Relation: "enables", Weight: 0.9})
    _ = g.AddEdge(graph.Edge{Source: "channel", Target: "concurrency", Relation: "enables", Weight: 0.8})

    // Get neighbors
    neighbors, _ := g.GetNeighbors("go")
    fmt.Println("Go features:")
    for _, n := range neighbors {
        fmt.Printf("  %s (%s)\n", n.ID, n.Type)
    }

    // Shortest path (BFS by hop count)
    path, _ := g.ShortestPath("go", "concurrency")
    fmt.Println("Path to concurrency:", path)

    // Extract subgraph within 1 hop
    nodes, edges, _ := g.Subgraph("go", 1)
    fmt.Printf("Subgraph: %d nodes, %d edges\n", len(nodes), len(edges))
}
```

### Graph Operations

| Method | Description |
|--------|-------------|
| `AddNode(node)` | Add a node (error if ID is empty or duplicate) |
| `AddEdge(edge)` | Add a directed edge (both nodes must exist) |
| `GetNode(id)` | Retrieve a node by ID |
| `GetNeighbors(id)` | Get all nodes directly reachable from a node (deduplicated) |
| `ShortestPath(from, to)` | BFS shortest path by hop count |
| `Subgraph(startID, maxDepth)` | All nodes and edges within maxDepth hops |
| `Nodes()` | All nodes in the graph |
| `Edges()` | All edges in the graph |

## Composing Packages Together

A common pattern is to extract entities from text, build a knowledge graph, and store memory entries:

```go
ctx := context.Background()

// 1. Set up memory manager
backend := store.NewInMemoryStore()
manager := mem0.NewManager(backend, nil)

// 2. Set up entity extraction
extractor := entity.NewPatternExtractor()

// 3. Set up knowledge graph
kg := graph.NewInMemoryGraph()

// 4. Process text
text := "Alice Smith uses Go. Go is a programming language."

// Store as memory
_ = manager.Add(ctx, &store.Memory{
    Content:  text,
    Scope:    store.ScopeUser,
    Metadata: map[string]any{"source": "conversation"},
})

// Extract entities and relations
entities, relations, _ := extractor.Extract(text)

// Populate knowledge graph
for _, e := range entities {
    _ = kg.AddNode(graph.Node{ID: e.Name, Type: e.Type})
}
for _, r := range relations {
    // Ensure nodes exist before adding edges
    _ = kg.AddNode(graph.Node{ID: r.Subject, Type: "extracted"})
    _ = kg.AddNode(graph.Node{ID: r.Object, Type: "extracted"})
    _ = kg.AddEdge(graph.Edge{
        Source:   r.Subject,
        Target:   r.Object,
        Relation: r.Predicate,
        Weight:   1.0,
    })
}
```

## Thread Safety

All implementations are safe for concurrent use:

- `store.InMemoryStore` uses `sync.RWMutex` and returns copies of stored data
- `graph.InMemoryGraph` uses `sync.RWMutex` and returns copies of stored data
- `mem0.Manager` uses `sync.RWMutex` for write operations and delegates reads to the backend

## Embedding Support

The `Memory` struct includes an `Embedding` field (`[]float32`) for vector embeddings. The in-memory store uses word-overlap scoring for search; for semantic similarity search, provide a custom `MemoryStore` implementation that uses the embedding vectors.

```go
mem := &store.Memory{
    Content:   "User prefers dark mode",
    Embedding: []float32{0.12, -0.34, 0.56, ...},  // From your embedding provider
    Scope:     store.ScopeUser,
}
```

Memories with embeddings receive a +0.10 importance boost from `CalculateImportance`.

## Testing

Run the full test suite:

```bash
go test ./... -count=1 -race
```

Run tests for a specific package:

```bash
go test -v ./pkg/store/...
go test -v ./pkg/mem0/...
go test -v ./pkg/entity/...
go test -v ./pkg/graph/...
```

Run benchmarks:

```bash
go test -bench=. ./...
```
