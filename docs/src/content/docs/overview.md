---
title: Overview
description: What GDA is and why it exists
---

> ⚠️ Pre-alpha (v0.0.1). Designed for individual researchers who want to version, restore, and share datasets. Sharing via push/pull works today. Simultaneous editing is not supported.

## The Problem

You've got a 500 GB fMRI dataset. You need to:

- Track which files belong to which analysis
- Save and restore checkpoints as preprocessing changes
- Rename and restructure without copying 500 GB
- Share data with collaborators
- Verify data integrity months later

Git can't handle 500 GB files. Git-annex can, but it's a beast. DVC needs a git repo and pipeline definitions. Rsync gives you copies, not versions.

None of these tools were built for how researchers actually work.

## What GDA Does

GDA versions research data. You add files, it hashes them, stores them by fingerprint, and replaces originals with symlinks. That's it.

```shell
gda init                    # Start tracking
gda add bids/               # Add files (instant dedup)
gda snapshot preprocessing  # Save a checkpoint
gda checkout raw            # Restore previous state
gda push origin             # Sync to remote storage
```

## Why It's Different

### No merge conflicts (because there's no merging)

GDA has no branches. You work on one dataset state at a time. When you `checkout`, GDA replaces your entire project folder — there's nothing to merge.

This is intentional for researchers working alone: you don't want to resolve a conflict in a 10 GB NIfTI file. Snapshot before a risky step, checkout if it goes wrong.

If you need parallel dataset versions (e.g. different preprocessing streams), that's branches — planned but not yet implemented.

### No git required

GDA identifies files by a fingerprint (SHA256) computed from their content, not by their filename or path. Add the same file twice — stored once. Move or rename — instant, because only the symlink changes.

### Snapshots, not commits

A snapshot records every file's fingerprint at a point in time. It's a bookmark, not a diff. No commit graph, no history rewriting, no merge commits.

### Integrity by design

Stored files are read-only and named by their fingerprint. `gda fsck` verifies every file against its hash. If something corrupted, you'll know immediately.

## Is GDA Right For You?

**Good fit:**
- You download datasets and want to track preprocessing steps
- You share data with labmates and need everyone on the same file tree
- You want to archive data with integrity verification
- You restructure your files constantly and hate waiting for copies
- Git-annex made you want to throw your laptop off a roof

**Not a good fit (yet):**
- You need two people editing the same dataset at the same time (branches/merging planned)
- You need S3 or HTTP remotes (rsync only for now)
- You need Windows support (planned)

## Comparison

| | GDA | git-annex | DVC | Git LFS |
|---|---|---|---|---|
| No git needed | ✓ | ✗ | ✗ | ✗ |
| Conflict-free (no branches) | ✓ | ✗ | ✗ | ✗ |
| Concurrent editing | ✗ | ∼ limited | ✓ | ✓ |
| Deduplicates identical files | ✓ | ✓ | ✓ | ✓ |
| Symlink file listing | ✓ | ✓ | ✗ | ✓ |
| Remote sync method | rsync | S3, rsync, many | S3, GCS | S3, GH |

**Why no concurrent editing?** GDA has no branches and no merge mechanism. It's designed for one working tree at a time — perfect for a single researcher or for sharing data where only one person modifies at a time. If you need two people editing the same dataset simultaneously, this isn't ready for that yet. Branches and merging are planned.

→ [Installation](/gda/installation/) — get GDA on your machine
→ [Quick Start](/gda/quick-start/) — track your first dataset in 5 minutes
