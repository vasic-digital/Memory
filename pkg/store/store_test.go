package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Scope constants ---

func TestScope_Constants(t *testing.T) {
	tests := []struct {
		name     string
		scope    Scope
		expected string
	}{
		{"User", ScopeUser, "user"},
		{"Session", ScopeSession, "session"},
		{"Conversation", ScopeConversation, "conversation"},
		{"Global", ScopeGlobal, "global"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, Scope(tt.expected), tt.scope)
		})
	}
}

// --- Memory struct ---

func TestMemory_Struct(t *testing.T) {
	now := time.Now()
	m := &Memory{
		ID:        "mem-1",
		Content:   "test content",
		Metadata:  map[string]any{"key": "value"},
		Scope:     ScopeUser,
		CreatedAt: now,
		UpdatedAt: now,
		Score:     0.95,
		Embedding: []float32{0.1, 0.2, 0.3},
	}

	assert.Equal(t, "mem-1", m.ID)
	assert.Equal(t, "test content", m.Content)
	assert.Equal(t, ScopeUser, m.Scope)
	assert.Equal(t, 0.95, m.Score)
	assert.Len(t, m.Embedding, 3)
	assert.Equal(t, "value", m.Metadata["key"])
}

// --- SearchOptions ---

func TestSearchOptions(t *testing.T) {
	now := time.Now()
	opts := &SearchOptions{
		TopK:     5,
		MinScore: 0.7,
		Scope:    ScopeSession,
		TimeRange: &TimeRange{
			Start: now.Add(-time.Hour),
			End:   now,
		},
		Filter: map[string]any{"category": "tech"},
	}

	assert.Equal(t, 5, opts.TopK)
	assert.Equal(t, 0.7, opts.MinScore)
	assert.Equal(t, ScopeSession, opts.Scope)
	assert.NotNil(t, opts.TimeRange)
	assert.Equal(t, "tech", opts.Filter["category"])
}

func TestDefaultSearchOptions(t *testing.T) {
	opts := DefaultSearchOptions()
	assert.Equal(t, 10, opts.TopK)
	assert.Equal(t, 0.0, opts.MinScore)
	assert.Equal(t, Scope(""), opts.Scope)
	assert.Nil(t, opts.TimeRange)
}

// --- ListOptions ---

func TestListOptions(t *testing.T) {
	opts := &ListOptions{
		Offset:  10,
		Limit:   50,
		OrderBy: "updated_at",
		Scope:   ScopeGlobal,
	}

	assert.Equal(t, 10, opts.Offset)
	assert.Equal(t, 50, opts.Limit)
	assert.Equal(t, "updated_at", opts.OrderBy)
	assert.Equal(t, ScopeGlobal, opts.Scope)
}

func TestDefaultListOptions(t *testing.T) {
	opts := DefaultListOptions()
	assert.Equal(t, 0, opts.Offset)
	assert.Equal(t, 100, opts.Limit)
	assert.Equal(t, "created_at", opts.OrderBy)
}

// --- InMemoryStore ---

func TestNewInMemoryStore(t *testing.T) {
	s := NewInMemoryStore()
	require.NotNil(t, s)
	assert.NotNil(t, s.memories)
}

func TestInMemoryStore_Add(t *testing.T) {
	tests := []struct {
		name   string
		memory *Memory
		checkID bool
	}{
		{
			name:    "WithID",
			memory:  &Memory{ID: "test-id", Content: "hello", Scope: ScopeUser},
			checkID: false,
		},
		{
			name:    "WithoutID",
			memory:  &Memory{Content: "auto id", Scope: ScopeSession},
			checkID: true,
		},
		{
			name: "WithTimestamps",
			memory: &Memory{
				ID:        "ts-id",
				Content:   "with time",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			checkID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStore()
			err := s.Add(context.Background(), tt.memory)
			require.NoError(t, err)

			if tt.checkID {
				assert.NotEmpty(t, tt.memory.ID)
			}
			assert.False(t, tt.memory.CreatedAt.IsZero())
			assert.False(t, tt.memory.UpdatedAt.IsZero())
		})
	}
}

func TestInMemoryStore_Get(t *testing.T) {
	tests := []struct {
		name      string
		setupID   string
		queryID   string
		expectErr bool
	}{
		{
			name:      "Found",
			setupID:   "mem-1",
			queryID:   "mem-1",
			expectErr: false,
		},
		{
			name:      "NotFound",
			setupID:   "mem-1",
			queryID:   "nonexistent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStore()
			ctx := context.Background()
			_ = s.Add(ctx, &Memory{ID: tt.setupID, Content: "test"})

			result, err := s.Get(ctx, tt.queryID)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "memory not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.setupID, result.ID)
			}
		})
	}
}

