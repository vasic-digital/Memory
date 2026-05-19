# Memory Test Coverage Ledger (round-247)

Round-247 deep-doc enrichment under CONST-035 / Article XI §11.9 / CONST-050(B).

This document is the authoritative mapping of every exported symbol in `pkg/{store,mem0,entity,graph,memfd,memory}` to the test sources that exercise it. Drift between this file and `go test -cover` output is a CONST-035 bluff at the documentation-truth layer — fix the document OR add the missing test, never silently leave the gap.

## Verbatim 2026-05-19 operator mandate (CONST-049 §11.4.17)

> "all existing tests and Challenges do work in anti-bluff manner - they MUST confirm that all tested codebase really works as expected! We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completition and full usability by end users of the product!"

## Test-type matrix (CONST-050(B))

| Test type | Location | Status |
|-----------|----------|--------|
| Unit | `pkg/*/*_test.go` | PRESENT — every package |
| Integration | `tests/integration/` | PRESENT |
| End-to-end | `tests/e2e/` | PRESENT |
| Security | `tests/security/` | PRESENT |
| Stress | `tests/stress/` | PRESENT |
| Benchmark | `tests/benchmark/` | PRESENT |
| Challenges | `challenges/scripts/` | PRESENT — 12 scripts incl. paired-mutation `_describe_` |
| Bilingual fixtures | `tests/fixtures/i18n/` | PRESENT (round-247) |

## `pkg/store`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `Scope` | type alias | `pkg/store/store_test.go` (Scope constants) |
| `ScopeUser` / `ScopeSession` / `ScopeConversation` / `ScopeGlobal` | const | `pkg/store/store_test.go` — also exercised by round-247 runner across all four scopes |
| `Memory` | struct | `pkg/store/store_test.go` (TestInMemoryStore_Add) |
| `SearchOptions` | struct | `pkg/store/store_test.go` (TestInMemoryStore_Search) |
| `TimeRange` | struct | `pkg/store/store_test.go` (TestInMemoryStore_Search_TimeRange) |
| `ListOptions` | struct | `pkg/store/store_test.go` (TestInMemoryStore_List) |
| `DefaultSearchOptions()` | constructor | `pkg/store/store_test.go` (TestDefaultSearchOptions) |
| `DefaultListOptions()` | constructor | `pkg/store/store_test.go` (TestDefaultListOptions) |
| `MemoryStore` | interface | `pkg/store/store_test.go` (interface conformance via InMemoryStore) |
| `InMemoryStore` | struct | `pkg/store/store_test.go` (TestNewInMemoryStore) |
| `NewInMemoryStore()` | constructor | `pkg/store/store_test.go` (TestNewInMemoryStore) — also runner |
| `InMemoryStore.Add` | method | `pkg/store/store_test.go` (TestInMemoryStore_Add) — also runner |
| `InMemoryStore.Get` | method | `pkg/store/store_test.go` (TestInMemoryStore_Get) — also runner |
| `InMemoryStore.Update` | method | `pkg/store/store_test.go` (TestInMemoryStore_Update) |
| `InMemoryStore.Delete` | method | `pkg/store/store_test.go` (TestInMemoryStore_Delete) |
| `InMemoryStore.Search` | method | `pkg/store/store_test.go` (TestInMemoryStore_Search) — also runner |
| `InMemoryStore.List` | method | `pkg/store/store_test.go` (TestInMemoryStore_List) — also runner |

## `pkg/mem0`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `Config` | struct | `pkg/mem0/mem0_test.go` (TestDefaultConfig) |
| `DefaultConfig()` | constructor | `pkg/mem0/mem0_test.go` (TestDefaultConfig) |
| `Manager` | struct | `pkg/mem0/mem0_test.go` (TestNewManager) |
| `NewManager(backend, cfg)` | constructor | `pkg/mem0/mem0_test.go` (TestNewManager, TestNewManager_NilConfig) — also runner |
| `Manager.Add` | method | `pkg/mem0/mem0_test.go` (TestManager_Add) — also runner (bilingual scope + importance assignment) |
| `Manager.Search` | method | `pkg/mem0/mem0_test.go` (TestManager_Search_WithDecay) — also runner |
| `Manager.Get` | method | `pkg/mem0/mem0_test.go` (TestManager_Get) — also runner |
| `Manager.Update` | method | `pkg/mem0/mem0_test.go` (TestManager_Update) |
| `Manager.Delete` | method | `pkg/mem0/mem0_test.go` (TestManager_Delete) |
| `Manager.List` | method | `pkg/mem0/mem0_test.go` (TestManager_List) |
| `Manager.Consolidate` | method | `pkg/mem0/mem0_test.go` (TestManager_Consolidate, TestManager_Consolidate_Cooldown) |
| `CalculateImportance(memory)` | func | `pkg/mem0/mem0_test.go` (TestCalculateImportance) — runner asserts in (0,1] for every bilingual fixture |
| `ApplyDecay(score, t0, now, rate)` | func | `pkg/mem0/mem0_test.go` (TestApplyDecay, TestApplyDecay_Zero) |

