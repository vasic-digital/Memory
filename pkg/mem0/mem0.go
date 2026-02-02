// Package mem0 provides Mem0-style memory management with consolidation,
// decay, and importance scoring on top of the core store interface.
package mem0

import (
	"context"
	"math"
	"strings"
	"sync"
	"time"

	"digital.vasic.memory/pkg/store"
	"github.com/google/uuid"
)

// Config configures the Mem0-style memory manager.
type Config struct {
	// DefaultScope is the scope assigned to memories without an explicit scope.
	DefaultScope store.Scope

	// MaxMemories is the maximum number of memories to retain per scope.
	// Zero means unlimited.
	MaxMemories int

	// ConsolidationInterval is the minimum time between consolidation runs.
	ConsolidationInterval time.Duration

	// DecayRate controls how quickly memory scores decay over time.
	// Value between 0 and 1. Zero disables decay.
	DecayRate float64

	// SimilarityThreshold is the minimum word-overlap score for two
	// memories to be considered similar during consolidation (0.0 to 1.0).
	SimilarityThreshold float64
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultScope:          store.ScopeUser,
		MaxMemories:           10000,
		ConsolidationInterval: 5 * time.Minute,
		DecayRate:             0.01,
		SimilarityThreshold:   0.7,
	}
}

// Manager implements store.MemoryStore with Mem0-style memory operations
// including consolidation, decay, and importance scoring.
type Manager struct {
	backend           store.MemoryStore
	config            *Config
	mu                sync.RWMutex
	lastConsolidation time.Time
}

// NewManager creates a new Mem0-style memory manager wrapping a backend store.
func NewManager(backend store.MemoryStore, config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	return &Manager{
		backend: backend,
		config:  config,
	}
}

// Add stores a new memory with automatic scope assignment and importance scoring.
func (m *Manager) Add(ctx context.Context, memory *store.Memory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}

	if memory.Scope == "" {
		memory.Scope = m.config.DefaultScope
	}

	if memory.Score == 0 {
		memory.Score = CalculateImportance(memory)
	}

	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	if memory.UpdatedAt.IsZero() {
		memory.UpdatedAt = now
	}

	return m.backend.Add(ctx, memory)
}

// Search returns memories matching the query with automatic decay applied.
func (m *Manager) Search(
	ctx context.Context,
	query string,
	opts *store.SearchOptions,
) ([]*store.Memory, error) {
	results, err := m.backend.Search(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	// Apply decay to scores
	if m.config.DecayRate > 0 {
		now := time.Now()
		for _, mem := range results {
			mem.Score = ApplyDecay(mem.Score, mem.CreatedAt, now, m.config.DecayRate)
		}
	}

	return results, nil
}

// Get retrieves a memory by ID.
func (m *Manager) Get(ctx context.Context, id string) (*store.Memory, error) {
	return m.backend.Get(ctx, id)
}

// Update modifies an existing memory and recalculates importance.
func (m *Manager) Update(ctx context.Context, memory *store.Memory) error {
	memory.UpdatedAt = time.Now()
	memory.Score = CalculateImportance(memory)
	return m.backend.Update(ctx, memory)
}

// Delete removes a memory by ID.
func (m *Manager) Delete(ctx context.Context, id string) error {
	return m.backend.Delete(ctx, id)
}

// List returns memories for the given scope with decay applied.
func (m *Manager) List(
	ctx context.Context,
	scope store.Scope,
	opts *store.ListOptions,
) ([]*store.Memory, error) {
	results, err := m.backend.List(ctx, scope, opts)
	if err != nil {
		return nil, err
	}

	if m.config.DecayRate > 0 {
		now := time.Now()
		for _, mem := range results {
			mem.Score = ApplyDecay(
				mem.Score, mem.CreatedAt, now, m.config.DecayRate,
			)
		}
	}

	return results, nil
}

// Consolidate merges similar memories within the given scope.
// It returns the number of memories consolidated.
func (m *Manager) Consolidate(ctx context.Context, scope store.Scope) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check cooldown
	if !m.lastConsolidation.IsZero() &&
		time.Since(m.lastConsolidation) < m.config.ConsolidationInterval {
		return 0, nil
	}

	memories, err := m.backend.List(ctx, scope, &store.ListOptions{
		Limit: m.config.MaxMemories,
	})
	if err != nil {
		return 0, err
	}

	consolidated := 0
	merged := make(map[string]bool)

	for i := 0; i < len(memories); i++ {
		if merged[memories[i].ID] {
			continue
		}
		for j := i + 1; j < len(memories); j++ {
			if merged[memories[j].ID] {
				continue
			}

			sim := wordOverlapSimilarity(
				memories[i].Content,
				memories[j].Content,
			)
			if sim >= m.config.SimilarityThreshold {
				// Merge j into i
				mergeMemories(memories[i], memories[j])
				merged[memories[j].ID] = true

				// Update the merged memory
				if err := m.backend.Update(ctx, memories[i]); err != nil {
					continue
				}

				// Delete the absorbed memory
				if err := m.backend.Delete(ctx, memories[j].ID); err != nil {
					continue
				}

				consolidated++
			}
		}
	}

	m.lastConsolidation = time.Now()
	return consolidated, nil
}

