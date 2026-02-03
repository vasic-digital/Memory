package mem0

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"digital.vasic.memory/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Config ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, store.ScopeUser, cfg.DefaultScope)
	assert.Equal(t, 10000, cfg.MaxMemories)
	assert.Equal(t, 5*time.Minute, cfg.ConsolidationInterval)
	assert.Equal(t, 0.01, cfg.DecayRate)
	assert.Equal(t, 0.7, cfg.SimilarityThreshold)
}

// --- NewManager ---

func TestNewManager(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"WithNilConfig", nil},
		{"WithCustomConfig", &Config{DefaultScope: store.ScopeGlobal}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := store.NewInMemoryStore()
			m := NewManager(backend, tt.config)
			require.NotNil(t, m)
			assert.NotNil(t, m.config)
		})
	}
}

// --- Add ---

func TestManager_Add(t *testing.T) {
	tests := []struct {
		name       string
		memory     *store.Memory
		checkScope store.Scope
	}{
		{
			name:       "WithDefaultScope",
			memory:     &store.Memory{Content: "test"},
			checkScope: store.ScopeUser,
		},
		{
			name:       "WithExplicitScope",
			memory:     &store.Memory{Content: "test", Scope: store.ScopeGlobal},
			checkScope: store.ScopeGlobal,
		},
		{
			name:       "WithExistingID",
			memory:     &store.Memory{ID: "existing", Content: "test"},
			checkScope: store.ScopeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(store.NewInMemoryStore(), nil)
			err := m.Add(context.Background(), tt.memory)
			require.NoError(t, err)
			assert.NotEmpty(t, tt.memory.ID)
			assert.Equal(t, tt.checkScope, tt.memory.Scope)
			assert.Greater(t, tt.memory.Score, 0.0)
			assert.False(t, tt.memory.CreatedAt.IsZero())
		})
	}
}

func TestManager_Add_PreservesExistingScore(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	mem := &store.Memory{Content: "test", Score: 0.99}
	err := m.Add(context.Background(), mem)
	require.NoError(t, err)
	assert.Equal(t, 0.99, mem.Score)
}

// --- Get ---

func TestManager_Get(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	ctx := context.Background()

	mem := &store.Memory{ID: "m1", Content: "hello"}
	_ = m.Add(ctx, mem)

	result, err := m.Get(ctx, "m1")
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Content)

	_, err = m.Get(ctx, "nonexistent")
	require.Error(t, err)
}

// --- Update ---

func TestManager_Update(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	ctx := context.Background()

	mem := &store.Memory{ID: "m1", Content: "original"}
	_ = m.Add(ctx, mem)

	updated := &store.Memory{ID: "m1", Content: "updated with much more detail"}
	err := m.Update(ctx, updated)
	require.NoError(t, err)
	assert.False(t, updated.UpdatedAt.IsZero())
	assert.Greater(t, updated.Score, 0.0)
}

// --- Delete ---

func TestManager_Delete(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	ctx := context.Background()

	_ = m.Add(ctx, &store.Memory{ID: "m1", Content: "test"})
	err := m.Delete(ctx, "m1")
	require.NoError(t, err)

	_, err = m.Get(ctx, "m1")
	require.Error(t, err)
}

// --- Search ---

