---
title: Overview
description: What GDA is and why it exists
---

> ⚠️ Pre-alpha: single-user only. Designed for individual researchers who want to version, restore, and share datasets. Not for real-time collaborative editing.

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

### No merge conflicts (because there's nothing to merge)

GDA is single-user. You work alone on your data. No branches. No parallel edits. When you `checkout`, GDA replaces your entire project folder — there's nothing to merge.

This is intentional. Research data doesn't benefit from merge conflict resolution. You don't want to manually resolve a conflict in a 10 GB NIfTI file. You want to snapshot before a risky preprocessing step and checkout if it goes wrong.

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
- You need multiple people editing the same dataset simultaneously (planned)
- You need S3 or HTTP remotes (rsync only for now)
- You need Windows support (planned)

## Comparison

| | GDA | git-annex | DVC | Git LFS |
|---|---|---|---|---|
| No git needed | ✓ | ✗ | ✗ | ✗ |
| Single-user (conflict-free) | ✓ | ✗ | ✗ | ✗ |
| Multi-user | ✗ planned | ✓ limited | ✓ | ✓ |
| Deduplicates identical files | ✓ | ✓ | ✓ | ✓ |
| Symlink file listing | ✓ | ✓ | ✗ | ✓ |
| Remote types | rsync | S3, rsync, many | S3, GCS | S3, GH |

**Why no merge conflicts?** GDA is designed for one person working alone. You snapshot, you checkout. No branches, no parallel edits, no conflict resolution screen. If you want to combine two snapshots, that's not supported yet — it would require a merge feature (planned).

→ [Installation](/gda/installation/) — get GDA on your machine
→ [Quick Start](/gda/quick-start/) — track your first dataset in 5 minutes
