package e2e

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

func TestFullMemoryWorkflowE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	backend := store.NewInMemoryStore()
	mgr := mem0.NewManager(backend, mem0.DefaultConfig())
	ctx := context.Background()

	// Add memories
	memories := []*store.Memory{
		{Content: "User prefers Go for backend development", Scope: store.ScopeUser},
		{Content: "User uses PostgreSQL for relational databases", Scope: store.ScopeUser},
		{Content: "Project uses Docker containers for deployment", Scope: store.ScopeSession},
		{Content: "Redis is used for caching layer", Scope: store.ScopeGlobal},
	}

	for _, mem := range memories {
		require.NoError(t, mgr.Add(ctx, mem))
		assert.NotEmpty(t, mem.ID)
		assert.Greater(t, mem.Score, 0.0)
		assert.False(t, mem.CreatedAt.IsZero())
	}

	// Search
	results, err := mgr.Search(ctx, "database", store.DefaultSearchOptions())
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// Update
	memories[0].Content = "User strongly prefers Go for backend development"
	require.NoError(t, mgr.Update(ctx, memories[0]))

	updated, err := mgr.Get(ctx, memories[0].ID)
	require.NoError(t, err)
	assert.Contains(t, updated.Content, "strongly prefers")

	// Delete
	require.NoError(t, mgr.Delete(ctx, memories[3].ID))
	_, err = mgr.Get(ctx, memories[3].ID)
	assert.Error(t, err)

	// List by scope
	userMems, err := mgr.List(ctx, store.ScopeUser, store.DefaultListOptions())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(userMems), 2)
}

func TestEntityExtractionPipelineE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	text := "John Smith works at Acme Corp. Jane Doe uses Python. " +
		"Acme Corp has an office at https://acme.example.com. " +
		"Contact john@acme.com for details."

	extractor := entity.NewPatternExtractor()
	entities, relations, err := extractor.Extract(text)
	require.NoError(t, err)

	// Should find URL and email entities
	foundEmail := false
	foundURL := false
	for _, e := range entities {
		if e.Type == "email" {
			foundEmail = true
		}
		if e.Type == "url" {
			foundURL = true
		}
	}
	assert.True(t, foundEmail, "should extract email entity")
	assert.True(t, foundURL, "should extract URL entity")

	// Should find relations
	assert.NotEmpty(t, relations)
}

func TestKnowledgeGraphE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	g := graph.NewInMemoryGraph()

	nodes := []graph.Node{
		{ID: "go", Type: "language", Properties: map[string]any{"paradigm": "compiled"}},
		{ID: "python", Type: "language", Properties: map[string]any{"paradigm": "interpreted"}},
		{ID: "rust", Type: "language", Properties: map[string]any{"paradigm": "compiled"}},
		{ID: "backend", Type: "domain"},
		{ID: "data-science", Type: "domain"},
		{ID: "systems", Type: "domain"},
	}

	for _, n := range nodes {
		require.NoError(t, g.AddNode(n))
	}

	edges := []graph.Edge{
		{Source: "go", Target: "backend", Relation: "used_for", Weight: 0.9},
		{Source: "python", Target: "data-science", Relation: "used_for", Weight: 0.95},
		{Source: "rust", Target: "systems", Relation: "used_for", Weight: 0.9},
		{Source: "go", Target: "systems", Relation: "used_for", Weight: 0.7},
		{Source: "backend", Target: "data-science", Relation: "connects_to", Weight: 0.5},
	}

	for _, e := range edges {
		require.NoError(t, g.AddEdge(e))
	}

	assert.Len(t, g.Nodes(), 6)
	assert.Len(t, g.Edges(), 5)

	// Traverse
	goNeighbors, err := g.GetNeighbors("go")
	require.NoError(t, err)
	assert.Len(t, goNeighbors, 2)

	// Shortest path
	path, err := g.ShortestPath("go", "data-science")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Equal(t, "go", path[0])
	assert.Equal(t, "data-science", path[len(path)-1])

	// Subgraph
	subNodes, subEdges, err := g.Subgraph("go", 2)
	require.NoError(t, err)
	assert.NotEmpty(t, subNodes)
	assert.NotEmpty(t, subEdges)
}

func TestScopedMemorySearchE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	require.NoError(t, s.Add(ctx, &store.Memory{
		ID: "u1", Content: "user preference dark mode", Scope: store.ScopeUser,
	}))
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID: "s1", Content: "session context dark theme", Scope: store.ScopeSession,
	}))
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID: "g1", Content: "global config dark setting", Scope: store.ScopeGlobal,
	}))

	userResults, err := s.Search(ctx, "dark", &store.SearchOptions{
		TopK:  10,
		Scope: store.ScopeUser,
	})
	require.NoError(t, err)
	for _, r := range userResults {
		assert.Equal(t, store.ScopeUser, r.Scope)
	}

	allResults, err := s.Search(ctx, "dark", &store.SearchOptions{TopK: 10})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allResults), len(userResults))
}

func TestTimeRangeSearchE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	s := store.NewInMemoryStore()
	ctx := context.Background()

	now := time.Now()

	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:        "recent",
		Content:   "recent memory about coding",
		Scope:     store.ScopeUser,
		CreatedAt: now.Add(-1 * time.Hour),
	}))
	require.NoError(t, s.Add(ctx, &store.Memory{
		ID:        "old",
		Content:   "old memory about coding",
		Scope:     store.ScopeUser,
		CreatedAt: now.Add(-48 * time.Hour),
	}))

	results, err := s.Search(ctx, "coding", &store.SearchOptions{
		TopK: 10,
		TimeRange: &store.TimeRange{
			Start: now.Add(-2 * time.Hour),
			End:   now,
		},
	})
	require.NoError(t, err)

	for _, r := range results {
		assert.Equal(t, "recent", r.ID)
	}
}

func TestCustomEntityPatternsE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	extractor := entity.NewPatternExtractor().
		WithEntityPattern("ip_address", "ip", `(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`).
		WithRelationPattern("connects_to", "connects_to",
			`(?i)(\w+)\s+connects to\s+(\w+)`)

	text := "Server at 192.168.1.100 connects to database at 10.0.0.1"

	entities, relations, err := extractor.Extract(text)
	require.NoError(t, err)

	ipFound := false
	for _, e := range entities {
		if e.Type == "ip" {
			ipFound = true
		}
	}
	assert.True(t, ipFound, "should extract IP address entities")
	assert.NotEmpty(t, relations)
}
