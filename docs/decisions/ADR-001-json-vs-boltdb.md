# ADR-001: Index Storage — JSON vs BoltDB

**Date:** 2026-06-05  
**Status:** Accepted  
**Decision:** Start with JSON, swap to BoltDB when/if JSON becomes a bottleneck.

## Context

GDA needs a persistent index mapping working-tree paths to content-addressed object keys. The index is mutated on every `add`, `mv`, `rm`, and read on every `status`, `checkout`, `snapshot`.

## Options Considered

### JSON (current implementation)

- Single human-readable file `.gda/index.json`
- Zero dependencies
- Trivially debuggable — open in editor, grep, git diff
- Full rewrite on every mutation — O(n) write cost
- Must deserialize entire file to memory

### SQLite

- ACID, concurrent access, fast queries
- CGo dependency — cross-compilation pain, static linking issues
- Overkill for a single-user CLI tool
- Adds 10+ MB to binary size

### BoltDB

- Embedded Go KV store, no CGo
- ACID, concurrent readers, single writer
- O(1) lookups, O(log n) writes (page splits)
- Used by etcd, Hugo, and many Go projects
- Adds ~700 KB to binary
- Binary format — not human-readable without a tool

## Decision

**Start with JSON.** The JSON index meets all prototype requirements:

1. Single-user CLI — no concurrent access needed
2. Human-readable — critical for debugging during development
3. Zero dependencies — keeps the binary small and cross-compilable
4. Performance is fine up to ~10,000 files (JSON writes take <100ms)

**Abstract the index interface** behind a Go `Indexer` interface so BoltDB can be swapped in later without changing business logic.

## Consequences

- Index writes are O(n) in file count — acceptable for prototype scale
- Index must be kept in sync with the object store (no partial writes)
- Filesystem-level corruption of a single JSON file loses all index state (mitigated by snapshots, which are separate JSON files)
- Process crash mid-write can corrupt the JSON file (mitigated by atomic rename: write to `.tmp`, rename to `index.json`)

## When to Switch to BoltDB

When benchmarks show:
- `gda status` takes >500ms on the user's largest dataset
- `gda add` of a single file takes >100ms (index write is dominant)
- Index file size exceeds 50 MB

At that point, BoltDB is a drop-in replacement with no API changes.