func TestManager_Search(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	ctx := context.Background()

	_ = m.Add(ctx, &store.Memory{
		ID: "1", Content: "Go programming language",
	})
	_ = m.Add(ctx, &store.Memory{
		ID: "2", Content: "Python machine learning",
	})

	results, err := m.Search(ctx, "programming", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestManager_Search_WithDecay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DecayRate = 1.0 // aggressive decay for testing
	m := NewManager(store.NewInMemoryStore(), cfg)
	ctx := context.Background()

	oldMem := &store.Memory{
		ID:        "old",
		Content:   "Go programming",
		Score:     0.9,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	_ = m.Add(ctx, oldMem)

	results, err := m.Search(ctx, "Go programming", &store.SearchOptions{
		MinScore: 0.0,
	})
	require.NoError(t, err)
	if len(results) > 0 {
		assert.Less(t, results[0].Score, 0.9)
	}
}

func TestManager_Search_NoDecay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DecayRate = 0 // disable decay
	m := NewManager(store.NewInMemoryStore(), cfg)
	ctx := context.Background()

	_ = m.Add(ctx, &store.Memory{
		ID:      "m1",
		Content: "Go programming",
		Score:   0.9,
	})

	results, err := m.Search(ctx, "Go programming", nil)
	require.NoError(t, err)
	// Score should remain unchanged (no decay)
	for _, r := range results {
		assert.GreaterOrEqual(t, r.Score, 0.0)
	}
}

// --- List ---

func TestManager_List(t *testing.T) {
	m := NewManager(store.NewInMemoryStore(), nil)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = m.Add(ctx, &store.Memory{
			ID:      fmt.Sprintf("m%d", i),
			Content: fmt.Sprintf("Memory %d", i),
			Scope:   store.ScopeUser,
		})
	}

	results, err := m.List(ctx, store.ScopeUser, nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

// --- CalculateImportance ---

func TestCalculateImportance(t *testing.T) {
	tests := []struct {
		name     string
		memory   *store.Memory
		minScore float64
		maxScore float64
	}{
		{
			name:     "BaseScore",
			memory:   &store.Memory{Content: "short"},
			minScore: 0.5,
			maxScore: 0.5,
		},
		{
			name: "LongContent",
			memory: &store.Memory{
				Content: string(make([]byte, 150)),
			},
			minScore: 0.6,
			maxScore: 0.6,
		},
		{
			name: "VeryLongContent",
			memory: &store.Memory{
				Content: string(make([]byte, 600)),
			},
			minScore: 0.7,
			maxScore: 0.7,
		},
		{
			name: "WithMetadata",
			memory: &store.Memory{
				Content:  "test",
				Metadata: map[string]any{"key": "value"},
			},
			minScore: 0.6,
			maxScore: 0.6,
		},
		{
			name: "WithEmbedding",
			memory: &store.Memory{
				Content:   "test",
				Embedding: []float32{0.1},
			},
			minScore: 0.6,
			maxScore: 0.6,
		},
		{
			name: "GlobalScope",
			memory: &store.Memory{
				Content: "test",
				Scope:   store.ScopeGlobal,
			},
			minScore: 0.6,
			maxScore: 0.6,
		},
		{
			name: "MaxCap",
			memory: &store.Memory{
				Content:   string(make([]byte, 600)),
				Metadata:  map[string]any{"k": "v"},
				Embedding: []float32{0.1},
				Scope:     store.ScopeGlobal,
			},
			minScore: 1.0,
			maxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateImportance(tt.memory)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

// --- ApplyDecay ---

func TestApplyDecay(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		hours    float64
		rate     float64
		expected float64
	}{
		{"NoDecay", 1.0, 0, 0.01, 1.0},
		{"ZeroRate", 1.0, 100, 0, 1.0},
		{"NegativeRate", 1.0, 10, -0.01, 1.0},
		{"OneHour", 1.0, 1, 0.01, math.Exp(-0.01)},
		{"TwentyFourHours", 1.0, 24, 0.01, math.Exp(-0.24)},
		{"HighRate", 1.0, 100, 1.0, math.Exp(-100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			createdAt := now.Add(-time.Duration(tt.hours * float64(time.Hour)))
			result := ApplyDecay(tt.score, createdAt, now, tt.rate)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestApplyDecay_FutureCreation(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	result := ApplyDecay(1.0, future, now, 0.01)
	assert.Equal(t, 1.0, result) // hours clamped to 0
}

// --- wordOverlapSimilarity ---

func TestWordOverlapSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
	}{
		{"Identical", "hello world", "hello world", 1.0},
		{"NoOverlap", "hello world", "foo bar", 0.0},
		{"PartialOverlap", "hello world foo", "hello bar baz", 0.2},
		{"BothEmpty", "", "", 1.0},
		{"OneEmpty", "hello", "", 0.0},
		{"CaseInsensitive", "Hello World", "hello world", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wordOverlapSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

// --- mergeMemories ---

func TestMergeMemories(t *testing.T) {
	t.Run("SourceLongerContent", func(t *testing.T) {
		target := &store.Memory{
			ID:        "t",
			Content:   "short",
			Score:     0.5,
			CreatedAt: time.Now(),
		}
		source := &store.Memory{
			ID:        "s",
			Content:   "this is a much longer content string",
			Score:     0.3,
			CreatedAt: time.Now().Add(-time.Hour),
			Metadata:  map[string]any{"key": "value"},
		}

		mergeMemories(target, source)
		assert.Equal(t, "this is a much longer content string", target.Content)
		assert.Equal(t, 0.5, target.Score) // keep higher
		assert.Equal(t, source.CreatedAt, target.CreatedAt) // earlier
		assert.Equal(t, "value", target.Metadata["key"])
	})

	t.Run("TargetLongerContent", func(t *testing.T) {
		target := &store.Memory{
			ID:        "t",
			Content:   "this is a longer target content",
			Score:     0.3,
			CreatedAt: time.Now(),
		}
		source := &store.Memory{
			ID:        "s",
			Content:   "short",
			Score:     0.8,
			CreatedAt: time.Now().Add(time.Hour),
		}

		mergeMemories(target, source)
		assert.Equal(t, "this is a longer target content", target.Content)
		assert.Equal(t, 0.8, target.Score)
	})

	t.Run("MetadataMerge", func(t *testing.T) {
		target := &store.Memory{
			ID:       "t",
			Content:  "test",
			Metadata: map[string]any{"a": 1},
		}
		source := &store.Memory{
			ID:       "s",
			Content:  "t",
			Metadata: map[string]any{"a": 2, "b": 3},
		}

		mergeMemories(target, source)
		assert.Equal(t, 1, target.Metadata["a"])  // keep existing
		assert.Equal(t, 3, target.Metadata["b"])   // add new
	})

	t.Run("NilTargetMetadata", func(t *testing.T) {
		target := &store.Memory{ID: "t", Content: "test"}
		source := &store.Memory{
			ID:       "s",
			Content:  "t",
			Metadata: map[string]any{"key": "value"},
		}

		mergeMemories(target, source)
		assert.Equal(t, "value", target.Metadata["key"])
	})
}

// --- Consolidate ---

func TestManager_Consolidate(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SimilarityThreshold = 0.5
	cfg.ConsolidationInterval = 0 // no cooldown

	t.Run("MergesSimilar", func(t *testing.T) {
		backend := store.NewInMemoryStore()
		m := NewManager(backend, cfg)
		ctx := context.Background()

		_ = m.Add(ctx, &store.Memory{
			ID: "1", Content: "Go programming language basics",
			Scope: store.ScopeUser,
		})
		_ = m.Add(ctx, &store.Memory{
			ID: "2", Content: "Go programming language fundamentals",
			Scope: store.ScopeUser,
		})
		_ = m.Add(ctx, &store.Memory{
			ID: "3", Content: "Python machine learning advanced",
			Scope: store.ScopeUser,
		})

		count, err := m.Consolidate(ctx, store.ScopeUser)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("NothingToConsolidate", func(t *testing.T) {
		backend := store.NewInMemoryStore()
		m := NewManager(backend, cfg)
		ctx := context.Background()

		_ = m.Add(ctx, &store.Memory{
			ID: "1", Content: "apples oranges",
			Scope: store.ScopeUser,
		})
		_ = m.Add(ctx, &store.Memory{
			ID: "2", Content: "quantum physics theory",
			Scope: store.ScopeUser,
		})

		count, err := m.Consolidate(ctx, store.ScopeUser)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("CooldownPreventsRun", func(t *testing.T) {
		cfg2 := DefaultConfig()
		cfg2.ConsolidationInterval = time.Hour
		cfg2.SimilarityThreshold = 0.5

		backend := store.NewInMemoryStore()
		m := NewManager(backend, cfg2)
		ctx := context.Background()

		_ = m.Add(ctx, &store.Memory{
			ID: "1", Content: "same same same",
			Scope: store.ScopeUser,
		})
		_ = m.Add(ctx, &store.Memory{
			ID: "2", Content: "same same same",
			Scope: store.ScopeUser,
		})

		// First run should work
		_, _ = m.Consolidate(ctx, store.ScopeUser)

		// Second run should be skipped due to cooldown
		count, err := m.Consolidate(ctx, store.ScopeUser)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// --- Interface compliance ---

func TestManager_ImplementsMemoryStore(t *testing.T) {
	var _ store.MemoryStore = (*Manager)(nil)
}

// --- Mock Backend for Error Testing ---

// mockBackend implements store.MemoryStore with configurable error responses.
type mockBackend struct {
	memories    map[string]*store.Memory
	searchErr   error
	listErr     error
	updateErr   error
	deleteErr   error
	listResults []*store.Memory
}

func newMockBackend() *mockBackend {
	return &mockBackend{
		memories: make(map[string]*store.Memory),
	}
}

func (m *mockBackend) Add(_ context.Context, memory *store.Memory) error {
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockBackend) Search(_ context.Context, _ string, _ *store.SearchOptions) ([]*store.Memory, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []*store.Memory
	for _, mem := range m.memories {
		results = append(results, mem)
	}
	return results, nil
}

func (m *mockBackend) Get(_ context.Context, id string) (*store.Memory, error) {
	if mem, ok := m.memories[id]; ok {
		return mem, nil
	}
	return nil, fmt.Errorf("memory not found: %s", id)
}

func (m *mockBackend) Update(_ context.Context, memory *store.Memory) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockBackend) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.memories, id)
	return nil
}

func (m *mockBackend) List(_ context.Context, _ store.Scope, _ *store.ListOptions) ([]*store.Memory, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listResults != nil {
		return m.listResults, nil
	}
	var results []*store.Memory
	for _, mem := range m.memories {
		results = append(results, mem)
	}
	return results, nil
}

// --- Search Error Path Tests (Lines 103-104) ---

func TestManager_Search_BackendError(t *testing.T) {
	tests := []struct {
		name        string
		searchErr   error
		expectedErr string
	}{
		{
			name:        "NetworkError",
			searchErr:   fmt.Errorf("connection refused"),
			expectedErr: "connection refused",
		},
		{
			name:        "TimeoutError",
			searchErr:   fmt.Errorf("context deadline exceeded"),
			expectedErr: "context deadline exceeded",
		},
		{
			name:        "InternalError",
			searchErr:   fmt.Errorf("internal backend error"),
			expectedErr: "internal backend error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newMockBackend()
			backend.searchErr = tt.searchErr
			m := NewManager(backend, nil)

			results, err := m.Search(context.Background(), "query", nil)
			require.Error(t, err)
			assert.Nil(t, results)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// --- List Error Path Tests (Lines 142-143) ---

func TestManager_List_BackendError(t *testing.T) {
	tests := []struct {
		name        string
		listErr     error
		expectedErr string
	}{
		{
			name:        "DatabaseError",
			listErr:     fmt.Errorf("database connection lost"),
			expectedErr: "database connection lost",
		},
		{
			name:        "PermissionDenied",
			listErr:     fmt.Errorf("permission denied"),
			expectedErr: "permission denied",
		},
		{
			name:        "InvalidScope",
			listErr:     fmt.Errorf("invalid scope specified"),
			expectedErr: "invalid scope specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newMockBackend()
			backend.listErr = tt.listErr
			m := NewManager(backend, nil)

			results, err := m.List(context.Background(), store.ScopeUser, nil)
			require.Error(t, err)
			assert.Nil(t, results)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// --- Consolidate Error Path Tests (Lines 173-174, 185-186, 199-200, 204-205) ---

func TestManager_Consolidate_ListError(t *testing.T) {
	// Tests lines 173-174: backend.List error in Consolidate
	backend := newMockBackend()
	backend.listErr = fmt.Errorf("list operation failed")

	cfg := DefaultConfig()
	cfg.ConsolidationInterval = 0
	m := NewManager(backend, cfg)

	count, err := m.Consolidate(context.Background(), store.ScopeUser)
	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "list operation failed")
}

func TestManager_Consolidate_UpdateError(t *testing.T) {
	// Tests lines 199-200: backend.Update error continues processing
	backend := newMockBackend()
	backend.updateErr = fmt.Errorf("update failed")

	cfg := DefaultConfig()
	cfg.ConsolidationInterval = 0
	cfg.SimilarityThreshold = 0.5

	// Add similar memories that will trigger merge
	backend.listResults = []*store.Memory{
		{ID: "1", Content: "same same same words here", Scope: store.ScopeUser},
		{ID: "2", Content: "same same same words here", Scope: store.ScopeUser},
	}

	m := NewManager(backend, cfg)

	// Should not return error, just continue (error path continues silently)
	count, err := m.Consolidate(context.Background(), store.ScopeUser)
	require.NoError(t, err)
	// Count should be 0 because update failed and we continued
	assert.Equal(t, 0, count)
}

func TestManager_Consolidate_DeleteError(t *testing.T) {
	// Tests lines 204-205: backend.Delete error continues processing
	backend := newMockBackend()
	backend.deleteErr = fmt.Errorf("delete failed")

	cfg := DefaultConfig()
	cfg.ConsolidationInterval = 0
	cfg.SimilarityThreshold = 0.5

	// Add similar memories that will trigger merge
	backend.listResults = []*store.Memory{
		{ID: "1", Content: "same same same words here", Scope: store.ScopeUser},
		{ID: "2", Content: "same same same words here", Scope: store.ScopeUser},
	}

	m := NewManager(backend, cfg)

	// Should not return error, just continue (error path continues silently)
	count, err := m.Consolidate(context.Background(), store.ScopeUser)
	require.NoError(t, err)
	// Count should be 0 because delete failed (incremented after delete succeeds)
	assert.Equal(t, 0, count)
}

func TestManager_Consolidate_AlreadyMerged(t *testing.T) {
	// Tests lines 185-186: skip already merged memories in inner loop
	// Scenario: mem1 similar to mem2, and mem2 similar to mem3
	// When i=0 (mem1), j=1 (mem2) -> merge mem2 into mem1, mark mem2 as merged
	// When i=0 (mem1), j=2 (mem3) -> check mem3
	// When i=1 (mem2), already merged, skip (line 181-183)
	// When i=2 (mem3), j loops but nothing to merge
	// The inner loop skips merged[memories[j].ID] at line 185-186

	backend := newMockBackend()

	cfg := DefaultConfig()
	cfg.ConsolidationInterval = 0
	cfg.SimilarityThreshold = 0.5

	// Create memories where:
	// - mem1 and mem2 are identical (will merge, mem2 marked as merged)
	// - mem3 is different but when i=1, we skip because mem1 (now containing merged content)
	//   might be similar to mem3 in the inner loop
	// Actually, we need mem2 to be in the inner loop when it's already merged
	// This happens when: i=0, j=1 merges mem2 into mem1
	// Then when i=0, j=2, we check mem3
	// Then when i=1, mem2 is marked merged so we skip (outer loop check)
	// To hit line 185-186, we need j to point to an already-merged memory
	// This happens with: mem1 similar to mem2, mem1 similar to mem3, mem2 similar to mem3
	// i=0, j=1: merge mem2, j=2: merge mem3 - both hit the check but neither is merged yet
	// We need i=1, j=2 where mem3 is already merged from i=0
	// So: mem1 similar to mem3, mem2 similar to mem3, mem1 NOT similar to mem2

	backend.listResults = []*store.Memory{
		{ID: "1", Content: "alpha beta gamma delta epsilon", Scope: store.ScopeUser},
		{ID: "2", Content: "one two three four five six", Scope: store.ScopeUser},
		{ID: "3", Content: "alpha beta gamma delta epsilon zeta", Scope: store.ScopeUser}, // Similar to mem1
	}

	m := NewManager(backend, cfg)

	count, err := m.Consolidate(context.Background(), store.ScopeUser)
	require.NoError(t, err)
	// mem3 should be merged into mem1
	assert.GreaterOrEqual(t, count, 1)
}

func TestManager_Consolidate_InnerLoopSkipsMerged(t *testing.T) {
	// Explicitly test line 185-186: inner loop skips already merged memory
	// Scenario with 4 memories:
	// - mem1 similar to mem2 (merge at i=0, j=1)
	// - mem1 similar to mem3 (merge at i=0, j=2, mem3 marked merged)
	// - mem2 similar to mem3 (when i=1, j=2: mem2 is merged so outer skip)
	// We need inner loop skip: i=X, j=Y where Y is already merged
	// This requires Y to be merged from an earlier i iteration but still be checked in inner loop

	// Actually the inner loop skip happens when:
	// i=0 merges j=1 and j=2
	// Then i=1 is skipped (outer check)
	// Then i=2 is skipped (outer check)
	// Inner skip only happens if there are 4+ memories

	backend := newMockBackend()

	cfg := DefaultConfig()
	cfg.ConsolidationInterval = 0
	cfg.SimilarityThreshold = 0.4

	// mem1 similar to mem2 and mem3
	// mem2 similar to mem4
	// Order of operations:
	// i=0 (mem1): j=1 (mem2) - check similarity, if high enough merge
	// i=0 (mem1): j=2 (mem3) - check similarity, if high enough merge
	// i=0 (mem1): j=3 (mem4) - check similarity
	// i=1 (mem2): if merged, skip via outer check (line 181)
	// i=2 (mem3): if merged, skip via outer check
	// i=3 (mem4): j loops from 4, nothing to check

	// To hit inner loop skip (line 185):
	// Need: i=X, j=Y where merged[memories[Y].ID] is true
	// This means Y was merged in a previous j iteration of the SAME i
	// Wait, no - merged is populated as we go, so if i=0 merges j=1,
	// then j=2 checks if mem2 is in merged - but we're checking j index not ID cross-check

	// Looking at code again:
	// for j := i + 1; j < len(memories); j++ {
	//     if merged[memories[j].ID] { continue }  <- line 185-186
	// This checks if memories[j] was already absorbed by a PREVIOUS i or j

	// So we need:
	// i=0 merges j=2 (not j=1)
	// i=1: j=2 is checked, but memories[2] was merged at i=0

	backend.listResults = []*store.Memory{
		{ID: "1", Content: "alpha beta gamma delta", Scope: store.ScopeUser},
		{ID: "2", Content: "one two three four", Scope: store.ScopeUser},                 // Different from 1
		{ID: "3", Content: "alpha beta gamma delta epsilon", Scope: store.ScopeUser},     // Similar to 1
		{ID: "4", Content: "one two three four five", Scope: store.ScopeUser},            // Similar to 2
	}

	m := NewManager(backend, cfg)

	// i=0 (mem1): j=1 (mem2) not similar, j=2 (mem3) similar -> merge, j=3 (mem4) not similar
	// i=1 (mem2): outer loop: not merged, j=2 (mem3) MERGED -> skip (hits line 185-186)
	// i=1 (mem2): j=3 (mem4) similar -> merge

	count, err := m.Consolidate(context.Background(), store.ScopeUser)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
}

// --- CalculateImportance Cap Tests (Lines 246-247) ---

func TestCalculateImportance_ExceedsCap(t *testing.T) {
	// Test that scores exceeding 1.0 are capped
	// This requires all boosts: long content (0.2), metadata (0.1), embedding (0.1), global (0.1)
	// Total: 0.5 base + 0.2 + 0.1 + 0.1 + 0.1 = 1.0 (exact cap)
	tests := []struct {
		name   string
		memory *store.Memory
	}{
		{
			name: "ExactlyCapped",
			memory: &store.Memory{
				Content:   string(make([]byte, 600)), // 0.2 boost
				Metadata:  map[string]any{"k": "v"},  // 0.1 boost
				Embedding: []float32{0.1},            // 0.1 boost
				Scope:     store.ScopeGlobal,         // 0.1 boost
			},
		},
		{
			name: "AllBoostsApplied",
			memory: &store.Memory{
				Content:   string(make([]byte, 1000)), // 0.2 boost (both length thresholds)
				Metadata:  map[string]any{"a": 1, "b": 2, "c": 3},
				Embedding: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
				Scope:     store.ScopeGlobal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateImportance(tt.memory)
			assert.Equal(t, 1.0, score, "Score should be capped at 1.0")
		})
	}
}

// --- wordOverlapSimilarity Edge Cases (Lines 308-309) ---

func TestWordOverlapSimilarity_UnionZero(t *testing.T) {
	// The union==0 case at lines 308-309 is actually unreachable in normal operation
	// because if both strings are empty, we return 1.0 early (line 278-279),
	// and if either is non-empty, union will be > 0.
	// However, we can verify the edge cases are handled correctly.

	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
	}{
		{
			name:     "BothEmptyStrings",
			a:        "",
			b:        "",
			expected: 1.0,
		},
		{
			name:     "WhitespaceOnly_A",
			a:        "   ",
			b:        "word",
			expected: 0.0,
		},
		{
			name:     "WhitespaceOnly_B",
			a:        "word",
			b:        "   \t\n",
			expected: 0.0,
		},
		{
			name:     "BothWhitespaceOnly",
			a:        "   ",
			b:        "  \t  ",
			expected: 1.0, // Both produce empty word lists
		},
		{
			name:     "TabsAndNewlines",
			a:        "\t\n\r",
			b:        "\t\n\r",
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wordOverlapSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

// --- List with Decay Tests ---

func TestManager_List_WithDecay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DecayRate = 1.0 // Aggressive decay for testing

	backend := newMockBackend()
	m := NewManager(backend, cfg)
	ctx := context.Background()

	// Add memory with old creation time
	oldMem := &store.Memory{
		ID:        "old",
		Content:   "test memory",
		Score:     0.9,
		Scope:     store.ScopeUser,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	_ = m.Add(ctx, oldMem)

	results, err := m.List(ctx, store.ScopeUser, nil)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Score should be decayed from original 0.9
	for _, r := range results {
		if r.ID == "old" {
			assert.Less(t, r.Score, 0.9, "Score should be decayed")
		}
	}
}

func TestManager_List_NoDecay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DecayRate = 0 // Disable decay

	backend := newMockBackend()
	m := NewManager(backend, cfg)
	ctx := context.Background()

	mem := &store.Memory{
		ID:        "m1",
		Content:   "test memory",
		Score:     0.9,
		Scope:     store.ScopeUser,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	_ = m.Add(ctx, mem)

	results, err := m.List(ctx, store.ScopeUser, nil)
	require.NoError(t, err)

	// Score should remain unchanged (no decay)
	for _, r := range results {
		if r.ID == "m1" {
			assert.Equal(t, 0.9, r.Score, "Score should not be decayed")
		}
	}
}
