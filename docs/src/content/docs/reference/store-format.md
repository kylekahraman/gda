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

- Objects are stored at `.gda/objects/XX/YYY...` where `XX` is the first 2 hex chars of the SHA256 hash and `YYY...` is the remaining 62 chars
- Objects are read-only (0444 permissions)
- Each object is a single file whose content matches its name (verified by `gda fsck`)

## Index

The index maps file paths to their content hashes, sizes, and modification times. Stored in BoltDB at `.gda/index/`.

## Snapshots

Snapshots are named manifests containing the full index state at creation time. Stored in `.gda/snapshots/`.
