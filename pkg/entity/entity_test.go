package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Entity struct ---

func TestEntity_Struct(t *testing.T) {
	e := Entity{
		Name:       "Go",
		Type:       "language",
		Attributes: map[string]any{"version": "1.24"},
	}
	assert.Equal(t, "Go", e.Name)
	assert.Equal(t, "language", e.Type)
	assert.Equal(t, "1.24", e.Attributes["version"])
}

// --- Relation struct ---

func TestRelation_Struct(t *testing.T) {
	r := Relation{
		Subject:   "Alice",
		Predicate: "knows",
		Object:    "Bob",
	}
	assert.Equal(t, "Alice", r.Subject)
	assert.Equal(t, "knows", r.Predicate)
	assert.Equal(t, "Bob", r.Object)
}

// --- NewPatternExtractor ---

func TestNewPatternExtractor(t *testing.T) {
	pe := NewPatternExtractor()
	require.NotNil(t, pe)
	assert.NotEmpty(t, pe.entityPatterns)
	assert.NotEmpty(t, pe.relationPatterns)
}

// --- WithEntityPattern ---

func TestPatternExtractor_WithEntityPattern(t *testing.T) {
	pe := NewPatternExtractor()
	initialCount := len(pe.entityPatterns)

	result := pe.WithEntityPattern("test", "custom", `\b(TEST\d+)\b`)
	assert.Same(t, pe, result)
	assert.Equal(t, initialCount+1, len(pe.entityPatterns))
}

// --- WithRelationPattern ---

func TestPatternExtractor_WithRelationPattern(t *testing.T) {
	pe := NewPatternExtractor()
	initialCount := len(pe.relationPatterns)

	result := pe.WithRelationPattern("test", "relates_to", `(\w+)\s+relates\s+to\s+(\w+)`)
	assert.Same(t, pe, result)
	assert.Equal(t, initialCount+1, len(pe.relationPatterns))
}

// --- Extract entities ---

func TestPatternExtractor_Extract_Entities(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		minEntities    int
		expectedTypes  []string
		expectedNames  []string
	}{
		{
			name:          "ExtractEmail",
			text:          "Contact us at test@example.com for info",
			minEntities:   1,
			expectedTypes: []string{"email"},
			expectedNames: []string{"test@example.com"},
		},
		{
			name:          "ExtractURL",
			text:          "Visit https://example.com for more",
			minEntities:   1,
			expectedTypes: []string{"url"},
			expectedNames: []string{"https://example.com"},
		},
		{
			name:          "ExtractCapitalizedPhrase",
			text:          "John Smith works at Acme Corp",
			minEntities:   1,
			expectedTypes: []string{"noun_phrase"},
		},
		{
			name:        "NoEntities",
			text:        "this is all lowercase text",
			minEntities: 0,
		},
		{
			name:        "EmptyText",
			text:        "",
			minEntities: 0,
		},
		{
			name:          "MultipleEmails",
			text:          "Send to alice@test.com and bob@test.com",
			minEntities:   2,
			expectedTypes: []string{"email", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewPatternExtractor()
			entities, _, err := pe.Extract(tt.text)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(entities), tt.minEntities)

			if len(tt.expectedNames) > 0 {
				names := make([]string, len(entities))
				for i, e := range entities {
					names[i] = e.Name
				}
				for _, expected := range tt.expectedNames {
					assert.Contains(t, names, expected)
				}
			}
		})
	}
}

func TestPatternExtractor_Extract_NoDuplicateEntities(t *testing.T) {
	pe := NewPatternExtractor()
	text := "Email test@test.com and also test@test.com"
	entities, _, err := pe.Extract(text)
	require.NoError(t, err)

	nameCount := make(map[string]int)
	for _, e := range entities {
		nameCount[e.Name]++
	}
	for _, count := range nameCount {
		assert.Equal(t, 1, count)
	}
}

// --- Extract relations ---

func TestPatternExtractor_Extract_Relations(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		minRelations  int
		expectedPreds []string
	}{
		{
			name:          "IsARelation",
			text:          "Go is a programming language",
			minRelations:  1,
			expectedPreds: []string{"is_a"},
		},
		{
			name:          "HasRelation",
			text:          "The system has multiple components",
			minRelations:  1,
			expectedPreds: []string{"has"},
		},
		{
			name:          "UsesRelation",
			text:          "The project uses Docker containers",
			minRelations:  1,
			expectedPreds: []string{"uses"},
		},
		{
			name:         "NoRelations",
			text:         "just some plain text here",
			minRelations: 0,
		},
		{
			name:          "MultipleRelations",
			text:          "Go is a language and it uses goroutines",
			minRelations:  2,
			expectedPreds: []string{"is_a", "uses"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewPatternExtractor()
			_, relations, err := pe.Extract(tt.text)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(relations), tt.minRelations)

			if len(tt.expectedPreds) > 0 {
				preds := make([]string, len(relations))
				for i, r := range relations {
					preds[i] = r.Predicate
				}
				for _, expected := range tt.expectedPreds {
					assert.Contains(t, preds, expected)
				}
			}
		})
	}
}

// --- Custom patterns ---

func TestPatternExtractor_CustomPatterns(t *testing.T) {
	pe := NewPatternExtractor()
	pe.WithEntityPattern("version", "version", `v(\d+\.\d+\.\d+)`)
	pe.WithRelationPattern("depends", "depends_on", `(\w+)\s+depends\s+on\s+(\w+)`)

	text := "Module v1.2.3 depends on core"
	entities, relations, err := pe.Extract(text)
	require.NoError(t, err)

	// Check version entity
	hasVersion := false
	for _, e := range entities {
		if e.Type == "version" {
			hasVersion = true
		}
	}
	assert.True(t, hasVersion)

	// Check depends_on relation
	hasDepends := false
	for _, r := range relations {
		if r.Predicate == "depends_on" {
			hasDepends = true
		}
	}
	assert.True(t, hasDepends)
}

// --- Interface compliance ---

func TestPatternExtractor_ImplementsExtractor(t *testing.T) {
	var _ Extractor = (*PatternExtractor)(nil)
}
