# API Reference

Complete reference for all exported types, functions, and methods in the `digital.vasic.memory` module.

---

## Package `store`

**Import**: `digital.vasic.memory/pkg/store`

Core memory store interfaces, types, and in-memory implementation.

### Types

#### `Scope`

```go
type Scope string
```

Defines the visibility boundary of a memory.

**Constants**:

| Constant | Value | Description |
|----------|-------|-------------|
| `ScopeUser` | `"user"` | Memory visible to a specific user |
| `ScopeSession` | `"session"` | Memory visible within a session |
| `ScopeConversation` | `"conversation"` | Memory visible within a conversation |
| `ScopeGlobal` | `"global"` | Memory visible to all users |

#### `Memory`

```go
type Memory struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    Metadata  map[string]any         `json:"metadata,omitempty"`
    Scope     Scope                  `json:"scope"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
    Score     float64                `json:"score"`
    Embedding []float32              `json:"embedding,omitempty"`
}
```

Represents a stored memory unit with content, metadata, embedding vector, and scoping information.

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier. Auto-generated (UUID) if empty on Add. |
| `Content` | `string` | The text content of the memory. |
| `Metadata` | `map[string]any` | Arbitrary key-value metadata. |
| `Scope` | `Scope` | Visibility boundary. |
| `CreatedAt` | `time.Time` | Creation timestamp. Auto-set if zero on Add. |
| `UpdatedAt` | `time.Time` | Last update timestamp. Auto-set if zero on Add, updated on Update. |
| `Score` | `float64` | Relevance or importance score. |
| `Embedding` | `[]float32` | Optional vector embedding for semantic search. |

#### `SearchOptions`

```go
type SearchOptions struct {
    TopK      int                `json:"top_k"`
    MinScore  float64            `json:"min_score"`
    Scope     Scope              `json:"scope,omitempty"`
    TimeRange *TimeRange         `json:"time_range,omitempty"`
    Filter    map[string]any     `json:"filter,omitempty"`
}
```

Configures memory search behaviour.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `TopK` | `int` | 10 | Maximum number of results to return. 0 means no limit. |
| `MinScore` | `float64` | 0.0 | Minimum match score threshold. |
| `Scope` | `Scope` | `""` | Filter by scope. Empty means all scopes. |
| `TimeRange` | `*TimeRange` | `nil` | Filter by creation time window. |
| `Filter` | `map[string]any` | `nil` | Metadata filters (reserved for custom implementations). |

#### `TimeRange`

```go
type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}
```

Restricts search results to a time window based on `CreatedAt`.

#### `ListOptions`

```go
type ListOptions struct {
    Offset  int    `json:"offset"`
    Limit   int    `json:"limit"`
    OrderBy string `json:"order_by"`
    Scope   Scope  `json:"scope,omitempty"`
}
```

Configures memory listing with pagination and ordering.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Offset` | `int` | 0 | Number of results to skip. |
| `Limit` | `int` | 100 | Maximum number of results. 0 means no limit. |
| `OrderBy` | `string` | `"created_at"` | Sort field: `"created_at"`, `"updated_at"`, or `"score"`. |
| `Scope` | `Scope` | `""` | Additional scope filter. |

### Functions

#### `DefaultSearchOptions`

```go
func DefaultSearchOptions() *SearchOptions
```

Returns sensible default search options: TopK=10, MinScore=0.0.

#### `DefaultListOptions`

```go
func DefaultListOptions() *ListOptions
```

Returns sensible default list options: Offset=0, Limit=100, OrderBy="created_at".

### Interfaces

#### `MemoryStore`

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

Defines the interface for memory storage operations.

| Method | Description |
|--------|-------------|
| `Add` | Stores a new memory. |
| `Search` | Returns memories matching the query and options. |
| `Get` | Retrieves a memory by ID. Returns error if not found. |
| `Update` | Modifies an existing memory. Returns error if not found. |
| `Delete` | Removes a memory by ID. Returns error if not found. |
| `List` | Returns memories matching the scope and options. |

### `InMemoryStore`

```go
type InMemoryStore struct { /* unexported fields */ }
```

