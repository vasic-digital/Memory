# Architecture -- Memory

## Purpose

Generic, reusable Go module for memory management with Mem0-style capabilities: scoped memory storage with search and pagination, automatic importance scoring with time-based decay, memory consolidation, pattern-based entity and relation extraction, an in-memory knowledge graph with BFS shortest path and subgraph extraction, and runtime memory leak detection with profiling utilities.

## Structure

```
pkg/
  store/    Core MemoryStore interface, Memory type, InMemoryStore (thread-safe, word-overlap search)
  mem0/     Mem0-style manager: importance scoring, exponential decay, consolidation of similar memories
  entity/   Pattern-based entity and relation extraction (email, URL, noun phrases, is_a/has/uses relations)
  graph/    In-memory directed weighted knowledge graph with BFS shortest path and bounded subgraph
  memory/   Memory leak detector, monitor with alert callbacks, heap/goroutine profiling utilities
```

## Key Components

- **`store.MemoryStore`** -- Interface: Add, Search, Get, Update, Delete, List with scope and pagination
- **`store.Memory`** -- Storage unit with ID, Content, Metadata, Scope (User/Session/Conversation/Global), Score, and Embedding
- **`mem0.Manager`** -- Wraps any MemoryStore with importance scoring (content length, metadata, embeddings, scope), exponential time decay, and Jaccard-similarity consolidation
- **`entity.PatternExtractor`** -- Regex-based entity and relation extraction with custom pattern support
- **`graph.InMemoryGraph`** -- Adjacency-list graph with AddNode, AddEdge, GetNeighbors, ShortestPath (BFS), Subgraph (bounded BFS)
- **`memory.LeakDetector`** -- Samples runtime.MemStats at intervals, tracks heap growth ratio and goroutine growth rate
- **`memory.MemoryMonitor`** -- Higher-level monitor with alert callbacks and report channel

## Data Flow

```
mem0.Manager.Add(memory) -> calculate importance -> assign scope -> store.Add()
mem0.Manager.Search(query, opts) -> store.Search() -> apply time decay to scores -> sort by decayed score
mem0.Manager.Consolidate(scope) -> pairwise Jaccard similarity -> merge above threshold

entity.Extract(text) -> entity patterns (email, URL, noun phrase) + relation patterns (is_a, has, uses)
    -> entities + relations -> graph.AddNode() + graph.AddEdge()

memory.LeakDetector.Start(ctx) -> periodic MemStats sampling -> heap growth ratio check
    -> PotentialLeak flag if growth > threshold or goroutine growth > 50%
```

## Dependencies

- `github.com/google/uuid` -- UUID generation for memory IDs
- `github.com/stretchr/testify` -- Test assertions

## Testing Strategy

Table-driven tests with `testify` and race detection. Tests cover memory CRUD, search scoring, scope filtering, importance calculation, decay application, consolidation merging, entity/relation extraction, graph traversal, shortest path, subgraph extraction, and leak detector sampling.
