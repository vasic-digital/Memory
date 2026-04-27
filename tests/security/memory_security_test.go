package security

import (
	"context"
	"strings"
	"testing"

	"digital.vasic.memory/pkg/entity"
	"digital.vasic.memory/pkg/graph"
	"digital.vasic.memory/pkg/mem0"
	"digital.vasic.memory/pkg/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNonExistentMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteNonExistentMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	err := s.Delete(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateNonExistentMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	err := s.Update(ctx, &store.Memory{
		ID:      "nonexistent",
		Content: "updated content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEmptyGraphNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	err := g.AddNode(graph.Node{ID: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDuplicateGraphNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	require.NoError(t, g.AddNode(graph.Node{ID: "A"}))

	err := g.AddNode(graph.Node{ID: "A"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEdgeWithMissingNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	require.NoError(t, g.AddNode(graph.Node{ID: "A"}))

	err := g.AddEdge(graph.Edge{Source: "A", Target: "B"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	err = g.AddEdge(graph.Edge{Source: "X", Target: "A"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestShortestPathNoRoute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	require.NoError(t, g.AddNode(graph.Node{ID: "A"}))
	require.NoError(t, g.AddNode(graph.Node{ID: "B"}))
	// No edge between A and B

	_, err := g.ShortestPath("A", "B")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no path")
}

func TestShortestPathNonExistentNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	require.NoError(t, g.AddNode(graph.Node{ID: "A"}))

	_, err := g.ShortestPath("A", "nonexistent")
	assert.Error(t, err)

	_, err = g.ShortestPath("nonexistent", "A")
	assert.Error(t, err)
}

func TestLargeMemoryContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	largeContent := strings.Repeat("x", 5*1024*1024) // 5MB
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:      "large",
		Content: largeContent,
		Scope:   store.ScopeUser,
	}))

	retrieved, err := s.Get(ctx, "large")
	require.NoError(t, err)
	assert.Equal(t, len(largeContent), len(retrieved.Content))
}

func TestEntityExtractionEmptyInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	extractor := entity.NewPatternExtractor()

	entities, relations, err := extractor.Extract("")
	require.NoError(t, err)
	assert.Empty(t, entities)
	assert.Empty(t, relations)

	entities, relations, err = extractor.Extract("   ")
	require.NoError(t, err)
	assert.Empty(t, entities)
	assert.Empty(t, relations)
}

func TestImportanceCalculationEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	// Empty memory
	score := mem0.CalculateImportance(&store.Memory{})
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Maximally rich memory
	rich := &store.Memory{
		Content:   strings.Repeat("word ", 200),
		Metadata:  map[string]any{"key": "value"},
		Embedding: make([]float32, 128),
		Scope:     store.ScopeGlobal,
	}
	richScore := mem0.CalculateImportance(rich)
	assert.Greater(t, richScore, score)
	assert.LessOrEqual(t, richScore, 1.0)
}

func TestSubgraphNonExistentNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping security test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	_, _, err := g.Subgraph("nonexistent", 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
