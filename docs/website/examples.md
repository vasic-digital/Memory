# Examples

## 1. Scoped Memory with Pagination

Store memories in different scopes and list them with pagination.

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.memory/pkg/store"
)

func main() {
    s := store.NewInMemoryStore()
    ctx := context.Background()

    // Add user-scoped memories
    for i := 0; i < 20; i++ {
        s.Add(ctx, &store.Memory{
            Content: fmt.Sprintf("User preference #%d", i),
            Scope:   store.ScopeUser,
        })
    }

    // Add global memories
    s.Add(ctx, &store.Memory{
        Content: "System-wide setting: timezone=UTC",
        Scope:   store.ScopeGlobal,
    })

    // Page through user memories
    page1, _ := s.List(ctx, store.ScopeUser, &store.ListOptions{
        Offset:  0,
        Limit:   10,
        OrderBy: "created_at",
    })
    fmt.Printf("Page 1: %d memories\n", len(page1))

    page2, _ := s.List(ctx, store.ScopeUser, &store.ListOptions{
        Offset:  10,
        Limit:   10,
        OrderBy: "created_at",
    })
    fmt.Printf("Page 2: %d memories\n", len(page2))
}
```

## 2. Entity Extraction to Knowledge Graph

Extract entities and relations from text, then build a graph.

```go
package main

import (
    "fmt"

    "digital.vasic.memory/pkg/entity"
    "digital.vasic.memory/pkg/graph"
)

func main() {
    ext := entity.NewPatternExtractor()

    texts := []string{
        "Alice uses PostgreSQL for data storage",
        "Bob uses Redis for caching",
        "Alice is a backend developer",
    }

    g := graph.NewInMemoryGraph()

    for _, text := range texts {
        entities, relations, _ := ext.Extract(text)

        for _, e := range entities {
            _ = g.AddNode(graph.Node{
                ID:   e.Name,
                Type: e.Type,
            })
        }
        for _, r := range relations {
            _ = g.AddNode(graph.Node{ID: r.Subject, Type: "entity"})
            _ = g.AddNode(graph.Node{ID: r.Object, Type: "entity"})
            _ = g.AddEdge(graph.Edge{
                Source:   r.Subject,
                Target:   r.Object,
                Relation: r.Predicate,
                Weight:   1.0,
            })
        }
    }

    fmt.Printf("Graph: %d nodes, %d edges\n",
        len(g.Nodes()), len(g.Edges()))

    neighbors, _ := g.GetNeighbors("Alice")
    for _, n := range neighbors {
        fmt.Printf("  Alice -> %s (%s)\n", n.ID, n.Type)
    }
}
```

## 3. Memory Leak Detection

Monitor heap growth and goroutine leaks at runtime.

```go
package main

import (
    "context"
    "fmt"
    "time"

    "digital.vasic.memory/pkg/memory"
)

func main() {
    monitor := memory.NewMemoryMonitor(
        2*time.Second, // sampling interval
        2.0,           // alert if heap doubles
    )

    monitor.SetAlertCallback(func(report memory.LeakReport) {
        fmt.Printf("ALERT: heap growth ratio %.2f, goroutines %d\n",
            report.HeapGrowthRatio, report.GoroutineCount)
    })

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    monitor.Start(ctx)
    defer monitor.Stop()

    // Read reports from the channel
    for report := range monitor.Reports() {
        fmt.Printf("Heap: %d MB, Objects: %d, Leak: %v\n",
            report.HeapAlloc/1024/1024,
            report.HeapObjects,
            report.PotentialLeak)
    }
}
```