## `pkg/entity`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `Entity` | struct | `pkg/entity/entity_test.go` (TestEntity_Construct) |
| `Relation` | struct | `pkg/entity/entity_test.go` (TestRelation_Construct) |
| `Extractor` | interface | `pkg/entity/entity_test.go` (PatternExtractor conformance) |
| `Pattern` | struct | `pkg/entity/entity_test.go` (TestPatternExtractor_WithEntityPattern) |
| `PatternExtractor` | struct | `pkg/entity/entity_test.go` (TestNewPatternExtractor) |
| `NewPatternExtractor()` | constructor | `pkg/entity/entity_test.go` (TestNewPatternExtractor) |
| `PatternExtractor.WithEntityPattern` | method | `pkg/entity/entity_test.go` (TestPatternExtractor_WithEntityPattern) |
| `PatternExtractor.WithRelationPattern` | method | `pkg/entity/entity_test.go` (TestPatternExtractor_WithRelationPattern) |
| `PatternExtractor.Extract` | method | `pkg/entity/entity_test.go` (TestPatternExtractor_Extract — email/url/noun_phrase + is_a/has/uses) |

## `pkg/graph`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `Node` | struct | `pkg/graph/graph_test.go` (TestInMemoryGraph_AddNode) |
| `Edge` | struct | `pkg/graph/graph_test.go` (TestInMemoryGraph_AddEdge) |
| `Graph` | interface | `pkg/graph/graph_test.go` (InMemoryGraph conformance) |
| `InMemoryGraph` | struct | `pkg/graph/graph_test.go` (TestNewInMemoryGraph) |
| `NewInMemoryGraph()` | constructor | `pkg/graph/graph_test.go` (TestNewInMemoryGraph) |
| `InMemoryGraph.AddNode` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_AddNode, TestInMemoryGraph_AddNode_EmptyID, TestInMemoryGraph_AddNode_Duplicate) |
| `InMemoryGraph.AddEdge` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_AddEdge, TestInMemoryGraph_AddEdge_MissingSource, TestInMemoryGraph_AddEdge_MissingTarget) |
| `InMemoryGraph.GetNode` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_GetNode) |
| `InMemoryGraph.GetNeighbors` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_GetNeighbors) |
| `InMemoryGraph.ShortestPath` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_ShortestPath, TestInMemoryGraph_ShortestPath_NoPath, TestInMemoryGraph_ShortestPath_Same) |
| `InMemoryGraph.Subgraph` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_Subgraph, TestInMemoryGraph_Subgraph_BoundedDepth) |
| `InMemoryGraph.Nodes` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_Nodes) |
| `InMemoryGraph.Edges` | method | `pkg/graph/graph_test.go` (TestInMemoryGraph_Edges) |

## `pkg/memfd`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `PSC` | struct | `pkg/memfd/memfd_test.go` (TestPSC_WriteRead) |
| `NewPSC(capacity)` | constructor | `pkg/memfd/memfd_test.go` (TestPSC_WriteRead) |
| `PSC.Write` | method | `pkg/memfd/memfd_test.go` (TestPSC_WriteRead) |
| `PSC.Read` | method | `pkg/memfd/memfd_test.go` (TestPSC_WriteRead) |
| `PSC.Close` | method | `pkg/memfd/memfd_test.go` (TestPSC_Close) |

## `pkg/memory`

