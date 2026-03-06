# Course: Memory Store, Entity Extraction, and Knowledge Graphs in Go

## Module Overview

This course covers the `digital.vasic.memory` module, which provides an in-memory store with search and decay, a Mem0-style manager layer with importance scoring and consolidation, entity and relation extraction via regex patterns, an in-memory knowledge graph with BFS traversal, and memory leak detection utilities. All implementations are thread-safe with zero external runtime dependencies beyond `uuid`.

## Prerequisites

- Intermediate Go knowledge (interfaces, sync primitives)
- Understanding of graph data structures and BFS
- Basic familiarity with text processing and regex
- Go 1.24+ installed

## Lessons

| # | Title | Duration |
|---|-------|----------|
| 1 | Memory Store Interface and In-Memory Implementation | 45 min |
| 2 | Mem0 Manager -- Decay, Importance, and Consolidation | 45 min |
| 3 | Entity Extraction and Knowledge Graphs | 45 min |
| 4 | Runtime Memory Monitoring and Leak Detection | 35 min |

## Source Files

- `pkg/store/` -- `MemoryStore` interface, `Memory` struct, `InMemoryStore`
- `pkg/mem0/` -- Manager with decay, importance scoring, and consolidation
- `pkg/entity/` -- `Extractor` interface, `PatternExtractor` with default patterns
- `pkg/graph/` -- `Graph` interface, `InMemoryGraph` with BFS shortest path
- `pkg/memory/` -- Memory leak detection and monitoring utilities
