# Architecture

## Design Philosophy

The Memory module is designed around three principles:

1. **Interface-first**: Core operations are defined as interfaces (`MemoryStore`, `Extractor`, `Graph`), enabling pluggable implementations.
2. **Zero external runtime dependencies**: The module depends only on `github.com/google/uuid` at runtime. All implementations are in-memory and self-contained.
3. **Composition over inheritance**: The `mem0.Manager` wraps a `MemoryStore` backend rather than extending it, adding decay, importance, and consolidation as a layer.

## Package Structure

```
digital.vasic.memory/
    pkg/
        store/          Core interfaces and in-memory implementation
            store.go        MemoryStore interface, Memory struct, Scope, options
            inmemory.go     InMemoryStore (thread-safe, word-overlap search)
            store_test.go
        mem0/           Mem0-style manager layer
            mem0.go         Manager, Config, importance, decay, consolidation
            mem0_test.go
        entity/         Entity and relation extraction
            entity.go       Extractor interface, PatternExtractor, default patterns
            entity_test.go
        graph/          Knowledge graph
            graph.go        Graph interface, InMemoryGraph (adjacency list, BFS)
            graph_test.go
```

## Design Patterns

### Strategy Pattern -- MemoryStore

The `MemoryStore` interface defines the contract for memory storage:

```go
type MemoryStore interface {
    Add(ctx context.Context, memory *Memory) error
    Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error)
    Get(ctx context.Context, id string) (*Memory, error)
    Update(ctx context.Context, memory *Memory) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, scope Scope, opts *ListOptions) ([]*Memory, error)
}
```

`InMemoryStore` is the built-in strategy. Alternative strategies (PostgreSQL, Redis, vector databases) can be implemented without modifying existing code. The `mem0.Manager` accepts any `MemoryStore` as its backend, making the storage mechanism interchangeable.

### Decorator Pattern -- Manager

`mem0.Manager` wraps a `MemoryStore` backend and adds cross-cutting concerns:

- **Importance scoring** on `Add` and `Update`
- **Exponential decay** on `Search` and `List`
- **Consolidation** via the additional `Consolidate` method

The Manager itself implements `MemoryStore`, so it can be used anywhere a `MemoryStore` is expected. This enables layered decoration:

```
Application --> Manager (decay, importance, consolidation) --> InMemoryStore (storage)
```

### Repository Pattern -- InMemoryStore and InMemoryGraph

Both `InMemoryStore` and `InMemoryGraph` follow the Repository pattern:

- Encapsulate data access behind a clean interface
- Return copies of stored objects to prevent external mutation
- Manage their own concurrency via `sync.RWMutex`
- Provide CRUD operations plus query methods (Search, ShortestPath, Subgraph)

### Builder Pattern -- PatternExtractor

`PatternExtractor` supports fluent configuration:

```go
pe := entity.NewPatternExtractor().
    WithEntityPattern("version", "version", `v(\d+\.\d+\.\d+)`).
    WithRelationPattern("depends", "depends_on", `(\w+)\s+depends\s+on\s+(\w+)`)
```

Each `With*` method returns the same `*PatternExtractor`, enabling method chaining. The constructor provides sensible defaults (email, URL, noun_phrase patterns; is_a, has, uses relations).

### Factory Pattern -- Constructors

Each implementation provides a factory function:

| Factory | Returns |
|---------|---------|
| `store.NewInMemoryStore()` | `*InMemoryStore` |
| `mem0.NewManager(backend, config)` | `*Manager` |
| `mem0.DefaultConfig()` | `*Config` |
| `entity.NewPatternExtractor()` | `*PatternExtractor` |
| `graph.NewInMemoryGraph()` | `*InMemoryGraph` |

Factory functions return concrete types (not interfaces) so that implementation-specific methods remain accessible. Interface compliance is verified at compile time via tests.

## Key Design Decisions

### 1. Context on All Store Operations

Every `MemoryStore` and `Graph` write/read method accepts `context.Context`. This enables:
- Cancellation propagation for long-running searches
- Deadline enforcement for database-backed implementations
- Tracing and observability via context values

The `Extractor` interface does not take a context because regex extraction is CPU-bound and non-blocking.

### 2. Copy Semantics

