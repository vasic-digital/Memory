# Memory

Generic, reusable Go module for memory management with Mem0-style capabilities, entity extraction, knowledge graph construction, and memory leak detection. Provides scoped memory storage with search and pagination, automatic importance scoring with time-based decay, memory consolidation, pattern-based entity and relation extraction, an in-memory knowledge graph with BFS shortest path and subgraph extraction, and runtime memory leak monitoring.

**Module**: `digital.vasic.memory` (Go 1.24+)

## Architecture

The module follows a layered architecture. The store package defines the core memory interface and provides an in-memory implementation. The mem0 package wraps any MemoryStore backend with Mem0-style intelligence: automatic importance scoring, exponential time decay, and memory consolidation. The entity package extracts structured entities and relations from unstructured text. The graph package builds and queries knowledge graphs. The memory package provides runtime memory leak detection.

```
pkg/
  store/     Core MemoryStore interface, Memory type, InMemoryStore
  mem0/      Mem0-style manager: importance, decay, consolidation
  entity/    Pattern-based entity and relation extraction
  graph/     In-memory knowledge graph with BFS and subgraph
  memory/    Memory leak detector, monitor, and profiling utilities
```

## Package Reference

### pkg/store -- Core Memory Store

Defines the fundamental memory types, scoping model, and storage interface.

**Types:**
- `Scope` -- Visibility boundaries: ScopeUser, ScopeSession, ScopeConversation, ScopeGlobal.
- `Memory` -- Storage unit with ID, Content, Metadata (map), Scope, CreatedAt, UpdatedAt, Score, and Embedding vector.
- `SearchOptions` -- TopK (10), MinScore (0.0), Scope filter, TimeRange, and metadata Filter.
- `TimeRange` -- Start and End time for temporal filtering.
- `ListOptions` -- Offset (0), Limit (100), OrderBy ("created_at"), and Scope filter.
- `MemoryStore` -- Interface:
  - `Add(ctx, memory) error` -- Store a new memory.
  - `Search(ctx, query, opts) ([]*Memory, error)` -- Search by text query.
  - `Get(ctx, id) (*Memory, error)` -- Retrieve by ID.
  - `Update(ctx, memory) error` -- Modify an existing memory.
  - `Delete(ctx, id) error` -- Remove by ID.
  - `List(ctx, scope, opts) ([]*Memory, error)` -- List with pagination and ordering.

**InMemoryStore:**
- `NewInMemoryStore() *InMemoryStore` -- Creates a thread-safe in-memory store.
- Auto-generates UUIDs for memories without IDs.
- Search uses word-overlap scoring (Jaccard-like on query terms vs content).
- List supports ordering by created_at, updated_at, or score with offset/limit pagination.
- All operations are thread-safe via `sync.RWMutex`.

### pkg/mem0 -- Mem0-Style Memory Manager

Wraps any MemoryStore backend with intelligent memory operations inspired by Mem0.

**Types:**
- `Config`:
  - DefaultScope (ScopeUser) -- Scope for memories without explicit scope.
  - MaxMemories (10000) -- Maximum memories per scope.
  - ConsolidationInterval (5m) -- Minimum time between consolidation runs.
  - DecayRate (0.01) -- Exponential decay rate per hour. Zero disables decay.
  - SimilarityThreshold (0.7) -- Jaccard similarity for consolidation merging.
- `Manager` -- Implements `store.MemoryStore` with enhanced operations.

**Key Functions:**
- `NewManager(backend store.MemoryStore, config *Config) *Manager` -- Creates a manager wrapping any backend.
- `Manager.Add(ctx, memory) error` -- Stores with automatic scope assignment and importance scoring.
- `Manager.Search(ctx, query, opts) ([]*Memory, error)` -- Searches with automatic decay applied to scores.
- `Manager.Update(ctx, memory) error` -- Updates and recalculates importance.
- `Manager.Consolidate(ctx, scope) (int, error)` -- Merges similar memories within a scope. Returns count of consolidated memories. Respects the cooldown interval.
- `CalculateImportance(memory) float64` -- Computes importance (0.0-1.0) based on:
  - Base score: 0.5
  - Content length > 100 chars: +0.1
  - Content length > 500 chars: +0.1
  - Has metadata: +0.1
  - Has embeddings: +0.1
  - Global scope: +0.1
- `ApplyDecay(score, createdAt, now, rate) float64` -- Exponential decay: `score * exp(-rate * hours)`.

