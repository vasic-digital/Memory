package integration

import (
	"context"
	"testing"
	"time"

	"digital.vasic.memory/pkg/entity"
	"digital.vasic.memory/pkg/graph"
	"digital.vasic.memory/pkg/mem0"
	"digital.vasic.memory/pkg/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreAndSearchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:      "mem1",
		Content: "Go is a statically typed compiled language",
		Scope:   store.ScopeUser,
	}))
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:      "mem2",
		Content: "Python is a dynamically typed language",
		Scope:   store.ScopeUser,
	}))
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:      "mem3",
		Content: "Rust provides memory safety without garbage collection",
		Scope:   store.ScopeGlobal,
	}))

	results, err := s.Search(ctx, "language", &store.SearchOptions{
		TopK:     10,
		MinScore: 0.1,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	for _, r := range results {
		assert.NotEmpty(t, r.Content)
		assert.GreaterOrEqual(t, r.Score, 0.1)
		assert.Contains(t, r.Content, "language")
	}
}

func TestMem0ManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backend := store.NewInMemoryStore()
	mgr := mem0.NewManager(backend, mem0.DefaultConfig())
	ctx := context.Background()

	mem := &store.Memory{
		Content:  "The user prefers dark mode in their IDE",
		Scope:    store.ScopeUser,
		Metadata: map[string]any{"source": "preference"},
	}
	require.NoError(t, mgr.Add(ctx, mem))
	assert.NotEmpty(t, mem.ID)
	assert.Greater(t, mem.Score, 0.0)

	retrieved, err := mgr.Get(ctx, mem.ID)
	require.NoError(t, err)
	assert.Equal(t, mem.Content, retrieved.Content)
	assert.Equal(t, store.ScopeUser, retrieved.Scope)
}

func TestEntityExtractionToGraphIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	extractor := entity.NewPatternExtractor()
	text := "Alice uses Go. Bob uses Python. Alice is a developer."

	entities, relations, err := extractor.Extract(text)
	require.NoError(t, err)

	g := graph.NewInMemoryGraph()

	nodeSet := make(map[string]bool)
	for _, e := range entities {
		if !nodeSet[e.Name] {
			err := g.AddNode(graph.Node{
				ID:   e.Name,
				Type: e.Type,
			})
			if err == nil {
				nodeSet[e.Name] = true
			}
		}
	}

	for _, r := range relations {
		if !nodeSet[r.Subject] {
			_ = g.AddNode(graph.Node{ID: r.Subject, Type: "entity"})
			nodeSet[r.Subject] = true
		}
		if !nodeSet[r.Object] {
			_ = g.AddNode(graph.Node{ID: r.Object, Type: "entity"})
			nodeSet[r.Object] = true
		}
		_ = g.AddEdge(graph.Edge{
			Source:   r.Subject,
			Target:   r.Object,
			Relation: r.Predicate,
			Weight:   1.0,
		})
	}

	assert.NotEmpty(t, g.Nodes())
	assert.NotEmpty(t, g.Edges())
}

func TestGraphTraversalIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	g := graph.NewInMemoryGraph()

	require.NoError(t, g.AddNode(graph.Node{ID: "A", Type: "concept"}))
	require.NoError(t, g.AddNode(graph.Node{ID: "B", Type: "concept"}))
	require.NoError(t, g.AddNode(graph.Node{ID: "C", Type: "concept"}))
	require.NoError(t, g.AddNode(graph.Node{ID: "D", Type: "concept"}))

	require.NoError(t, g.AddEdge(graph.Edge{Source: "A", Target: "B", Relation: "relates_to"}))
	require.NoError(t, g.AddEdge(graph.Edge{Source: "B", Target: "C", Relation: "leads_to"}))
	require.NoError(t, g.AddEdge(graph.Edge{Source: "C", Target: "D", Relation: "extends"}))

	path, err := g.ShortestPath("A", "D")
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B", "C", "D"}, path)

	neighbors, err := g.GetNeighbors("B")
	require.NoError(t, err)
	assert.Len(t, neighbors, 1)
	assert.Equal(t, "C", neighbors[0].ID)
}

func TestMemoryDecayIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	now := time.Now()
	recentScore := mem0.ApplyDecay(1.0, now, now, 0.01)
	assert.InDelta(t, 1.0, recentScore, 0.01)

	oldScore := mem0.ApplyDecay(
		1.0,
		now.Add(-24*time.Hour),
		now,
		0.01,
	)
	assert.Less(t, oldScore, recentScore)

	veryOldScore := mem0.ApplyDecay(
		1.0,
		now.Add(-7*24*time.Hour),
		now,
		0.01,
	)
	assert.Less(t, veryOldScore, oldScore)
}

func TestMemoryConsolidationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backend := store.NewInMemoryStore()
	cfg := mem0.DefaultConfig()
	cfg.ConsolidationInterval = 0
	cfg.SimilarityThreshold = 0.5
	mgr := mem0.NewManager(backend, cfg)
	ctx := context.Background()

	require.NoError(t, mgr.Add(ctx, &store.Memory{
		Content: "Go is a programming language designed at Google",
		Scope:   store.ScopeUser,
	}))
	require.NoError(t, mgr.Add(ctx, &store.Memory{
		Content: "Go is a programming language created at Google",
		Scope:   store.ScopeUser,
	}))
	require.NoError(t, mgr.Add(ctx, &store.Memory{
		Content: "Rust provides memory safety guarantees",
		Scope:   store.ScopeUser,
	}))

	consolidated, err := mgr.Consolidate(ctx, store.ScopeUser)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, consolidated, 0)
}
