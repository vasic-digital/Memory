# Changelog

All notable changes to the `digital.vasic.memory` module will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-03

### Added

- **pkg/store**: Core memory store package.
  - `MemoryStore` interface with Add, Search, Get, Update, Delete, List operations.
  - `Memory` struct with ID, Content, Metadata, Scope, CreatedAt, UpdatedAt, Score, Embedding fields.
  - `Scope` type with four constants: `ScopeUser`, `ScopeSession`, `ScopeConversation`, `ScopeGlobal`.
  - `SearchOptions` with TopK, MinScore, Scope, TimeRange, and Filter fields.
  - `ListOptions` with Offset, Limit, OrderBy, and Scope fields.
  - `InMemoryStore`: thread-safe in-memory implementation using `sync.RWMutex`.
  - Word-overlap search scoring (case-insensitive Jaccard-style).
  - Copy semantics on Add and Get to prevent external mutation.
  - `DefaultSearchOptions()` and `DefaultListOptions()` factory functions.
  - Full test suite with table-driven tests and concurrency safety tests.

- **pkg/mem0**: Mem0-style memory management package.
  - `Manager` type wrapping any `MemoryStore` backend (implements `MemoryStore`).
  - `Config` with DefaultScope, MaxMemories, ConsolidationInterval, DecayRate, SimilarityThreshold.
  - `DefaultConfig()` factory with sensible defaults.
  - `CalculateImportance()` -- importance scoring based on content length, metadata, embedding, and scope.
  - `ApplyDecay()` -- exponential time-based decay: `score * exp(-rate * hours)`.
  - `Consolidate()` -- merges similar memories using Jaccard word-overlap similarity with cooldown.
  - Automatic scope assignment and ID generation on Add.
  - Decay applied at read time (Search, List), not write time.
  - Full test suite including decay, consolidation, and interface compliance tests.

- **pkg/entity**: Entity and relation extraction package.
  - `Extractor` interface with `Extract(text) ([]Entity, []Relation, error)`.
  - `Entity` struct with Name, Type, Attributes.
  - `Relation` struct with Subject, Predicate, Object (SPO triples).
  - `PatternExtractor` with builder pattern for custom patterns.
  - Default entity patterns: email, URL, capitalized noun phrases.
  - Default relation patterns: is_a, has, uses.
  - `WithEntityPattern()` and `WithRelationPattern()` builder methods.
  - Entity deduplication by name.
  - Full test suite with custom pattern tests and interface compliance.

- **pkg/graph**: Knowledge graph package.
  - `Graph` interface with AddNode, AddEdge, GetNode, GetNeighbors, ShortestPath, Subgraph, Nodes, Edges.
  - `Node` struct with ID, Type, Properties.
  - `Edge` struct with Source, Target, Relation, Weight.
  - `InMemoryGraph`: thread-safe in-memory implementation using adjacency lists and `sync.RWMutex`.
  - BFS shortest path by hop count.
  - Subgraph extraction within configurable depth.
  - Neighbor deduplication for multiple edges to same target.
  - Full test suite with path finding, subgraph, concurrency, and interface compliance tests.

- **Documentation**: CLAUDE.md, README.md, AGENTS.md, User Guide, Architecture, API Reference, Contributing Guide, Changelog, Mermaid diagrams.
