# GDA Roadmap

## v0.0.x — Pre-alpha (current)

**Goal:** Core features working. One real user (me) hammering on it.

### v0.0.1 ✅ (released)
- init, add, status, mv, rm
- snapshot, log, checkout, gc
- fsck, unlock, lock
- remote push/pull (rsync)

### v0.0.2 ✅ (current)
- .gdaignore support + auto-create on init
- Hard .gda/ skip in add (fixes self-corruption)
- gda reindex (rebuild index from symlinks)
- gda undo (revert last lock)
- Progress counter in add

### v0.0.3 (next)
- Parallel hashing in add (2-3x faster on SSD)
- Status speed optimization (parallel stat)
- Progress bars for long ops (push, pull, add)
- gda unlock . (materialize all)

---

## v0.1.x — Alpha

**Goal:** Feature-complete core, real-user testing, documentation.

- Cross-platform: macOS + Linux (including ppc64le for HPC)
- Remote: S3 backend (optional)
- Documentation site (Astro Starlight)
- Basic CI for all platforms
- Windows support (developer mode symlinks)

---

## v0.2.x–0.9.x — Beta

**Goal:** Polish, hardening, edge cases.

- Batch operations (gda add --dry-run, gda rm --cascade)
- Parallel hashing for push (rsync from objects dir)
- Snapshot diff (gda diff snap1 snap2)
- Performance benchmarks vs git-annex / DataLad
- BIDS-aware commands (gda add --subject sub-*)
- Windows CI

---

## v1.0.0 — Stable

**Goal:** Production-ready for neuroscience research data.

- All edge cases documented and tested
- Real-world validation on 3+ datasets
- GUI? (optional, TUI?)
- DataLad special remote compatibility (optional)
