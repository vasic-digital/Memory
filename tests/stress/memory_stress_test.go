package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"digital.vasic.memory/pkg/entity"
	"digital.vasic.memory/pkg/graph"
	"digital.vasic.memory/pkg/mem0"
	"digital.vasic.memory/pkg/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentMemoryAddAndSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	const writers = 50
	const readers = 50

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := s.Add(ctx, &store.Memory{
				ID:      fmt.Sprintf("mem-%d", id),
				Content: fmt.Sprintf("memory content %d about topic %d", id, id%5),
				Scope:   store.ScopeUser,
			})
			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = s.Search(ctx, fmt.Sprintf("topic %d", id%5),
				store.DefaultSearchOptions())
		}(i)
	}

	wg.Wait()

	// Verify all writes completed
	for i := 0; i < writers; i++ {
		mem, err := s.Get(ctx, fmt.Sprintf("mem-%d", i))
		require.NoError(t, err)
		assert.NotEmpty(t, mem.Content)
	}
}

func TestConcurrentManagerOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	backend := store.NewInMemoryStore()
	mgr := mem0.NewManager(backend, mem0.DefaultConfig())
	ctx := context.Background()

	var wg sync.WaitGroup
	const goroutines = 80

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mem := &store.Memory{
				Content:  fmt.Sprintf("concurrent memory %d about topic %d", id, id%10),
				Scope:    store.ScopeUser,
				Metadata: map[string]any{"writer": id},
			}
			err := mgr.Add(ctx, mem)
			assert.NoError(t, err)
			assert.NotEmpty(t, mem.ID)
		}(i)
	}

	wg.Wait()

	results, err := mgr.Search(ctx, "concurrent memory", &store.SearchOptions{
		TopK: 100,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestConcurrentGraphOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()
	const numNodes = 100

	// Add nodes sequentially (IDs must exist before edges)
	for i := 0; i < numNodes; i++ {
		require.NoError(t, g.AddNode(graph.Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: "test",
		}))
	}

	// Add edges concurrently
	var wg sync.WaitGroup
	for i := 0; i < numNodes-1; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = g.AddEdge(graph.Edge{
				Source:   fmt.Sprintf("node-%d", id),
				Target:   fmt.Sprintf("node-%d", id+1),
				Relation: "next",
				Weight:   1.0,
			})
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	const readers = 60
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("node-%d", id%numNodes)
			_, err := g.GetNode(nodeID)
			assert.NoError(t, err)
			_, _ = g.GetNeighbors(nodeID)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numNodes, len(g.Nodes()))
}

func TestConcurrentEntityExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	extractor := entity.NewPatternExtractor()

	var wg sync.WaitGroup
	const goroutines = 80

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			text := fmt.Sprintf(
				"User%d uses Go. Contact user%d@example.com. Visit https://example.com/%d",
				id, id, id,
			)
			entities, relations, err := extractor.Extract(text)
			assert.NoError(t, err)
			_ = entities
			_ = relations
		}(i)
	}

	wg.Wait()
}

func TestConcurrentSubgraphExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()

	// Build a tree-like graph
	for i := 0; i < 50; i++ {
		require.NoError(t, g.AddNode(graph.Node{
			ID:   fmt.Sprintf("n%d", i),
			Type: "node",
		}))
	}
	for i := 1; i < 50; i++ {
		parent := (i - 1) / 2
		_ = g.AddEdge(graph.Edge{
			Source:   fmt.Sprintf("n%d", parent),
			Target:   fmt.Sprintf("n%d", i),
			Relation: "child",
		})
	}

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("n%d", id%50)
			nodes, edges, err := g.Subgraph(nodeID, 3)
			assert.NoError(t, err)
			assert.NotEmpty(t, nodes)
			_ = edges
		}(i)
	}

	wg.Wait()
}

func TestConcurrentMemoryUpdateAndDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	// Pre-populate
	const count = 50
	for i := 0; i < count; i++ {
		require.NoError(t, s.Add(ctx, &store.Memory{
			ID:      fmt.Sprintf("mem-%d", i),
			Content: fmt.Sprintf("original content %d", i),
			Scope:   store.ScopeUser,
		}))
	}

	var wg sync.WaitGroup

	// Concurrent updates on first half
	for i := 0; i < count/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = s.Update(ctx, &store.Memory{
				ID:      fmt.Sprintf("mem-%d", id),
				Content: fmt.Sprintf("updated content %d", id),
				Scope:   store.ScopeUser,
			})
		}(i)
	}

	// Concurrent deletes on second half
	for i := count / 2; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = s.Delete(ctx, fmt.Sprintf("mem-%d", id))
		}(i)
	}

	wg.Wait()

	// Verify first half updated
	for i := 0; i < count/2; i++ {
		mem, err := s.Get(ctx, fmt.Sprintf("mem-%d", i))
		require.NoError(t, err)
		assert.Contains(t, mem.Content, "updated")
	}

	// Verify second half deleted
	for i := count / 2; i < count; i++ {
		_, err := s.Get(ctx, fmt.Sprintf("mem-%d", i))
		assert.Error(t, err)
	}
}