Thread-safe in-memory implementation of `MemoryStore`. Uses `sync.RWMutex` for concurrency safety. Returns copies of stored memories to prevent external mutation.

#### `NewInMemoryStore`

```go
func NewInMemoryStore() *InMemoryStore
```

Creates a new empty in-memory store.

#### Methods

```go
func (s *InMemoryStore) Add(ctx context.Context, memory *Memory) error
```

Stores a new memory. Generates a UUID if `memory.ID` is empty. Sets `CreatedAt` and `UpdatedAt` to current time if zero. Stores a copy of the input.

```go
func (s *InMemoryStore) Get(ctx context.Context, id string) (*Memory, error)
```

Retrieves a memory by ID. Returns a copy. Error if not found.

```go
func (s *InMemoryStore) Update(ctx context.Context, memory *Memory) error
```

Modifies an existing memory. Sets `UpdatedAt` to current time. Error if not found.

```go
func (s *InMemoryStore) Delete(ctx context.Context, id string) error
```

Removes a memory by ID. Error if not found.

```go
func (s *InMemoryStore) Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error)
```

Returns memories matching the query. Uses word-overlap scoring: the fraction of query words (case-insensitive) found in the memory content. Results sorted by descending score, limited by `TopK`. Applies scope and time range filters.

```go
func (s *InMemoryStore) List(ctx context.Context, scope Scope, opts *ListOptions) ([]*Memory, error)
```

Returns memories matching the scope. Supports pagination (`Offset`, `Limit`) and ordering (`OrderBy`). Pass empty scope to list all memories.

---

## Package `mem0`

**Import**: `digital.vasic.memory/pkg/mem0`

Mem0-style memory management with consolidation, decay, and importance scoring.

### Types

#### `Config`

```go
type Config struct {
    DefaultScope          store.Scope
    MaxMemories           int
    ConsolidationInterval time.Duration
    DecayRate             float64
    SimilarityThreshold   float64
}
```