**Consolidation Algorithm:** Iterates all memories in a scope, computes pairwise Jaccard word-overlap similarity, merges memories above the threshold by keeping the longer content, combining metadata, retaining the higher score, and using the earlier creation time.

### pkg/entity -- Entity and Relation Extraction

Extracts structured entities and relations from unstructured text using regex patterns.

**Types:**
- `Entity` -- Name, Type, and Attributes map.
- `Relation` -- Subject, Predicate, Object triple.
- `Extractor` -- Interface with `Extract(text) ([]Entity, []Relation, error)`.
- `Pattern` -- Named regex pattern with entity type.
- `PatternExtractor` -- Configurable extractor with entity and relation patterns.

**Default Entity Patterns:**
- Email addresses (type: "email")
- URLs (type: "url")
- Capitalized phrases (type: "noun_phrase") -- e.g., "Machine Learning"

**Default Relation Patterns:**
- `is_a` -- "X is a Y" / "X is an Y"
- `has` -- "X has Y"
- `uses` -- "X uses Y"

**Key Functions:**
- `NewPatternExtractor() *PatternExtractor` -- Creates an extractor with default patterns.
- `PatternExtractor.WithEntityPattern(name, entityType, pattern) *PatternExtractor` -- Adds a custom entity pattern. Returns the extractor for chaining.
- `PatternExtractor.WithRelationPattern(name, predicate, pattern) *PatternExtractor` -- Adds a custom relation pattern.
- `PatternExtractor.Extract(text) ([]Entity, []Relation, error)` -- Extracts entities and relations.

### pkg/graph -- Knowledge Graph

In-memory directed weighted knowledge graph with BFS shortest path and bounded subgraph extraction.

**Types:**
- `Node` -- ID, Type, and Properties map.
- `Edge` -- Source, Target, Relation, and Weight.
- `Graph` -- Interface:
  - `AddNode(node) error`
  - `AddEdge(edge) error`
  - `GetNode(id) (Node, error)`
  - `GetNeighbors(id) ([]Node, error)` -- Outgoing neighbors.
  - `ShortestPath(from, to) ([]string, error)` -- BFS shortest path by hop count.
  - `Subgraph(startID, maxDepth) ([]Node, []Edge, error)` -- Bounded BFS subgraph.
  - `Nodes() []Node` / `Edges() []Edge`
- `InMemoryGraph` -- Thread-safe implementation using adjacency lists.

**Key Functions:**
- `NewInMemoryGraph() *InMemoryGraph` -- Creates an empty graph.
- `InMemoryGraph.AddNode(node) error` -- Rejects empty IDs and duplicates.
- `InMemoryGraph.AddEdge(edge) error` -- Requires both source and target nodes to exist.
- `InMemoryGraph.ShortestPath(from, to) ([]string, error)` -- Returns node IDs from start to end. Uses BFS for shortest hop count. Returns error if no path exists.
- `InMemoryGraph.Subgraph(startID, maxDepth) ([]Node, []Edge, error)` -- Returns all nodes and edges reachable within maxDepth hops via BFS.

### pkg/memory -- Leak Detection and Monitoring

Runtime memory leak detection with configurable thresholds, periodic sampling, goroutine monitoring, and profiling utilities.

**Types:**
- `LeakDetector` -- Samples `runtime.MemStats` at configurable intervals. Tracks heap growth ratio and goroutine growth rate.
- `LeakReport` -- HeapAlloc, HeapSys, HeapInUse, HeapObjects, StackInUse, GoroutineCount, GCCount, HeapGrowthRatio, PotentialLeak flag, GoroutineGrowthRate.
- `MemoryMonitor` -- Higher-level monitor with alert callbacks and a report channel.

**Key Functions:**
- `NewLeakDetector(interval, thresholdRatio) *LeakDetector` -- Creates a detector.
- `LeakDetector.Start(ctx) error` -- Begins background monitoring.
- `LeakDetector.Stop()` -- Stops monitoring.
- `LeakDetector.GetReport() LeakReport` -- Snapshot report. `PotentialLeak` is true when heap growth exceeds threshold or goroutine growth exceeds 50%.
- `LeakDetector.GetSamples() []runtime.MemStats` -- Historical samples (up to 100).
- `NewMemoryMonitor(interval, thresholdRatio) *MemoryMonitor` -- Creates a monitor.
- `MemoryMonitor.SetAlertCallback(cb func(LeakReport))` -- Registers leak alert handler.
- `MemoryMonitor.Start(ctx) error` / `MemoryMonitor.Stop()`
- `MemoryMonitor.Reports() <-chan LeakReport` -- Channel of periodic reports.
- `WriteHeapProfile(filename) error` -- Writes pprof heap profile.
- `WriteGoroutineProfile(filename) error` -- Writes goroutine profile.
- `ForceGC()` -- Forces two GC cycles.
- `GetCurrentMemoryUsage() string` -- Human-readable memory stats.

