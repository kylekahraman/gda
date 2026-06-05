---
title: Store Format
description: How GDA stores data on disk
---

## Layout

```
.gda/
├── objects/
│   ├── a1/          ← first 2 characters of SHA256
│   │   └── b2c3d4...  ← remaining 62 characters
│   ├── ff/
│   │   └── ee...
│   └── ...
├── index/           ← BoltDB database (path → hash mappings)
└── snapshots/       ← Snapshot manifests
```

## Object Storage

- Objects are named by the SHA256 hash OF their content — the filename is a fingerprint, not the data itself
- Objects are read-only (0444 permissions)
- An object's content, when hashed, always produces the filename — this is how `gda fsck` verifies integrity

## Index

The index maps file paths to their content hashes, sizes, and modification times. Stored in BoltDB at `.gda/index/`.

## Snapshots

Snapshots are named manifests containing the full index state at creation time. Stored in `.gda/snapshots/`.
