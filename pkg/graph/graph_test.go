package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Node/Edge structs ---

func TestNode_Struct(t *testing.T) {
	n := Node{
		ID:         "n1",
		Type:       "person",
		Properties: map[string]any{"name": "Alice"},
	}
	assert.Equal(t, "n1", n.ID)
	assert.Equal(t, "person", n.Type)
	assert.Equal(t, "Alice", n.Properties["name"])
}

func TestEdge_Struct(t *testing.T) {
	e := Edge{
		Source:   "n1",
		Target:   "n2",
		Relation: "knows",
		Weight:   0.9,
	}
	assert.Equal(t, "n1", e.Source)
	assert.Equal(t, "n2", e.Target)
	assert.Equal(t, "knows", e.Relation)
	assert.Equal(t, 0.9, e.Weight)
}

// --- NewInMemoryGraph ---

func TestNewInMemoryGraph(t *testing.T) {
	g := NewInMemoryGraph()
	require.NotNil(t, g)
	assert.Empty(t, g.nodes)
	assert.Empty(t, g.edges)
	assert.Empty(t, g.adj)
}

// --- AddNode ---

func TestInMemoryGraph_AddNode(t *testing.T) {
	tests := []struct {
		name      string
		node      Node
		setup     func(g *InMemoryGraph)
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Success",
			node:      Node{ID: "n1", Type: "person"},
			expectErr: false,
		},
		{
			name:      "EmptyID",
			node:      Node{ID: "", Type: "person"},
			expectErr: true,
			errMsg:    "node ID cannot be empty",
		},
		{
			name: "Duplicate",
			node: Node{ID: "n1", Type: "person"},
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1", Type: "person"})
			},
			expectErr: true,
			errMsg:    "node already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			err := g.AddNode(tt.node)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- AddEdge ---

func TestInMemoryGraph_AddEdge(t *testing.T) {
	tests := []struct {
		name      string
		edge      Edge
		setup     func(g *InMemoryGraph)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			edge: Edge{Source: "n1", Target: "n2", Relation: "knows", Weight: 1.0},
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1"})
				_ = g.AddNode(Node{ID: "n2"})
			},
			expectErr: false,
		},
		{
			name: "MissingSource",
			edge: Edge{Source: "missing", Target: "n2", Relation: "knows"},
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n2"})
			},
			expectErr: true,
			errMsg:    "source node not found",
		},
		{
			name: "MissingTarget",
			edge: Edge{Source: "n1", Target: "missing", Relation: "knows"},
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1"})
			},
			expectErr: true,
			errMsg:    "target node not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			err := g.AddEdge(tt.edge)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- GetNode ---

func TestInMemoryGraph_GetNode(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setup     func(g *InMemoryGraph)
		expectErr bool
	}{
		{
			name: "Found",
			id:   "n1",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1", Type: "person"})
			},
			expectErr: false,
		},
		{
			name:      "NotFound",
			id:        "nonexistent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			node, err := g.GetNode(tt.id)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, node.ID)
			}
		})
	}
}

// --- GetNeighbors ---

func TestInMemoryGraph_GetNeighbors(t *testing.T) {
	tests := []struct {
		name      string
		nodeID    string
		setup     func(g *InMemoryGraph)
		expected  int
		expectErr bool
	}{
		{
			name:   "WithNeighbors",
			nodeID: "n1",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1"})
				_ = g.AddNode(Node{ID: "n2"})
				_ = g.AddNode(Node{ID: "n3"})
				_ = g.AddEdge(Edge{Source: "n1", Target: "n2"})
				_ = g.AddEdge(Edge{Source: "n1", Target: "n3"})
			},
			expected:  2,
			expectErr: false,
		},
		{
			name:   "NoNeighbors",
			nodeID: "n1",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1"})
			},
			expected:  0,
			expectErr: false,
		},
		{
			name:      "NodeNotFound",
			nodeID:    "missing",
			expectErr: true,
		},
		{
			name:   "DuplicateEdges",
			nodeID: "n1",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "n1"})
				_ = g.AddNode(Node{ID: "n2"})
				_ = g.AddEdge(Edge{Source: "n1", Target: "n2", Relation: "a"})
				_ = g.AddEdge(Edge{Source: "n1", Target: "n2", Relation: "b"})
			},
			expected:  1, // deduplicated
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			neighbors, err := g.GetNeighbors(tt.nodeID)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, neighbors, tt.expected)
			}
		})
	}
}

// --- ShortestPath ---

