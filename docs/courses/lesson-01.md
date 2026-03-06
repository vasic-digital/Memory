# Lesson 1: Memory Store Fundamentals

## Objectives

- Understand the `MemoryStore` interface and the `Memory` struct
- Use `InMemoryStore` for CRUD, search, and scoped listing
- Configure search and list options

## Concepts

### The Memory Struct

A `Memory` holds content text, a scope (user, session, conversation, global), metadata, an optional embedding vector, and a score:

```go
type Memory struct {
    ID        string
    Content   string
    Metadata  map[string]any
    Scope     Scope
    CreatedAt time.Time
    UpdatedAt time.Time
    Score     float64
    Embedding []float32
}
```

### The MemoryStore Interface

All storage backends implement six methods:

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

### Scopes

Four built-in scopes control memory visibility:

- `ScopeUser` -- visible to a specific user
- `ScopeSession` -- visible within a session
- `ScopeConversation` -- visible within a conversation
- `ScopeGlobal` -- visible to everyone

## Code Walkthrough

### Creating a store and adding memories

```go
s := store.NewInMemoryStore()
ctx := context.Background()

err := s.Add(ctx, &store.Memory{
    Content:  "The server runs on port 8080",
    Scope:    store.ScopeGlobal,
    Metadata: map[string]any{"source": "config"},
})
// ID and timestamps are set automatically
```

### Searching

```go
results, err := s.Search(ctx, "server port", &store.SearchOptions{
    TopK:     5,
    MinScore: 0.5,
    Scope:    store.ScopeGlobal,
})
```

The in-memory implementation scores by word overlap: if both "server" and "port" appear in the content, the score is 1.0 (2/2 query words matched).

### Listing with pagination

```go
page, err := s.List(ctx, store.ScopeUser, &store.ListOptions{
    Offset:  0,
    Limit:   25,
    OrderBy: "created_at", // or "updated_at", "score"
})
```

## Summary

The `store` package defines a clean CRUD+search interface. `InMemoryStore` is the default implementation, suitable for development and testing. For production, implement `MemoryStore` against a persistent or vector database.
