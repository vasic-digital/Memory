# Lesson 3: Entity Extraction and Knowledge Graphs

## Objectives

- Extract entities and relations from text using `PatternExtractor`
- Build an `InMemoryGraph` from extracted data
- Query the graph with `ShortestPath` and `Subgraph`

## Concepts

### Entity Extraction

The `PatternExtractor` uses regex patterns to identify entities (emails, URLs, capitalized phrases) and relations (is_a, has, uses) in text. The `Extractor` interface returns entities and relations:

```go
type Extractor interface {
    Extract(text string) ([]Entity, []Relation, error)
}
```

### The Graph Interface

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

`InMemoryGraph` stores nodes in a map and edges in an adjacency list. It is thread-safe via `sync.RWMutex`.

## Code Walkthrough

### Custom entity patterns

```go
ext := entity.NewPatternExtractor().
    WithEntityPattern("hashtag", "tag", `#(\w+)`).
    WithRelationPattern("created_by", "created_by",
        `(\w+)\s+created by\s+(\w+)`)
```

### Building a graph from extracted data

```go
g := graph.NewInMemoryGraph()

entities, relations, _ := ext.Extract("Alice uses Go")

for _, e := range entities {
    g.AddNode(graph.Node{ID: e.Name, Type: e.Type})
}
for _, r := range relations {
    g.AddNode(graph.Node{ID: r.Subject, Type: "entity"})
    g.AddNode(graph.Node{ID: r.Object, Type: "entity"})
    g.AddEdge(graph.Edge{
        Source: r.Subject, Target: r.Object,
        Relation: r.Predicate, Weight: 1.0,
    })
}
```

### Querying the graph

BFS shortest path:

```go
path, err := g.ShortestPath("Alice", "Go")
// ["Alice", "Go"]
```

Depth-bounded subgraph:

```go
nodes, edges, err := g.Subgraph("Alice", 3)
// All nodes and edges reachable within 3 hops from Alice
```

## Summary

Combine `PatternExtractor` with `InMemoryGraph` to build knowledge graphs from unstructured text. Extend the default patterns with domain-specific regex for richer entity models.
