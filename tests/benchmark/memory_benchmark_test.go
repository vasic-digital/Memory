package benchmark

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"digital.vasic.memory/pkg/entity"
	"digital.vasic.memory/pkg/graph"
	"digital.vasic.memory/pkg/mem0"
	"digital.vasic.memory/pkg/store"
)

func BenchmarkInMemoryStoreAdd(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Add(ctx, &store.Memory{
			ID:      fmt.Sprintf("mem-%d", i),
			Content: fmt.Sprintf("benchmark memory content %d", i),
			Scope:   store.ScopeUser,
		})
	}
}

func BenchmarkInMemoryStoreSearch(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	for i := 0; i < 500; i++ {
		_ = s.Add(ctx, &store.Memory{
			ID:      fmt.Sprintf("mem-%d", i),
			Content: fmt.Sprintf("document %d about topic %d with keywords", i, i%20),
			Scope:   store.ScopeUser,
		})
	}

	opts := store.DefaultSearchOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Search(ctx, "topic keywords", opts)
	}
}

func BenchmarkMem0ManagerAdd(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	backend := store.NewInMemoryStore()
	mgr := mem0.NewManager(backend, mem0.DefaultConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Add(ctx, &store.Memory{
			Content: fmt.Sprintf("managed memory %d", i),
			Scope:   store.ScopeUser,
		})
	}
}

func BenchmarkCalculateImportance(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	mem := &store.Memory{
		Content:   strings.Repeat("word ", 100),
		Metadata:  map[string]any{"key": "value", "source": "test"},
		Embedding: make([]float32, 128),
		Scope:     store.ScopeGlobal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem0.CalculateImportance(mem)
	}
}

func BenchmarkEntityExtraction(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	extractor := entity.NewPatternExtractor()
	text := "Alice Smith works at Big Corp. Contact alice@bigcorp.com or " +
		"visit https://bigcorp.com. Bob Jones uses Python."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = extractor.Extract(text)
	}
}

func BenchmarkGraphAddNodeAndEdge(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := graph.NewInMemoryGraph()
		_ = g.AddNode(graph.Node{ID: "a", Type: "test"})
		_ = g.AddNode(graph.Node{ID: "b", Type: "test"})
		_ = g.AddEdge(graph.Edge{Source: "a", Target: "b", Relation: "link"})
	}
}

func BenchmarkGraphShortestPath(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	g := graph.NewInMemoryGraph()
	const chainLen = 50
	for i := 0; i < chainLen; i++ {
		_ = g.AddNode(graph.Node{ID: fmt.Sprintf("n%d", i), Type: "node"})
	}
	for i := 0; i < chainLen-1; i++ {
		_ = g.AddEdge(graph.Edge{
			Source: fmt.Sprintf("n%d", i), Target: fmt.Sprintf("n%d", i+1),
			Relation: "next",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.ShortestPath("n0", fmt.Sprintf("n%d", chainLen-1))
	}
}

func BenchmarkGraphSubgraph(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test in short mode")
	}

	g := graph.NewInMemoryGraph()
	for i := 0; i < 100; i++ {
		_ = g.AddNode(graph.Node{ID: fmt.Sprintf("n%d", i), Type: "node"})
	}
	for i := 1; i < 100; i++ {
		parent := (i - 1) / 2
		_ = g.AddEdge(graph.Edge{
			Source:   fmt.Sprintf("n%d", parent),
			Target:   fmt.Sprintf("n%d", i),
			Relation: "child",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = g.Subgraph("n0", 4)
	}
}