func TestInMemoryGraph_ShortestPath(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		setup     func(g *InMemoryGraph)
		expected  []string
		expectErr bool
		errMsg    string
	}{
		{
			name: "DirectPath",
			from: "A", to: "B",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
			},
			expected: []string{"A", "B"},
		},
		{
			name: "MultiHopPath",
			from: "A", to: "C",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddNode(Node{ID: "C"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
				_ = g.AddEdge(Edge{Source: "B", Target: "C"})
			},
			expected: []string{"A", "B", "C"},
		},
		{
			name: "SameNode",
			from: "A", to: "A",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
			},
			expected: []string{"A"},
		},
		{
			name: "NoPath",
			from: "A", to: "C",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddNode(Node{ID: "C"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
				// No edge from B to C
			},
			expectErr: true,
			errMsg:    "no path",
		},
		{
			name:      "FromNodeNotFound",
			from:      "missing", to: "A",
			expectErr: true,
			errMsg:    "node not found",
		},
		{
			name: "ToNodeNotFound",
			from: "A", to: "missing",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
			},
			expectErr: true,
			errMsg:    "node not found",
		},
		{
			name: "ShortestOfMultiplePaths",
			from: "A", to: "D",
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddNode(Node{ID: "C"})
				_ = g.AddNode(Node{ID: "D"})
				// Long path: A -> B -> C -> D
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
				_ = g.AddEdge(Edge{Source: "B", Target: "C"})
				_ = g.AddEdge(Edge{Source: "C", Target: "D"})
				// Short path: A -> D
				_ = g.AddEdge(Edge{Source: "A", Target: "D"})
			},
			expected: []string{"A", "D"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			path, err := g.ShortestPath(tt.from, tt.to)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, path)
			}
		})
	}
}

// --- Subgraph ---

func TestInMemoryGraph_Subgraph(t *testing.T) {
	tests := []struct {
		name          string
		startID       string
		maxDepth      int
		setup         func(g *InMemoryGraph)
		expectedNodes int
		expectedEdges int
		expectErr     bool
	}{
		{
			name:     "DepthZero",
			startID:  "A",
			maxDepth: 0,
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
			},
			expectedNodes: 1,
			expectedEdges: 0,
		},
		{
			name:     "DepthOne",
			startID:  "A",
			maxDepth: 1,
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddNode(Node{ID: "C"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
				_ = g.AddEdge(Edge{Source: "B", Target: "C"})
			},
			expectedNodes: 2,
			expectedEdges: 1,
		},
		{
			name:     "DepthTwo",
			startID:  "A",
			maxDepth: 2,
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
				_ = g.AddNode(Node{ID: "B"})
				_ = g.AddNode(Node{ID: "C"})
				_ = g.AddEdge(Edge{Source: "A", Target: "B"})
				_ = g.AddEdge(Edge{Source: "B", Target: "C"})
			},
			expectedNodes: 3,
			expectedEdges: 2,
		},
		{
			name:      "StartNotFound",
			startID:   "missing",
			maxDepth:  1,
			expectErr: true,
		},
		{
			name:     "IsolatedNode",
			startID:  "A",
			maxDepth: 5,
			setup: func(g *InMemoryGraph) {
				_ = g.AddNode(Node{ID: "A"})
			},
			expectedNodes: 1,
			expectedEdges: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewInMemoryGraph()
			if tt.setup != nil {
				tt.setup(g)
			}

			nodes, edges, err := g.Subgraph(tt.startID, tt.maxDepth)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, nodes, tt.expectedNodes)
				assert.Len(t, edges, tt.expectedEdges)
			}
		})
	}
}

// --- Nodes/Edges ---

func TestInMemoryGraph_Nodes(t *testing.T) {
	g := NewInMemoryGraph()
	assert.Empty(t, g.Nodes())

	_ = g.AddNode(Node{ID: "n1"})
	_ = g.AddNode(Node{ID: "n2"})
	assert.Len(t, g.Nodes(), 2)
}

func TestInMemoryGraph_Edges(t *testing.T) {
	g := NewInMemoryGraph()
	assert.Empty(t, g.Edges())

	_ = g.AddNode(Node{ID: "n1"})
	_ = g.AddNode(Node{ID: "n2"})
	_ = g.AddEdge(Edge{Source: "n1", Target: "n2"})
	assert.Len(t, g.Edges(), 1)
}

// --- Concurrency ---

func TestInMemoryGraph_ConcurrentAccess(t *testing.T) {
	g := NewInMemoryGraph()

	// Pre-create nodes
	for i := 0; i < 50; i++ {
		_ = g.AddNode(Node{ID: fmt.Sprintf("n%d", i)})
	}

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			id := fmt.Sprintf("n%d", idx)
			_, _ = g.GetNode(id)
			_, _ = g.GetNeighbors(id)
			_ = g.Nodes()
			_ = g.Edges()
		}(i)
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}

// --- Interface compliance ---

func TestInMemoryGraph_ImplementsGraph(t *testing.T) {
	var _ Graph = (*InMemoryGraph)(nil)
}

// --- reconstructPath ---

func TestReconstructPath(t *testing.T) {
	tests := []struct {
		name     string
		parent   map[string]string
		from     string
		to       string
		expected []string
	}{
		{
			name:     "SingleHop",
			parent:   map[string]string{"B": "A"},
			from:     "A",
			to:       "B",
			expected: []string{"A", "B"},
		},
		{
			name:     "MultiHop",
			parent:   map[string]string{"B": "A", "C": "B", "D": "C"},
			from:     "A",
			to:       "D",
			expected: []string{"A", "B", "C", "D"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconstructPath(tt.parent, tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}
