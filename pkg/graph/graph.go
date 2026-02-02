// Package graph provides an in-memory knowledge graph with nodes, edges,
// neighbor traversal, shortest path, and subgraph extraction.
package graph

import (
	"fmt"
	"sync"
)

// Node represents a vertex in the knowledge graph.
type Node struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Edge represents a directed, weighted connection between two nodes.
type Edge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Relation string  `json:"relation"`
	Weight   float64 `json:"weight"`
}

// Graph defines the interface for knowledge graph operations.
type Graph interface {
	// AddNode adds a node to the graph.
	AddNode(node Node) error

	// AddEdge adds a directed edge between two existing nodes.
	AddEdge(edge Edge) error

	// GetNode returns a node by ID.
	GetNode(id string) (Node, error)

	// GetNeighbors returns all nodes directly connected to the given node.
	GetNeighbors(id string) ([]Node, error)

	// ShortestPath finds the shortest path (by hop count) between two nodes.
	// Returns the list of node IDs along the path, including start and end.
	ShortestPath(from, to string) ([]string, error)

	// Subgraph returns all nodes and edges reachable within maxDepth hops
	// from the starting node.
	Subgraph(startID string, maxDepth int) ([]Node, []Edge, error)

	// Nodes returns all nodes in the graph.
	Nodes() []Node

	// Edges returns all edges in the graph.
	Edges() []Edge
}

// InMemoryGraph is a thread-safe in-memory implementation of Graph
// using adjacency lists.
type InMemoryGraph struct {
	nodes map[string]Node
	edges []Edge
	// adjacency list: nodeID -> list of (targetID, edgeIndex)
	adj map[string][]adjEntry
	mu  sync.RWMutex
}

type adjEntry struct {
	targetID  string
	edgeIndex int
}

// NewInMemoryGraph creates a new empty in-memory graph.
func NewInMemoryGraph() *InMemoryGraph {
	return &InMemoryGraph{
		nodes: make(map[string]Node),
		edges: make([]Edge, 0),
		adj:   make(map[string][]adjEntry),
	}
}

// AddNode adds a node to the graph. Returns an error if the ID is empty
// or the node already exists.
func (g *InMemoryGraph) AddNode(node Node) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if _, exists := g.nodes[node.ID]; exists {
		return fmt.Errorf("node already exists: %s", node.ID)
	}

	g.nodes[node.ID] = node
	return nil
}

// AddEdge adds a directed edge. Both source and target nodes must exist.
func (g *InMemoryGraph) AddEdge(edge Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[edge.Source]; !exists {
		return fmt.Errorf("source node not found: %s", edge.Source)
	}
	if _, exists := g.nodes[edge.Target]; !exists {
		return fmt.Errorf("target node not found: %s", edge.Target)
	}

	idx := len(g.edges)
	g.edges = append(g.edges, edge)
	g.adj[edge.Source] = append(g.adj[edge.Source], adjEntry{
		targetID:  edge.Target,
		edgeIndex: idx,
	})

	return nil
}

// GetNode returns a node by ID.
func (g *InMemoryGraph) GetNode(id string) (Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[id]
	if !exists {
		return Node{}, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

// GetNeighbors returns all nodes directly connected from the given node.
func (g *InMemoryGraph) GetNeighbors(id string) ([]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[id]; !exists {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	entries := g.adj[id]
	seen := make(map[string]bool)
	var neighbors []Node

	for _, entry := range entries {
		if !seen[entry.targetID] {
			seen[entry.targetID] = true
			if node, exists := g.nodes[entry.targetID]; exists {
				neighbors = append(neighbors, node)
			}
		}
	}

	return neighbors, nil
}

// ShortestPath finds the shortest path by hop count using BFS.
// Returns node IDs from start to end, or an error if no path exists.
func (g *InMemoryGraph) ShortestPath(from, to string) ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[from]; !exists {
		return nil, fmt.Errorf("node not found: %s", from)
	}
	if _, exists := g.nodes[to]; !exists {
		return nil, fmt.Errorf("node not found: %s", to)
	}

	if from == to {
		return []string{from}, nil
	}

	// BFS
	visited := map[string]bool{from: true}
	parent := make(map[string]string)
	queue := []string{from}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, entry := range g.adj[current] {
			if !visited[entry.targetID] {
				visited[entry.targetID] = true
				parent[entry.targetID] = current
				if entry.targetID == to {
					return reconstructPath(parent, from, to), nil
				}
				queue = append(queue, entry.targetID)
			}
		}
	}

	return nil, fmt.Errorf("no path from %s to %s", from, to)
}

// Subgraph returns all nodes and edges reachable within maxDepth hops.
func (g *InMemoryGraph) Subgraph(
	startID string,
	maxDepth int,
) ([]Node, []Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[startID]; !exists {
		return nil, nil, fmt.Errorf("node not found: %s", startID)
	}

	visitedNodes := map[string]bool{startID: true}
	visitedEdges := make(map[int]bool)

	// BFS with depth tracking
	type bfsItem struct {
		id    string
		depth int
	}
	queue := []bfsItem{{id: startID, depth: 0}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= maxDepth {
			continue
		}

		for _, entry := range g.adj[item.id] {
			visitedEdges[entry.edgeIndex] = true
			if !visitedNodes[entry.targetID] {
				visitedNodes[entry.targetID] = true
				queue = append(queue, bfsItem{
					id:    entry.targetID,
					depth: item.depth + 1,
				})
			}
		}
	}

	var nodes []Node
	for id := range visitedNodes {
		nodes = append(nodes, g.nodes[id])
	}

	var edges []Edge
	for idx := range visitedEdges {
		edges = append(edges, g.edges[idx])
	}

	return nodes, edges, nil
}

// Nodes returns all nodes in the graph.
func (g *InMemoryGraph) Nodes() []Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// Edges returns all edges in the graph.
func (g *InMemoryGraph) Edges() []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]Edge, len(g.edges))
	copy(result, g.edges)
	return result
}

// reconstructPath rebuilds the path from BFS parent map.
func reconstructPath(parent map[string]string, from, to string) []string {
	var path []string
	current := to
	for current != from {
		path = append([]string{current}, path...)
		current = parent[current]
	}
	return append([]string{from}, path...)
}
