# Getting Started

## Install

```bash
go get digital.vasic.memory
```

## Store a Memory

The `store` package provides the `MemoryStore` interface and a thread-safe in-memory implementation.

```go
import (
    "context"
    "digital.vasic.memory/pkg/store"
)

s := store.NewInMemoryStore()
ctx := context.Background()

err := s.Add(ctx, &store.Memory{
    Content:  "User prefers dark mode",
    Scope:    store.ScopeUser,
    Metadata: map[string]any{"source": "settings"},
})
```

IDs are auto-generated (UUID v4) when empty. Timestamps are set automatically.

## Search Memories

```go
opts := &store.SearchOptions{
    TopK:     5,
    MinScore: 0.3,
    Scope:    store.ScopeUser,
}
results, err := s.Search(ctx, "dark mode", opts)
```

The in-memory backend uses word-overlap scoring. Results are sorted by score descending.

## Use the Mem0 Manager

The `mem0.Manager` wraps any `MemoryStore` and adds importance scoring, exponential decay, and consolidation.

```go
import "digital.vasic.memory/pkg/mem0"

cfg := mem0.DefaultConfig()
cfg.DecayRate = 0.05          // faster decay
cfg.SimilarityThreshold = 0.8 // stricter consolidation

mgr := mem0.NewManager(s, cfg)
mgr.Add(ctx, &store.Memory{Content: "User likes Go"})
```

Importance is calculated from content length, metadata richness, embedding presence, and scope. Search results have decay applied automatically.

## Consolidate Similar Memories

```go
merged, err := mgr.Consolidate(ctx, store.ScopeUser)
fmt.Printf("Consolidated %d duplicate memories\n", merged)
```

Consolidation uses Jaccard word-overlap similarity. It respects a cooldown interval (`ConsolidationInterval`) to avoid running too frequently.

## Extract Entities

```go
import "digital.vasic.memory/pkg/entity"

ext := entity.NewPatternExtractor()
entities, relations, _ := ext.Extract(
    "Alice uses Go and Bob is a developer",
)
// entities: [{Name:"Alice" Type:"noun_phrase"}, ...]
// relations: [{Subject:"Alice" Predicate:"uses" Object:"Go"}, ...]
```

Add custom patterns with the builder API:

```go
ext.WithEntityPattern("version", "version", `v(\d+\.\d+\.\d+)`)
ext.WithRelationPattern("depends_on", "depends_on", `(\w+)\s+depends on\s+(\w+)`)
```

## Build a Knowledge Graph

```go
import "digital.vasic.memory/pkg/graph"

g := graph.NewInMemoryGraph()
g.AddNode(graph.Node{ID: "go", Type: "language"})
g.AddNode(graph.Node{ID: "alice", Type: "person"})
g.AddEdge(graph.Edge{Source: "alice", Target: "go", Relation: "uses", Weight: 1.0})

path, _ := g.ShortestPath("alice", "go")
// ["alice", "go"]

nodes, edges, _ := g.Subgraph("alice", 2)
```