// CalculateImportance computes an importance score for a memory based on
// content length, metadata richness, and scope.
func CalculateImportance(memory *store.Memory) float64 {
	importance := 0.5 // Base

	// Boost for longer content
	if len(memory.Content) > 100 {
		importance += 0.1
	}
	if len(memory.Content) > 500 {
		importance += 0.1
	}

	// Boost for metadata richness
	if len(memory.Metadata) > 0 {
		importance += 0.1
	}

	// Boost for embeddings
	if len(memory.Embedding) > 0 {
		importance += 0.1
	}

	// Scope-based boost
	if memory.Scope == store.ScopeGlobal {
		importance += 0.1
	}

	// Cap at 1.0
	if importance > 1.0 {
		importance = 1.0
	}

	// Round to avoid floating point drift
	return math.Round(importance*100) / 100
}

// ApplyDecay reduces a score based on time elapsed since creation.
// Uses exponential decay: score * exp(-rate * hours).
func ApplyDecay(
	score float64,
	createdAt time.Time,
	now time.Time,
	rate float64,
) float64 {
	if rate <= 0 {
		return score
	}
	hours := now.Sub(createdAt).Hours()
	if hours < 0 {
		hours = 0
	}
	return score * math.Exp(-rate*hours)
}

// wordOverlapSimilarity computes Jaccard similarity based on word overlap.
func wordOverlapSimilarity(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 && len(wordsB) == 0 {
		return 1.0
	}
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	setA := make(map[string]bool, len(wordsA))
	for _, w := range wordsA {
		setA[w] = true
	}

	setB := make(map[string]bool, len(wordsB))
	for _, w := range wordsB {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA)
	for w := range setB {
		if !setA[w] {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// mergeMemories merges source into target by combining content and metadata.
func mergeMemories(target, source *store.Memory) {
	// Keep the longer content, or combine if both are substantial
	if len(source.Content) > len(target.Content) {
		target.Content = source.Content
	}

	// Merge metadata
	if target.Metadata == nil && source.Metadata != nil {
		target.Metadata = make(map[string]any)
	}
	for k, v := range source.Metadata {
		if _, exists := target.Metadata[k]; !exists {
			target.Metadata[k] = v
		}
	}

	// Keep higher score
	if source.Score > target.Score {
		target.Score = source.Score
	}

	// Use the earlier creation time
	if source.CreatedAt.Before(target.CreatedAt) {
		target.CreatedAt = source.CreatedAt
	}

	target.UpdatedAt = time.Now()
}