func TestInMemoryStore_Get_ReturnsCopy(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	_ = s.Add(ctx, &Memory{ID: "m1", Content: "original"})

	result, _ := s.Get(ctx, "m1")
	result.Content = "mutated"

	original, _ := s.Get(ctx, "m1")
	assert.Equal(t, "original", original.Content)
}

func TestInMemoryStore_Update(t *testing.T) {
	tests := []struct {
		name      string
		setupID   string
		updateID  string
		content   string
		expectErr bool
	}{
		{
			name:      "Success",
			setupID:   "mem-1",
			updateID:  "mem-1",
			content:   "updated content",
			expectErr: false,
		},
		{
			name:      "NotFound",
			setupID:   "mem-1",
			updateID:  "nonexistent",
			content:   "updated",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStore()
			ctx := context.Background()
			_ = s.Add(ctx, &Memory{ID: tt.setupID, Content: "original"})

			err := s.Update(ctx, &Memory{ID: tt.updateID, Content: tt.content})
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				result, _ := s.Get(ctx, tt.updateID)
				assert.Equal(t, tt.content, result.Content)
				assert.False(t, result.UpdatedAt.IsZero())
			}
		})
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupID   string
		deleteID  string
		expectErr bool
	}{
		{
			name:      "Success",
			setupID:   "mem-1",
			deleteID:  "mem-1",
			expectErr: false,
		},
		{
			name:      "NotFound",
			setupID:   "mem-1",
			deleteID:  "nonexistent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStore()
			ctx := context.Background()
			_ = s.Add(ctx, &Memory{ID: tt.setupID, Content: "test"})

			err := s.Delete(ctx, tt.deleteID)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				_, err = s.Get(ctx, tt.deleteID)
				require.Error(t, err)
			}
		})
	}
}

func TestInMemoryStore_Search(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	now := time.Now()
	memories := []*Memory{
		{
			ID: "1", Content: "Go programming language",
			Scope: ScopeUser, CreatedAt: now,
		},
		{
			ID: "2", Content: "Python machine learning",
			Scope: ScopeUser, CreatedAt: now.Add(-time.Hour),
		},
		{
			ID: "3", Content: "JavaScript web development",
			Scope: ScopeSession, CreatedAt: now,
		},
	}
	for _, m := range memories {
		_ = s.Add(ctx, m)
	}

	tests := []struct {
		name      string
		query     string
		opts      *SearchOptions
		minCount  int
		maxCount  int
	}{
		{
			name:     "BasicSearch",
			query:    "programming",
			opts:     nil,
			minCount: 1,
			maxCount: 3,
		},
		{
			name:     "SearchByScope",
			query:    "programming development",
			opts:     &SearchOptions{Scope: ScopeSession, MinScore: 0.0},
			minCount: 0,
			maxCount: 1,
		},
		{
			name:     "SearchWithTopK",
			query:    "programming",
			opts:     &SearchOptions{TopK: 1, MinScore: 0.0},
			minCount: 0,
			maxCount: 1,
		},
		{
			name: "SearchWithTimeRange",
			query: "programming machine",
			opts: &SearchOptions{
				TimeRange: &TimeRange{
					Start: now.Add(-30 * time.Minute),
					End:   now.Add(time.Minute),
				},
				MinScore: 0.0,
			},
			minCount: 0,
			maxCount: 2,
		},
		{
			name:     "SearchNoResults",
			query:    "nonexistent xyz abc",
			opts:     &SearchOptions{MinScore: 0.5},
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "SearchWithMinScore",
			query:    "Go programming",
			opts:     &SearchOptions{MinScore: 1.0},
			minCount: 0,
			maxCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := s.Search(ctx, tt.query, tt.opts)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(results), tt.minCount)
			assert.LessOrEqual(t, len(results), tt.maxCount)
		})
	}
}

func TestInMemoryStore_Search_ScoreOrdering(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	_ = s.Add(ctx, &Memory{ID: "1", Content: "Go programming language basics"})
	_ = s.Add(ctx, &Memory{ID: "2", Content: "Go"})

	results, err := s.Search(ctx, "Go programming language basics", nil)
	require.NoError(t, err)
	if len(results) >= 2 {
		assert.GreaterOrEqual(t, results[0].Score, results[1].Score)
	}
}