| Symbol | Kind | Test source(s) |
|--------|------|---------------|
| `LeakDetector` | struct | `pkg/memory/leak_detector_test.go` (TestNewLeakDetector) |
| `LeakReport` | struct | `pkg/memory/leak_detector_test.go` (TestLeakDetector_GetReport) |
| `NewLeakDetector(interval, ratio)` | constructor | `pkg/memory/leak_detector_test.go` (TestNewLeakDetector) |
| `LeakDetector.Start` | method | `pkg/memory/leak_detector_test.go` (TestLeakDetector_Start, TestLeakDetector_DoubleStart) |
| `LeakDetector.Stop` | method | `pkg/memory/leak_detector_test.go` (TestLeakDetector_Stop) |
| `LeakDetector.GetReport` | method | `pkg/memory/leak_detector_test.go` (TestLeakDetector_GetReport) |
| `LeakDetector.GetSamples` | method | `pkg/memory/leak_detector_coverage_test.go` (TestLeakDetector_GetSamples) |
| `MemoryMonitor` | struct | `pkg/memory/leak_detector_test.go` (TestNewMemoryMonitor) |
| `NewMemoryMonitor(interval, ratio)` | constructor | `pkg/memory/leak_detector_test.go` (TestNewMemoryMonitor) |
| `MemoryMonitor.SetAlertCallback` | method | `pkg/memory/leak_detector_coverage_test.go` (TestMemoryMonitor_AlertCallback) |
| `MemoryMonitor.Start` | method | `pkg/memory/leak_detector_test.go` (TestMemoryMonitor_StartStop) |
| `MemoryMonitor.Stop` | method | `pkg/memory/leak_detector_test.go` (TestMemoryMonitor_StartStop) |
| `MemoryMonitor.Reports` | method | `pkg/memory/leak_detector_coverage_test.go` (TestMemoryMonitor_Reports) |
| `WriteHeapProfile(filename)` | func | `pkg/memory/leak_detector_edge_test.go` (TestWriteHeapProfile) |
| `WriteGoroutineProfile(filename)` | func | `pkg/memory/leak_detector_edge_test.go` (TestWriteGoroutineProfile) |
| `ForceGC()` | func | `pkg/memory/leak_detector_test.go` (TestForceGC) |
| `GetCurrentMemoryUsage()` | func | `pkg/memory/leak_detector_test.go` (TestGetCurrentMemoryUsage) |

## Edge cases (round-247)

- Empty store List — `pkg/store/store_test.go` (TestInMemoryStore_List_Empty)
- Memory-not-found Get/Update/Delete — `pkg/store/store_test.go` (corresponding `_NotFound` cases)
- Mem0 Add with zero score — runner asserts `CalculateImportance` reassigns to (0,1]
- Mem0 consolidation cooldown — `pkg/mem0/mem0_test.go` (TestManager_Consolidate_Cooldown)
- Mem0 decay with rate=0 — `pkg/mem0/mem0_test.go` (TestApplyDecay_Zero)
- Empty graph ShortestPath start=end — `pkg/graph/graph_test.go` (TestInMemoryGraph_ShortestPath_Same)
- Subgraph maxDepth=0 — `pkg/graph/graph_test.go` (TestInMemoryGraph_Subgraph_BoundedDepth)
- LeakDetector double Start — `pkg/memory/leak_detector_test.go` (TestLeakDetector_DoubleStart)
- LeakDetector concurrent Start/Stop (race) — `go test -race ./pkg/memory/...`
- UTF-8 / bilingual `Content` + `Metadata` round-trip — `tests/fixtures/i18n/payloads.json` exercised by `challenges/scripts/memory_describe_challenge.sh`

## Paired-mutation Challenge

`challenges/scripts/memory_describe_challenge.sh` accepts `--anti-bluff-mutate` to plant a deliberate ledger-vs-source mismatch (renames one tracked symbol in the ledger) and asserts the gate FAILS with exit 99. Without the flag the gate runs normal validation and MUST exit 0. Composition: CONST-035 (anti-bluff) × CONST-050(B) (paired mutation) × CONST-047 (cascade).

## Anti-bluff acceptance criteria

1. `go test -count=1 -race ./...` exits 0 — all packages PASS (verified round-247).
2. `bash challenges/scripts/memory_describe_challenge.sh` exits 0 (gate PASS on clean tree).
3. `bash challenges/scripts/memory_describe_challenge.sh --anti-bluff-mutate` exits 99 (gate correctly fails on planted mutation).
4. Every symbol in this ledger appears in the listed test source verbatim — no metadata-only / configuration-only ledger entries.
