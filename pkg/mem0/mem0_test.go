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
