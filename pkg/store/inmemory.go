package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryStore provides a thread-safe in-memory implementation of MemoryStore.
type InMemoryStore struct {
	memories map[string]*Memory
	mu       sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories: make(map[string]*Memory),
	}
}

// Add stores a new memory. If the memory has no ID, one is generated.
func (s *InMemoryStore) Add(ctx context.Context, memory *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}

	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	if memory.UpdatedAt.IsZero() {
		memory.UpdatedAt = now
	}

	// Store a copy to avoid external mutation
	stored := *memory
	s.memories[memory.ID] = &stored
	return nil
}

// Get retrieves a memory by ID.
func (s *InMemoryStore) Get(ctx context.Context, id string) (*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memory, exists := s.memories[id]
	if !exists {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	result := *memory
	return &result, nil
}

// Update modifies an existing memory.
func (s *InMemoryStore) Update(ctx context.Context, memory *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.memories[memory.ID]; !exists {
		return fmt.Errorf("memory not found: %s", memory.ID)
	}

	memory.UpdatedAt = time.Now()
	stored := *memory
	s.memories[memory.ID] = &stored
	return nil
}

// Delete removes a memory by ID.
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.memories[id]; !exists {
		return fmt.Errorf("memory not found: %s", id)
	}

	delete(s.memories, id)
	return nil
}

// Search returns memories matching the query string and options.
// It uses simple word-overlap scoring for the in-memory implementation.
func (s *InMemoryStore) Search(
	ctx context.Context,
	query string,
	opts *SearchOptions,
) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts == nil {
		opts = DefaultSearchOptions()
	}

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	var results []*Memory

	for _, memory := range s.memories {
		// Filter by scope
		if opts.Scope != "" && memory.Scope != opts.Scope {
			continue
		}

		// Filter by time range
		if opts.TimeRange != nil {
			if memory.CreatedAt.Before(opts.TimeRange.Start) ||
				memory.CreatedAt.After(opts.TimeRange.End) {
				continue
			}
		}

		// Calculate match score
		score := calculateMatchScore(queryWords, memory.Content)
		if score >= opts.MinScore {
			result := *memory
			result.Score = score
			results = append(results, &result)
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if opts.TopK > 0 && len(results) > opts.TopK {
		results = results[:opts.TopK]
	}

	return results, nil
}

// List returns memories matching the scope and options.
func (s *InMemoryStore) List(
	ctx context.Context,
	scope Scope,
	opts *ListOptions,
) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts == nil {
		opts = DefaultListOptions()
	}

	var results []*Memory

	for _, memory := range s.memories {
		if scope != "" && memory.Scope != scope {
			continue
		}
		if opts.Scope != "" && memory.Scope != opts.Scope {
			continue
		}
		result := *memory
		results = append(results, &result)
	}

	// Sort
	sortMemories(results, opts.OrderBy)

	// Pagination
	start := opts.Offset
	if start > len(results) {
		return []*Memory{}, nil
	}

	end := len(results)
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}

	return results[start:end], nil
}

// calculateMatchScore computes a word-overlap score between query words
// and content.
func calculateMatchScore(queryWords []string, content string) float64 {
	if len(queryWords) == 0 {
		return 0
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, word := range queryWords {
		if strings.Contains(contentLower, word) {
			matches++
		}
	}

	return float64(matches) / float64(len(queryWords))
}

// sortMemories sorts a slice of memories by the specified field.
func sortMemories(memories []*Memory, orderBy string) {
	sort.Slice(memories, func(i, j int) bool {
		switch orderBy {
		case "updated_at":
			return memories[i].UpdatedAt.Before(memories[j].UpdatedAt)
		case "score":
			return memories[i].Score > memories[j].Score
		default: // "created_at"
			return memories[i].CreatedAt.Before(memories[j].CreatedAt)
		}
	})
}
