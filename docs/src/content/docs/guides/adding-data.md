---
title: Adding Data
description: Track files and directories with GDA
---

## Adding Files

```shell
gda add bids/sub-01/meg/run-01.fif
gda add sub-02/beh/data.csv
```

## Adding Directories

GDA walks directories recursively:

```shell
gda add sub-01/
gda add bids/
```

Skips symlinks and directories automatically.

## Add All

```shell
gda add .
```

Adds every file in the repository. Excludes `.gda/` automatically.

## What Happens

When you add a file, GDA:
1. Computes its SHA256 hash while writing to the object store (single pass)
2. Stores the object in `.gda/objects/XX/YYY...` (read-only, 0444 permissions)
3. Replaces the original with a **relative symlink** to the object
4. Records the path→hash mapping in the index

## Deduplication

Identical content is stored once. If two files have the same hash, they share the same object in the store.