func TestInMemoryStore_List(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = s.Add(ctx, &Memory{
			ID:        fmt.Sprintf("mem-%d", i),
			Content:   fmt.Sprintf("Memory %d", i),
			Scope:     ScopeUser,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}
	_ = s.Add(ctx, &Memory{
		ID:      "mem-global",
		Content: "Global memory",
		Scope:   ScopeGlobal,
	})

	tests := []struct {
		name     string
		scope    Scope
		opts     *ListOptions
		expected int
	}{
		{
			name:     "AllUserScope",
			scope:    ScopeUser,
			opts:     nil,
			expected: 5,
		},
		{
			name:     "GlobalScope",
			scope:    ScopeGlobal,
			opts:     nil,
			expected: 1,
		},
		{
			name:     "WithLimit",
			scope:    ScopeUser,
			opts:     &ListOptions{Limit: 2},
			expected: 2,
		},
		{
			name:     "WithOffset",
			scope:    ScopeUser,
			opts:     &ListOptions{Offset: 3, Limit: 10},
			expected: 2,
		},
		{
			name:     "OffsetExceedsLength",
			scope:    ScopeUser,
			opts:     &ListOptions{Offset: 100},
			expected: 0,
		},
		{
			name:     "EmptyScope",
			scope:    "",
			opts:     nil,
			expected: 6,
		},
		{
			name:     "NonexistentScope",
			scope:    Scope("private"),
			opts:     nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := s.List(ctx, tt.scope, tt.opts)
			require.NoError(t, err)
			assert.Len(t, results, tt.expected)
		})
	}
}

func TestInMemoryStore_List_Ordering(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		_ = s.Add(ctx, &Memory{
			ID:        fmt.Sprintf("m%d", i),
			Content:   fmt.Sprintf("Memory %d", i),
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
			UpdatedAt: base.Add(time.Duration(2-i) * time.Hour),
		})
	}

	tests := []struct {
		name    string
		orderBy string
		firstID string
	}{
		{"ByCreatedAt", "created_at", "m0"},
		{"ByUpdatedAt", "updated_at", "m2"},
		{"DefaultOrder", "", "m0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := s.List(ctx, "", &ListOptions{
				OrderBy: tt.orderBy,
				Limit:   10,
			})
			require.NoError(t, err)
			require.NotEmpty(t, results)
			assert.Equal(t, tt.firstID, results[0].ID)
		})
	}
}

// --- calculateMatchScore ---

func TestCalculateMatchScore(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		content  string
		expected float64
	}{
		{"FullMatch", []string{"go", "programming"}, "Go programming language", 1.0},
		{"PartialMatch", []string{"go", "python"}, "Go programming", 0.5},
		{"NoMatch", []string{"python", "rust"}, "Go programming", 0.0},
		{"EmptyQuery", []string{}, "Go programming", 0.0},
		{"CaseInsensitive", []string{"go"}, "GO PROGRAMMING", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMatchScore(tt.words, tt.content)
			assert.InDelta(t, tt.expected, score, 0.001)
		})
	}
}

// --- sortMemories ---

func TestSortMemories(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	memories := []*Memory{
		{ID: "a", CreatedAt: base.Add(2 * time.Hour), UpdatedAt: base, Score: 0.3},
		{ID: "b", CreatedAt: base, UpdatedAt: base.Add(2 * time.Hour), Score: 0.9},
		{ID: "c", CreatedAt: base.Add(time.Hour), UpdatedAt: base.Add(time.Hour), Score: 0.6},
	}

	tests := []struct {
		name    string
		orderBy string
		firstID string
	}{
		{"ByCreatedAt", "created_at", "b"},
		{"ByUpdatedAt", "updated_at", "a"},
		{"ByScore", "score", "b"},
		{"Default", "unknown", "b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mems := make([]*Memory, len(memories))
			copy(mems, memories)
			sortMemories(mems, tt.orderBy)
			assert.Equal(t, tt.firstID, mems[0].ID)
		})
	}
}

// --- Concurrency ---

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			id := fmt.Sprintf("mem-%d", idx)
			_ = s.Add(ctx, &Memory{ID: id, Content: "content"})
			_, _ = s.Get(ctx, id)
			_, _ = s.Search(ctx, "content", nil)
			_, _ = s.List(ctx, "", nil)
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	results, err := s.List(ctx, "", nil)
	require.NoError(t, err)
	assert.Equal(t, 100, len(results))
}
