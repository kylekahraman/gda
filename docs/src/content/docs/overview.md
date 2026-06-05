---
title: Overview
description: What GDA is and why it exists
---

## The Problem

You've got a 500 GB fMRI dataset. You need to:

- Track which files belong to which analysis
- Save and restore checkpoints as preprocessing changes
- Rename and restructure without copying 500 GB
- Share data with collaborators
- Verify data integrity months later

Git can't handle 500 GB files. Git-annex technically can, but using it is a full-time job. DVC needs a git repo and pipeline definitions. Rsync gives you copies, not versions.

None of these tools were built for the way researchers actually work.

## What GDA Does

GDA is a content-addressed data versioning tool. It hashes files, stores them by their SHA256, and replaces originals with symlinks. That's it.

```shell
gda init                    # Start tracking
gda add bids/               # Add files (instant dedup)
gda snapshot preprocessing  # Save a checkpoint
gda checkout raw            # Restore previous state
gda push origin             # Sync to remote storage
```

## Why It's Different

### No git required

GDA identifies files by their content, not their location. Add the same file twice — zero extra storage. Move or rename — instant, because only the symlink changes.

### Snapshots, not commits

A snapshot is a complete manifest of your dataset at a point in time. No diffs, no merge conflicts. Checkout is always instant.

### Integrity by design

Objects are stored read-only with content-validating filenames. `gda fsck` verifies every file matches its hash. If something's corrupted, you'll know.

## Is GDA Right For You?

**Good fit:**
- You version large research datasets (MEG, fMRI, iEEG, behavioral)
- You want snapshots and checkout without managing a git repo
- You need remote sync (rsync to a server or cold storage)
- Git-annex makes you want to throw your computer out the window

**Not a good fit (yet):**
- You need concurrent multi-user editing (planned)
- You need S3 or HTTP remote backends (planned)
- You need Windows support (planned)

## Comparison

| Feature | GDA | git-annex | DVC | Git LFS |
|---------|-----|-----------|-----|---------|
| No git required | ✓ | ✗ | ✗ | ✗ |
| Content-addressed | ✓ | ✓ | ✓ | ✓ |
| Symlink working tree | ✓ | ✓ | ✗ | ✓ |
| Snapshots | ✓ | ~ (commits) | ~ (commits) | ~ (commits) |
| Merge conflicts | ✗ | ✓ | ✓ | ✓ |
| Remote backends | rsync | S3, rsync, many | S3, GCS | S3, GH |
| Setup complexity | 1 command | hours | moderate | moderate |

→ [Installation](/gda/installation/) — get GDA on your machine
→ [Quick Start](/gda/quick-start/) — track your first dataset in 5 minutes