Configures the Mem0-style memory manager.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DefaultScope` | `store.Scope` | `ScopeUser` | Scope assigned when memory has no explicit scope. |
| `MaxMemories` | `int` | 10000 | Max memories per scope for consolidation listing. 0 = unlimited. |
| `ConsolidationInterval` | `time.Duration` | 5 min | Minimum time between consolidation runs. |
| `DecayRate` | `float64` | 0.01 | Exponential decay rate. 0 disables decay. |
| `SimilarityThreshold` | `float64` | 0.7 | Minimum Jaccard similarity for merging (0.0--1.0). |

#### `Manager`

```go
type Manager struct { /* unexported fields */ }
```

Implements `store.MemoryStore` with Mem0-style memory operations. Wraps a backend `MemoryStore` adding decay, importance scoring, and consolidation.

### Functions

#### `DefaultConfig`

```go
func DefaultConfig() *Config
```

Returns a sensible default configuration.

#### `NewManager`

```go
func NewManager(backend store.MemoryStore, config *Config) *Manager
```

Creates a new Mem0-style memory manager wrapping a backend store. If `config` is nil, `DefaultConfig()` is used.

#### `CalculateImportance`

```go
func CalculateImportance(memory *store.Memory) float64
```

Computes an importance score (0.0--1.0) for a memory based on:
- Base: 0.5
- Content > 100 chars: +0.1
- Content > 500 chars: +0.1
- Has metadata: +0.1
- Has embedding: +0.1
- Global scope: +0.1
- Capped at 1.0, rounded to 2 decimal places.

#### `ApplyDecay`

```go
func ApplyDecay(score float64, createdAt time.Time, now time.Time, rate float64) float64
```

Reduces a score based on time elapsed using exponential decay: `score * exp(-rate * hours)`. Returns the original score if `rate <= 0`. Clamps negative hours to 0.

### Manager Methods

```go
func (m *Manager) Add(ctx context.Context, memory *store.Memory) error
```

Stores a new memory with automatic scope assignment (if empty), ID generation (if empty), and importance scoring (if score is zero). Delegates storage to the backend.

```go
func (m *Manager) Search(ctx context.Context, query string, opts *store.SearchOptions) ([]*store.Memory, error)
```

Searches the backend and applies exponential decay to result scores (if `DecayRate > 0`).

```go
func (m *Manager) Get(ctx context.Context, id string) (*store.Memory, error)
```

Retrieves a memory by ID from the backend. No decay applied.

```go
func (m *Manager) Update(ctx context.Context, memory *store.Memory) error
```

Updates a memory, recalculates importance, and sets `UpdatedAt`.

```go
func (m *Manager) Delete(ctx context.Context, id string) error
```

Deletes a memory from the backend.

```go
func (m *Manager) List(ctx context.Context, scope store.Scope, opts *store.ListOptions) ([]*store.Memory, error)
```

Lists memories from the backend and applies decay to scores (if `DecayRate > 0`).

```go
func (m *Manager) Consolidate(ctx context.Context, scope store.Scope) (int, error)
```

Merges similar memories within the given scope. Returns the number of memories consolidated. Respects `ConsolidationInterval` cooldown. Uses Jaccard word-overlap similarity. Merging keeps longer content, merges metadata (existing keys preserved), keeps higher score, uses earlier creation time.

---

## Package `entity`

**Import**: `digital.vasic.memory/pkg/entity`

Pattern-based entity and relation extraction from text.

### Types

#### `Entity`

```go
type Entity struct {
    Name       string         `json:"name"`
    Type       string         `json:"type"`
    Attributes map[string]any `json:"attributes,omitempty"`
}
```

Represents an extracted named entity.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | The entity's name/value as extracted from text. |
| `Type` | `string` | Category: `"email"`, `"url"`, `"noun_phrase"`, or custom. |
| `Attributes` | `map[string]any` | Optional attributes (not populated by default patterns). |

#### `Relation`

```go
type Relation struct {
    Subject   string `json:"subject"`
    Predicate string `json:"predicate"`
    Object    string `json:"object"`
}
```

Represents a directed relationship between two entities (subject-predicate-object triple).

| Field | Type | Description |
|-------|------|-------------|
| `Subject` | `string` | The source entity name. |
| `Predicate` | `string` | The relationship type: `"is_a"`, `"has"`, `"uses"`, or custom. |
| `Object` | `string` | The target entity name. |

#### `Pattern`

```go
type Pattern struct {
    Name       string
    Type       string
    Expression *regexp.Regexp
}
```

Defines a named regex pattern for entity extraction. The regex must contain one capture group for the entity name.

### Interfaces

#### `Extractor`

```go
type Extractor interface {
    Extract(text string) ([]Entity, []Relation, error)
}
```

Defines the interface for entity and relation extraction from text.

### `PatternExtractor`

```go
type PatternExtractor struct { /* unexported fields */ }
```

Extracts entities and relations using regex patterns. Implements `Extractor`.

#### `NewPatternExtractor`

```go
func NewPatternExtractor() *PatternExtractor
```

Creates a new `PatternExtractor` with default patterns:
- **Entity patterns**: email, url, capitalized_phrase (noun_phrase)
- **Relation patterns**: is_a, has, uses

#### Methods

```go
func (pe *PatternExtractor) WithEntityPattern(name, entityType, pattern string) *PatternExtractor
```

Adds a custom entity extraction pattern. Returns the same `*PatternExtractor` for chaining. The regex `pattern` must contain one capture group. Panics if the pattern is invalid (uses `regexp.MustCompile`).

```go
func (pe *PatternExtractor) WithRelationPattern(name, predicate, pattern string) *PatternExtractor
```

Adds a custom relation extraction pattern. Returns the same `*PatternExtractor` for chaining. The regex `pattern` must contain two capture groups (subject, object). Panics if the pattern is invalid.

```go
func (pe *PatternExtractor) Extract(text string) ([]Entity, []Relation, error)
```

Extracts entities and relations from the given text. Entities are deduplicated by name. Returns `(entities, relations, nil)` -- the error is always nil for the pattern-based implementation.

---

## Package `graph`

**Import**: `digital.vasic.memory/pkg/graph`

In-memory knowledge graph with directed weighted edges, BFS shortest path, and subgraph extraction.

### Types

#### `Node`

```go
type Node struct {
    ID         string         `json:"id"`
    Type       string         `json:"type"`
    Properties map[string]any `json:"properties,omitempty"`
}
```

Represents a vertex in the knowledge graph.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique node identifier. Must be non-empty. |
| `Type` | `string` | Node category (e.g., `"person"`, `"language"`, `"concept"`). |
| `Properties` | `map[string]any` | Arbitrary key-value properties. |

#### `Edge`

```go
type Edge struct {
    Source   string  `json:"source"`
    Target   string  `json:"target"`
    Relation string  `json:"relation"`
    Weight   float64 `json:"weight"`
}
```

Represents a directed, weighted connection between two nodes.

| Field | Type | Description |
|-------|------|-------------|
| `Source` | `string` | ID of the source node. |
| `Target` | `string` | ID of the target node. |
| `Relation` | `string` | Relationship type label. |
| `Weight` | `float64` | Edge weight (not used in BFS shortest path). |

### Interfaces

#### `Graph`

```go
type Graph interface {
    AddNode(node Node) error
    AddEdge(edge Edge) error
    GetNode(id string) (Node, error)
    GetNeighbors(id string) ([]Node, error)
    ShortestPath(from, to string) ([]string, error)
    Subgraph(startID string, maxDepth int) ([]Node, []Edge, error)
    Nodes() []Node
    Edges() []Edge
}
```

Defines the interface for knowledge graph operations.

| Method | Description |
|--------|-------------|
| `AddNode` | Adds a node. Error if ID is empty or node already exists. |
| `AddEdge` | Adds a directed edge. Error if source or target node not found. |
| `GetNode` | Returns a node by ID. Error if not found. |
| `GetNeighbors` | Returns all nodes directly connected from the given node (deduplicated). Error if node not found. |
| `ShortestPath` | Finds shortest path by hop count (BFS). Returns node IDs from start to end. Error if no path or node not found. |
| `Subgraph` | Returns all nodes and edges reachable within `maxDepth` hops from start. Error if start node not found. |
| `Nodes` | Returns all nodes in the graph. |
| `Edges` | Returns all edges in the graph. |

### `InMemoryGraph`

```go
type InMemoryGraph struct { /* unexported fields */ }
```

Thread-safe in-memory implementation of `Graph` using adjacency lists. Uses `sync.RWMutex` for concurrency safety.

#### `NewInMemoryGraph`

```go
func NewInMemoryGraph() *InMemoryGraph
```

Creates a new empty in-memory graph.

#### Methods

```go
func (g *InMemoryGraph) AddNode(node Node) error
```

Adds a node. Returns error if ID is empty (`"node ID cannot be empty"`) or already exists (`"node already exists: <id>"`).

```go
func (g *InMemoryGraph) AddEdge(edge Edge) error
```

Adds a directed edge. Both source and target nodes must exist. Returns error if source (`"source node not found: <id>"`) or target (`"target node not found: <id>"`) is missing.

```go
func (g *InMemoryGraph) GetNode(id string) (Node, error)
```

Returns a node by ID. Returns error (`"node not found: <id>"`) if not found.

```go
func (g *InMemoryGraph) GetNeighbors(id string) ([]Node, error)
```

Returns all nodes directly reachable via outgoing edges from the given node. Deduplicated: if multiple edges point to the same target, the target appears once. Returns error if the node is not found.

```go
func (g *InMemoryGraph) ShortestPath(from, to string) ([]string, error)
```

Finds the shortest path by hop count using BFS. Returns a slice of node IDs from `from` to `to`, inclusive. If `from == to`, returns `[]string{from}`. Returns error if either node is not found or if no path exists (`"no path from <from> to <to>"`).

```go
func (g *InMemoryGraph) Subgraph(startID string, maxDepth int) ([]Node, []Edge, error)
```

Returns all nodes and edges reachable within `maxDepth` hops from `startID` using BFS. At depth 0, returns only the start node with no edges. Returns error if the start node is not found.

```go
func (g *InMemoryGraph) Nodes() []Node
```

Returns all nodes in the graph as a new slice.

```go
func (g *InMemoryGraph) Edges() []Edge
```

Returns all edges in the graph as a new slice (copy of internal slice).