`InMemoryStore.Add` stores a copy of the input memory. `Get` returns a copy of the stored memory. This prevents callers from accidentally mutating stored state through retained pointers. The same pattern applies to `InMemoryGraph`.

### 3. Word-Overlap Search Scoring

The in-memory search uses Jaccard-style word overlap (fraction of query words present in content). This is a deliberate simplification:
- No external dependencies required
- Predictable scoring behavior for testing
- Production systems should implement `MemoryStore` with vector similarity (using the `Embedding` field)

### 4. Exponential Decay Model

Memory decay uses `score * exp(-rate * hours)`:
- Applied at read time, not write time, preserving original scores
- Rate = 0 disables decay entirely
- Future creation times are clamped to 0 hours (no negative decay)
- Configurable per-manager, not per-memory

### 5. Consolidation as Explicit Operation

Consolidation is triggered explicitly via `Manager.Consolidate()` rather than running automatically in the background. This gives the caller full control over when consolidation occurs and avoids hidden goroutines. A cooldown interval prevents accidental rapid re-runs.

### 6. Directed Graph with BFS

The knowledge graph uses directed, weighted edges with an adjacency list representation. BFS is used for shortest path (by hop count, not edge weight) because:
- Hop count is the most common traversal metric for knowledge graphs
- BFS guarantees optimal hop-count paths
- Weighted shortest path (Dijkstra) can be added as a separate method if needed

### 7. Independent entity and graph Packages

`entity` and `graph` have zero dependencies on each other or on `store`/`mem0`. This allows:
- Independent versioning and testing
- Use in projects that need only extraction or only graph capabilities
- Clean dependency trees with no circular imports

## Concurrency Model

### Locking Strategy

Both `InMemoryStore` and `InMemoryGraph` use `sync.RWMutex`:

| Operation | Lock Type |
|-----------|-----------|
| Add, Update, Delete | `Lock` (exclusive) |
| Get, Search, List, Nodes, Edges, GetNode, GetNeighbors, ShortestPath, Subgraph | `RLock` (shared) |

`mem0.Manager` adds its own `sync.RWMutex` for:
- `Add`: exclusive lock (scope/ID assignment)
- `Consolidate`: exclusive lock (read-modify-delete cycle)
- `Search`, `List`, `Get`: no additional lock (delegated to backend)

### Thread Safety Guarantees

1. Multiple goroutines can safely call any method on the same instance
2. Write operations are serialized; read operations are concurrent
3. No goroutine leaks: no background goroutines are created
4. All returned data is safe to mutate without affecting stored state

## Extension Points

### Custom MemoryStore Backend

Implement the `MemoryStore` interface for production storage:

```go
type PostgresStore struct { db *pgx.Pool }

func (s *PostgresStore) Add(ctx context.Context, memory *store.Memory) error { ... }
func (s *PostgresStore) Search(ctx context.Context, query string, opts *store.SearchOptions) ([]*store.Memory, error) { ... }
// ... remaining methods
```

Wrap with `mem0.Manager` to add decay and consolidation.

### Custom Entity Patterns

Use `WithEntityPattern` and `WithRelationPattern` to add domain-specific extraction rules without modifying the default patterns.

### Custom Graph Implementation

Implement the `Graph` interface for persistent graph storage (Neo4j, Dgraph, etc.) while maintaining the same traversal API.

## Data Flow

### Memory Store and Retrieve

```
Add:     Application -> Manager.Add -> CalculateImportance -> backend.Add -> InMemoryStore (mutex, copy, store)
Search:  Application -> Manager.Search -> backend.Search -> InMemoryStore (mutex, score, sort, TopK) -> ApplyDecay -> results
```

### Entity Extraction to Knowledge Graph

```
Text -> PatternExtractor.Extract -> ([]Entity, []Relation)
    Entity -> graph.AddNode(Node{ID: entity.Name, Type: entity.Type})
    Relation -> graph.AddEdge(Edge{Source: r.Subject, Target: r.Object, Relation: r.Predicate})
```

### Consolidation

```
Manager.Consolidate(scope)
    -> backend.List(scope)
    -> pairwise wordOverlapSimilarity
    -> if sim >= threshold: mergeMemories(target, source)
    -> backend.Update(target)
    -> backend.Delete(source)
```
