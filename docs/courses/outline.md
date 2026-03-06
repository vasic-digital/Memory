# Course: Building Memory Systems in Go

An introduction to `digital.vasic.memory` covering memory storage, intelligent management, entity extraction, and knowledge graphs.

## Lessons

1. **Memory Store Fundamentals** -- The `MemoryStore` interface, `Memory` type, scoping, search, and the in-memory implementation.
2. **Intelligent Memory with Mem0** -- The `Manager` decorator pattern, importance scoring, exponential decay, and memory consolidation.
3. **Entity Extraction and Knowledge Graphs** -- Extracting entities and relations from text with `PatternExtractor`, building and querying an `InMemoryGraph` with BFS.
4. **Runtime Memory Monitoring** -- Using `LeakDetector` and `MemoryMonitor` to track heap growth, goroutine leaks, and write profiling snapshots.
