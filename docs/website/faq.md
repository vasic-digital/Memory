# FAQ

## How does search scoring work in the in-memory store?

The `InMemoryStore` uses word-overlap scoring: each query word is checked for containment in the memory content (case-insensitive). The score is `matchingWords / totalQueryWords`, ranging from 0.0 to 1.0. For production use with vector similarity, implement a custom `MemoryStore` backed by a vector database.

## What does the Mem0 Manager's decay do?

The Manager applies exponential time decay to search and list results: `score * exp(-rate * hours)`. This means older memories naturally rank lower over time. Set `DecayRate` to 0 to disable decay entirely.

## How does consolidation decide which memories to merge?

Consolidation computes Jaccard similarity (intersection / union of word sets) between all memory pairs in a scope. When similarity exceeds `SimilarityThreshold` (default 0.7), the shorter memory is absorbed into the longer one. Metadata is merged (existing keys are not overwritten), and the higher score is kept.

## Can I add custom entity patterns?

Yes. Use the builder methods on `PatternExtractor`:

```go
ext := entity.NewPatternExtractor().
    WithEntityPattern("ip", "ip_address", `(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`).
    WithRelationPattern("owns", "owns", `(\w+)\s+owns\s+(\w+)`)
```

Patterns must have at least one capture group for entities and at least two for relations (subject, object).

## Is the knowledge graph thread-safe?

Yes. `InMemoryGraph` uses `sync.RWMutex` to protect all read and write operations. Multiple goroutines can safely call `AddNode`, `AddEdge`, `GetNeighbors`, `ShortestPath`, and `Subgraph` concurrently.
