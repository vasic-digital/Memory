// Package store provides core memory store interfaces and types for
// persistent memory management with scoping, search, and metadata support.
package store

import (
	"context"
	"time"
)

// Scope defines the visibility boundary of a memory.
type Scope string

const (
	// ScopeUser indicates memory visible to a specific user.
	ScopeUser Scope = "user"
	// ScopeSession indicates memory visible within a session.
	ScopeSession Scope = "session"
	// ScopeConversation indicates memory visible within a conversation.
	ScopeConversation Scope = "conversation"
	// ScopeGlobal indicates memory visible to all users.
	ScopeGlobal Scope = "global"
)

// Memory represents a stored memory unit with content, metadata,
// embedding vector, and scoping information.
type Memory struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]any         `json:"metadata,omitempty"`
	Scope     Scope                  `json:"scope"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Score     float64                `json:"score"`
	Embedding []float32              `json:"embedding,omitempty"`
}

// SearchOptions configures memory search behaviour.
type SearchOptions struct {
	TopK      int        `json:"top_k"`
	MinScore  float64    `json:"min_score"`
	Scope     Scope      `json:"scope,omitempty"`
	TimeRange *TimeRange `json:"time_range,omitempty"`
	Filter    map[string]any `json:"filter,omitempty"`
}

// TimeRange restricts search results to a time window.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ListOptions configures memory listing with pagination and ordering.
type ListOptions struct {
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
	Scope   Scope  `json:"scope,omitempty"`
}

// DefaultSearchOptions returns sensible default search options.
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		TopK:     10,
		MinScore: 0.0,
	}
}

// DefaultListOptions returns sensible default list options.
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Offset:  0,
		Limit:   100,
		OrderBy: "created_at",
	}
}

// MemoryStore defines the interface for memory storage operations.
type MemoryStore interface {
	// Add stores a new memory.
	Add(ctx context.Context, memory *Memory) error

	// Search returns memories matching the query and options.
	Search(ctx context.Context, query string, opts *SearchOptions) ([]*Memory, error)

	// Get retrieves a memory by ID.
	Get(ctx context.Context, id string) (*Memory, error)

	// Update modifies an existing memory.
	Update(ctx context.Context, memory *Memory) error

	// Delete removes a memory by ID.
	Delete(ctx context.Context, id string) error

	// List returns memories matching the scope and options.
	List(ctx context.Context, scope Scope, opts *ListOptions) ([]*Memory, error)
}
