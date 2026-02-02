// Package entity provides entity extraction from text using pattern-based
// matching, producing entities and relations suitable for knowledge graph
// construction.
package entity

import (
	"regexp"
	"strings"
)

// Entity represents an extracted named entity.
type Entity struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Relation represents a directed relationship between two entities.
type Relation struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// Extractor defines the interface for entity and relation extraction.
type Extractor interface {
	// Extract extracts entities and relations from text.
	Extract(text string) ([]Entity, []Relation, error)
}

// Pattern defines a named regex pattern for entity extraction.
type Pattern struct {
	Name       string
	Type       string
	Expression *regexp.Regexp
}

// PatternExtractor extracts entities and relations using regex patterns.
type PatternExtractor struct {
	entityPatterns   []Pattern
	relationPatterns []relationPattern
}

// relationPattern matches subject-predicate-object triples.
type relationPattern struct {
	Name       string
	Predicate  string
	Expression *regexp.Regexp
}

// NewPatternExtractor creates a new PatternExtractor with default patterns.
func NewPatternExtractor() *PatternExtractor {
	return &PatternExtractor{
		entityPatterns:   defaultEntityPatterns(),
		relationPatterns: defaultRelationPatterns(),
	}
}

// WithEntityPattern adds a custom entity extraction pattern.
func (pe *PatternExtractor) WithEntityPattern(
	name, entityType, pattern string,
) *PatternExtractor {
	pe.entityPatterns = append(pe.entityPatterns, Pattern{
		Name:       name,
		Type:       entityType,
		Expression: regexp.MustCompile(pattern),
	})
	return pe
}

// WithRelationPattern adds a custom relation extraction pattern.
func (pe *PatternExtractor) WithRelationPattern(
	name, predicate, pattern string,
) *PatternExtractor {
	pe.relationPatterns = append(pe.relationPatterns, relationPattern{
		Name:       name,
		Predicate:  predicate,
		Expression: regexp.MustCompile(pattern),
	})
	return pe
}

// Extract extracts entities and relations from the given text.
func (pe *PatternExtractor) Extract(text string) ([]Entity, []Relation, error) {
	entities := pe.extractEntities(text)
	relations := pe.extractRelations(text)
	return entities, relations, nil
}

// extractEntities extracts entities using configured patterns.
func (pe *PatternExtractor) extractEntities(text string) []Entity {
	seen := make(map[string]bool)
	var entities []Entity

	for _, p := range pe.entityPatterns {
		matches := p.Expression.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			name := strings.TrimSpace(match[1])
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			entities = append(entities, Entity{
				Name: name,
				Type: p.Type,
			})
		}
	}

	return entities
}

// extractRelations extracts relations using configured patterns.
func (pe *PatternExtractor) extractRelations(text string) []Relation {
	var relations []Relation

	for _, rp := range pe.relationPatterns {
		matches := rp.Expression.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			subject := strings.TrimSpace(match[1])
			object := strings.TrimSpace(match[2])
			if subject == "" || object == "" {
				continue
			}
			relations = append(relations, Relation{
				Subject:   subject,
				Predicate: rp.Predicate,
				Object:    object,
			})
		}
	}

	return relations
}

// defaultEntityPatterns returns built-in entity extraction patterns.
func defaultEntityPatterns() []Pattern {
	return []Pattern{
		{
			Name:       "email",
			Type:       "email",
			Expression: regexp.MustCompile(`([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`),
		},
		{
			Name:       "url",
			Type:       "url",
			Expression: regexp.MustCompile(`(https?://[^\s]+)`),
		},
		{
			Name:       "capitalized_phrase",
			Type:       "noun_phrase",
			Expression: regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)+)\b`),
		},
	}
}

// defaultRelationPatterns returns built-in relation extraction patterns.
func defaultRelationPatterns() []relationPattern {
	return []relationPattern{
		{
			Name:      "is_a",
			Predicate: "is_a",
			Expression: regexp.MustCompile(
				`(?i)(\w+(?:\s+\w+)?)\s+is\s+(?:a|an)\s+(\w+(?:\s+\w+)?)`,
			),
		},
		{
			Name:      "has",
			Predicate: "has",
			Expression: regexp.MustCompile(
				`(?i)(\w+(?:\s+\w+)?)\s+has\s+(\w+(?:\s+\w+)?)`,
			),
		},
		{
			Name:      "uses",
			Predicate: "uses",
			Expression: regexp.MustCompile(
				`(?i)(\w+(?:\s+\w+)?)\s+uses\s+(\w+(?:\s+\w+)?)`,
			),
		},
	}
}
