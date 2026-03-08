#!/usr/bin/env bash
# memory_functionality_challenge.sh - Validates Memory module core functionality and structure
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MODULE_NAME="Memory"

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

echo "=== ${MODULE_NAME} Functionality Challenge ==="
echo ""

# Test 1: Required packages exist
echo "Test: Required packages exist"
pkgs_ok=true
for pkg in entity graph mem0 store; do
    if [ ! -d "${MODULE_DIR}/pkg/${pkg}" ]; then
        fail "Missing package: pkg/${pkg}"
        pkgs_ok=false
    fi
done
if [ "$pkgs_ok" = true ]; then
    pass "All required packages present (entity, graph, mem0, store)"
fi

# Test 2: MemoryStore interface is defined
echo "Test: MemoryStore interface is defined"
if grep -rq "type MemoryStore interface" "${MODULE_DIR}/pkg/store/"; then
    pass "MemoryStore interface is defined in pkg/store"
else
    fail "MemoryStore interface not found in pkg/store"
fi

# Test 3: Memory struct is defined
echo "Test: Memory struct is defined"
if grep -rq "type Memory struct" "${MODULE_DIR}/pkg/store/"; then
    pass "Memory struct is defined in pkg/store"
else
    fail "Memory struct not found in pkg/store"
fi

# Test 4: Entity struct is defined
echo "Test: Entity struct is defined"
if grep -rq "type Entity struct" "${MODULE_DIR}/pkg/entity/"; then
    pass "Entity struct is defined in pkg/entity"
else
    fail "Entity struct not found in pkg/entity"
fi

# Test 5: Graph interface is defined
echo "Test: Graph interface is defined"
if grep -rq "type Graph interface" "${MODULE_DIR}/pkg/graph/"; then
    pass "Graph interface is defined in pkg/graph"
else
    fail "Graph interface not found in pkg/graph"
fi

# Test 6: Mem0-style memory support
echo "Test: Mem0-style memory support exists"
if grep -rq "type\s\+\w\+\s\+struct\|Manager\|Client" "${MODULE_DIR}/pkg/mem0/"; then
    pass "Mem0-style memory support found in pkg/mem0"
else
    fail "No Mem0-style memory support found"
fi

# Test 7: Memory scope support
echo "Test: Memory scope support exists"
if grep -rq "Scope\|scope" "${MODULE_DIR}/pkg/store/"; then
    pass "Memory scope support found"
else
    fail "No memory scope support found"
fi

# Test 8: Search capability
echo "Test: Search/semantic search capability exists"
if grep -rq "Search\|SearchOptions\|Query" "${MODULE_DIR}/pkg/store/"; then
    pass "Search capability found in pkg/store"
else
    fail "No search capability found"
fi

# Test 9: Entity extraction support
echo "Test: Entity extraction support exists"
if grep -rq "Extractor\|Extract\|extract" "${MODULE_DIR}/pkg/entity/"; then
    pass "Entity extraction support found"
else
    fail "No entity extraction support found"
fi

# Test 10: InMemoryStore implementation
echo "Test: InMemoryStore implementation exists"
if grep -rq "type InMemoryStore struct\|InMemory" "${MODULE_DIR}/pkg/store/"; then
    pass "InMemoryStore implementation found"
else
    fail "InMemoryStore implementation not found"
fi

echo ""
echo "=== Results: ${PASS}/${TOTAL} passed, ${FAIL} failed ==="
[ "${FAIL}" -eq 0 ] && exit 0 || exit 1
