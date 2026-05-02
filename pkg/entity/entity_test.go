package entity

import (
	"regexp"
	"strings"
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

// --- Additional coverage tests ---

func TestPatternExtractor_Extract_EmptyMatches(t *testing.T) {
	pe := NewPatternExtractor()
	// Text that matches pattern but with empty capture group
	pe.WithEntityPattern("empty", "empty", `()`)

	entities, _, err := pe.Extract("some text")
	require.NoError(t, err)
	// Should not include empty entities
	for _, e := range entities {
		assert.NotEmpty(t, e.Name)
	}
}

func TestPatternExtractor_Extract_WhitespaceOnlyMatch(t *testing.T) {
	pe := NewPatternExtractor()
	// Add pattern that might match whitespace
	pe.WithEntityPattern("space", "space", `(\s+)`)

	entities, _, err := pe.Extract("hello   world")
	require.NoError(t, err)
	// Should not include whitespace-only entities
	for _, e := range entities {
		assert.NotEmpty(t, strings.TrimSpace(e.Name))
	}
}

func TestPatternExtractor_Extract_RelationEmptySubject(t *testing.T) {
	pe := &PatternExtractor{
		entityPatterns:   defaultEntityPatterns(),
		relationPatterns: []relationPattern{},
	}
	// Add pattern that might match with empty subject
	pe.WithRelationPattern("test", "test_pred", `()\s+is\s+(\w+)`)

	_, relations, err := pe.Extract(" is thing")
	require.NoError(t, err)
	// Should not include relations with empty subject
	for _, r := range relations {
		assert.NotEmpty(t, r.Subject)
	}
}

func TestPatternExtractor_Extract_RelationEmptyObject(t *testing.T) {
	pe := &PatternExtractor{
		entityPatterns:   defaultEntityPatterns(),
		relationPatterns: []relationPattern{},
	}
	// Add pattern that might match with empty object
	pe.WithRelationPattern("test", "test_pred", `(\w+)\s+is\s+()`)

	_, relations, err := pe.Extract("thing is ")
	require.NoError(t, err)
	// Should not include relations with empty object
	for _, r := range relations {
		assert.NotEmpty(t, r.Object)
	}
}

func TestPatternExtractor_Extract_SingleCaptureGroup(t *testing.T) {
	pe := &PatternExtractor{
		entityPatterns:   defaultEntityPatterns(),
		relationPatterns: []relationPattern{},
	}
	// Add relation pattern with only one capture group (invalid for relations)
	pe.relationPatterns = append(pe.relationPatterns, relationPattern{
		Name:       "single",
		Predicate:  "single",
		Expression: regexp.MustCompile(`(\w+)`),
	})

	_, relations, err := pe.Extract("word")
	require.NoError(t, err)
	// Should not include relations from single capture group
	hasRelation := false
	for _, r := range relations {
		if r.Predicate == "single" {
			hasRelation = true
		}
	}
	assert.False(t, hasRelation)
}

func TestPatternExtractor_Extract_NoCaptureGroups(t *testing.T) {
	pe := &PatternExtractor{
		entityPatterns:   []Pattern{},
		relationPatterns: []relationPattern{},
	}
	// Add pattern with no capture groups
	pe.entityPatterns = append(pe.entityPatterns, Pattern{
		Name:       "nocap",
		Type:       "nocap",
		Expression: regexp.MustCompile(`test`),
	})

	entities, _, err := pe.Extract("this is a test")
	require.NoError(t, err)
	// Should not include entities from no capture group pattern
	hasNocap := false
	for _, e := range entities {
		if e.Type == "nocap" {
			hasNocap = true
		}
	}
	assert.False(t, hasNocap)
}

func TestPattern_Struct(t *testing.T) {
	p := Pattern{
		Name:       "test",
		Type:       "test_type",
		Expression: regexp.MustCompile(`(\w+)`),
	}
	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "test_type", p.Type)
	assert.NotNil(t, p.Expression)
}

func TestRelationPattern_Struct(t *testing.T) {
	rp := relationPattern{
		Name:       "test",
		Predicate:  "test_pred",
		Expression: regexp.MustCompile(`(\w+)\s+to\s+(\w+)`),
	}
	assert.Equal(t, "test", rp.Name)
	assert.Equal(t, "test_pred", rp.Predicate)
	assert.NotNil(t, rp.Expression)
}

func TestDefaultEntityPatterns(t *testing.T) {
	patterns := defaultEntityPatterns()
	assert.Len(t, patterns, 3)

	names := make([]string, len(patterns))
	for i, p := range patterns {
		names[i] = p.Name
	}
	assert.Contains(t, names, "email")
	assert.Contains(t, names, "url")
	assert.Contains(t, names, "capitalized_phrase")
}

func TestDefaultRelationPatterns(t *testing.T) {
	patterns := defaultRelationPatterns()
	assert.Len(t, patterns, 3)

	predicates := make([]string, len(patterns))
	for i, p := range patterns {
		predicates[i] = p.Predicate
	}
	assert.Contains(t, predicates, "is_a")
	assert.Contains(t, predicates, "has")
	assert.Contains(t, predicates, "uses")
}

func TestPatternExtractor_ChainedWithMethods(t *testing.T) {
	pe := NewPatternExtractor().
		WithEntityPattern("custom1", "type1", `(CUSTOM1)`).
		WithEntityPattern("custom2", "type2", `(CUSTOM2)`).
		WithRelationPattern("rel1", "pred1", `(\w+)\s+does\s+(\w+)`).
		WithRelationPattern("rel2", "pred2", `(\w+)\s+makes\s+(\w+)`)

	// Verify all patterns were added
	hasCustom1 := false
	hasCustom2 := false
	for _, p := range pe.entityPatterns {
		if p.Name == "custom1" {
			hasCustom1 = true
		}
		if p.Name == "custom2" {
			hasCustom2 = true
		}
	}
	assert.True(t, hasCustom1)
	assert.True(t, hasCustom2)

	hasRel1 := false
	hasRel2 := false
	for _, r := range pe.relationPatterns {
		if r.Name == "rel1" {
			hasRel1 = true
		}
		if r.Name == "rel2" {
			hasRel2 = true
		}
	}
	assert.True(t, hasRel1)
	assert.True(t, hasRel2)
}

func TestPatternExtractor_Extract_ComplexText(t *testing.T) {
	pe := NewPatternExtractor()

	text := `
Contact John Smith at john@example.com or visit https://example.com.
The System has multiple Components and uses Docker containers.
Go is a programming language that uses goroutines.
`

	entities, relations, err := pe.Extract(text)
	require.NoError(t, err)

	// Verify entities
	entityNames := make([]string, len(entities))
	for i, e := range entities {
		entityNames[i] = e.Name
	}
	assert.Contains(t, entityNames, "john@example.com")
	assert.Contains(t, entityNames, "https://example.com.")

	// Verify relations
	predicates := make([]string, len(relations))
	for i, r := range relations {
		predicates[i] = r.Predicate
	}
	assert.Contains(t, predicates, "has")
	assert.Contains(t, predicates, "uses")
	assert.Contains(t, predicates, "is_a")
}