## Usage Examples

### Basic Memory Store

```go
backend := store.NewInMemoryStore()

backend.Add(ctx, &store.Memory{
    Content: "User prefers concise responses",
    Scope:   store.ScopeUser,
})

results, _ := backend.Search(ctx, "preferences", &store.SearchOptions{
    TopK:     5,
    MinScore: 0.3,
    Scope:    store.ScopeUser,
})
```

### Mem0-Style Memory Manager

```go
backend := store.NewInMemoryStore()
manager := mem0.NewManager(backend, &mem0.Config{
    DefaultScope:        store.ScopeUser,
    DecayRate:           0.01,
    SimilarityThreshold: 0.7,
})

manager.Add(ctx, &store.Memory{
    Content: "User prefers dark mode and concise answers",
})

// Search with automatic time-based decay
results, _ := manager.Search(ctx, "user preferences", nil)

// Consolidate similar memories
merged, _ := manager.Consolidate(ctx, store.ScopeUser)
fmt.Printf("Consolidated %d memories\n", merged)
```

### Entity Extraction

```go
extractor := entity.NewPatternExtractor().
    WithEntityPattern("version", "version", `v(\d+\.\d+\.\d+)`).
    WithRelationPattern("depends_on", "depends_on",
        `(?i)(\w+)\s+depends on\s+(\w+)`)

entities, relations, _ := extractor.Extract(
    "Go uses concurrency. contact@example.com is the admin.",
)
```

### Knowledge Graph

```go
g := graph.NewInMemoryGraph()
g.AddNode(graph.Node{ID: "go", Type: "language"})
g.AddNode(graph.Node{ID: "concurrency", Type: "feature"})
g.AddNode(graph.Node{ID: "goroutine", Type: "concept"})

g.AddEdge(graph.Edge{Source: "go", Target: "concurrency", Relation: "has"})
g.AddEdge(graph.Edge{Source: "concurrency", Target: "goroutine", Relation: "uses"})

path, _ := g.ShortestPath("go", "goroutine")
// ["go", "concurrency", "goroutine"]

nodes, edges, _ := g.Subgraph("go", 2)
```

### Memory Leak Detection

```go
monitor := memory.NewMemoryMonitor(5*time.Second, 2.0)
monitor.SetAlertCallback(func(report memory.LeakReport) {
    log.Printf("Potential leak: heap growth %.2fx", report.HeapGrowthRatio)
})
monitor.Start(ctx)
defer monitor.Stop()

// Periodic reports
go func() {
    for report := range monitor.Reports() {
        if report.PotentialLeak {
            memory.WriteHeapProfile("/tmp/heap.prof")
        }
    }
}()
```

## Configuration

All packages use Config structs with `DefaultConfig()` constructors. Key defaults:
- Store: TopK=10, Limit=100, OrderBy="created_at"
- Mem0: MaxMemories=10000, DecayRate=0.01, SimilarityThreshold=0.7, ConsolidationInterval=5m
- LeakDetector: configurable interval and threshold ratio

## Testing

```bash
go test ./... -count=1 -race    # All tests with race detection
go test ./... -short             # Unit tests only
go test -bench=. ./...           # Benchmarks
```

## Integration with HelixAgent

The Memory module is the primary memory system for HelixAgent:
- Mem0 Manager provides the default memory backend for debate sessions and user preferences
- Entity extraction feeds the knowledge graph during debate orchestration
- Knowledge graph enables semantic relationship traversal for context enrichment
- Memory scoping separates user, session, conversation, and global contexts
- Consolidation automatically merges redundant memories to prevent bloat
- Time-based decay ensures recent memories are prioritized over stale ones
- Leak detection monitors HelixAgent's runtime memory health

The internal adapter at `internal/adapters/memory/` bridges these generic types to HelixAgent-specific interfaces. HelixMemory (the higher-level cognitive memory engine) orchestrates this module alongside Cognee, Letta, and Graphiti.

## License

Proprietary.
