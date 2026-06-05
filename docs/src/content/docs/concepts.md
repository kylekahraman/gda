---
title: How GDA Works
description: Architecture and core concepts
---

Most tools identify files by where they live (file path). GDA identifies files by a cryptographic fingerprint (SHA256 hash) computed FROM their content. Two files with identical content have identical fingerprints — GDA stores them once. Change one byte, and the fingerprint changes — GDA treats it as a different file.

## Core Concepts

### How files are stored

When you add a file, GDA:
1. Reads the file and computes a SHA256 fingerprint (hash)
2. Writes the content to `.gda/objects/XX/YYYYYY...` where `XXYYYYYY...` is the full fingerprint
3. Makes the file read-only (0444) — stored files never change

Same content = same fingerprint = stored once.

```
Original files:               Object store:
  sub-01/data.nii.gz          .gda/objects/a1/b2c3d4e5...  ← 0444, read-only
  sub-02/data.nii.gz  ──→    (same object, because same hash → deduplicated)
```

### Symlink-based file listing

After adding, the original file is replaced with a **relative symlink** to the object. This means:
- You can browse and open files normally (no special tools needed)
- Renaming or restructuring only changes the symlink (instant, no data copy)
- Your project folder always shows real files, not hashes

```shell
$ ls -l sub-01/meg/
lrwxrwxrwx  ...  run.fif -> ../../../.gda/objects/a1/b2c3d4...
```

### Index

The index maps every tracked path to its content hash, size, and modification time. Stored in `.gda/index/` (BoltDB database).

### Snapshots

A snapshot is a named, immutable copy of the index. It records every file's path and hash at a point in time. Snapshots are NOT copies of the data — they're tiny manifests (hashes only).

```
Snapshot "raw-v1"
├── sub-01/meg/run.fif     → a1b2c3d4e5f6...
├── sub-01/beh/data.csv    → f6e5d4c3b2a1...
└── sub-02/meg/run.fif     → 9a8b7c6d5e4f...
```

Restoring a snapshot recreates your project folder by placing symlinks for every entry.

## Data Flow

```
[your files] → hash (SHA256) + copy → [.gda/objects/]
                                   ↘ [index (path → hash)]
                                              ↕
                                     [snapshots (named index copies)]
```

**Adding a file:** Read → hash → write to store → symlink → index entry
**Creating a snapshot:** Copy current index to snapshot manifest
**Checkout:** Read snapshot → create symlinks → restore index
**GC:** Find unreferenced objects → delete

## Design Decisions

### Why not git?

Git is designed for tracking changes to text files. Its model — trees of commits with diffs — breaks down with large binary files. Git-annex works around this by storing content outside git, but introduces substantial complexity. GDA starts from scratch with a model designed for research data.

### Why SHA256?

SHA256 is the standard for content addressing. It's fast (hardware-accelerated on modern CPUs), collision-resistant, and widely supported.

### Why BoltDB?

The index needs to handle millions of entries with fast lookups. BoltDB is an embedded key-value store written in Go — no external dependencies, ACID transactions, good performance.

### Why rsync for remotes?

Rsync is available on every Linux and macOS system. It's been around forever and works for large data transfers. SSH-based rsync doesn't need special server software. S3 and other backends can be added later.
