# Lesson 2: Intelligent Memory with Mem0

## Objectives

- Wrap a `MemoryStore` with the Mem0 `Manager`
- Understand importance scoring and exponential decay
- Consolidate duplicate memories

## Concepts

### The Decorator Pattern

`mem0.Manager` implements `store.MemoryStore` and wraps any backend store. It adds three behaviors without modifying the backend:

1. **Importance scoring** on `Add` and `Update`
2. **Exponential decay** on `Search` and `List`
3. **Consolidation** of similar memories

### Configuration

```go
cfg := &mem0.Config{
    DefaultScope:          store.ScopeUser,
    MaxMemories:           10000,
    ConsolidationInterval: 5 * time.Minute,
    DecayRate:             0.01,
    SimilarityThreshold:   0.7,
}
```

### Importance Scoring

`CalculateImportance` assigns a base score of 0.5, then adds bonuses:

- +0.1 for content > 100 characters
- +0.1 for content > 500 characters
- +0.1 for having metadata
- +0.1 for having embeddings
- +0.1 for global scope

The result is capped at 1.0.

### Exponential Decay

`ApplyDecay` reduces scores over time: `score * exp(-rate * hours)`. With rate=0.01, a memory loses about 1% of its score per hour.

## Code Walkthrough

### Setting up the Manager

```go
backend := store.NewInMemoryStore()
mgr := mem0.NewManager(backend, mem0.DefaultConfig())
```

### Adding memories with auto-scoring

```go
mgr.Add(ctx, &store.Memory{
    Content: "PostgreSQL supports JSONB columns for flexible schemas",
    Metadata: map[string]any{"topic": "database"},
})
// Score is automatically calculated (0.7: base + content>100 + metadata)
```

### Search with decay

```go
results, _ := mgr.Search(ctx, "JSONB", nil)
// Scores are adjusted by time decay before being returned
```

### Consolidating duplicates

```go
merged, err := mgr.Consolidate(ctx, store.ScopeUser)
```

Consolidation computes Jaccard similarity between memory pairs. When two memories exceed the threshold, the shorter is absorbed: the longer content wins, metadata is merged, the higher score is kept, and the earlier creation time is preserved.

## Summary

The Mem0 Manager adds intelligent behavior on top of any memory store. Importance scoring surfaces relevant memories, decay lets old memories fade, and consolidation prevents duplication.
