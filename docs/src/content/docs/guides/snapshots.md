---
title: Snapshots & Checkout
description: Version your dataset state with snapshots
---

Snapshots are named, immutable checkpoints of your dataset's file index. They let you return to any previous state.

## Creating a Snapshot

```shell
# Auto-named (timestamp-based)
gda snapshot

# Named
gda snapshot pre-processing-v2
```

A snapshot records every tracked file's path, hash, size, and modification time.

## Listing Snapshots

```shell
gda log
```

Output:
```
  pre-processing-v1   2026-06-01 14:32:05  143 files  2.3 GB
  pre-processing-v2   2026-06-03 09:15:22  158 files  2.4 GB
  raw-v1              2026-05-28 11:00:00  127 files  1.8 GB
```

## Restoring a Snapshot

```shell
gda checkout pre-processing-v2
```

Restores your project folder to match the snapshot:
- Creates symlinks for every tracked file
- Removes files tracked in the current index but absent in the snapshot
- Warns about unlocked modified files without overwriting them

## Safety

Checkout will **not** overwrite unlocked files that have been modified. You'll see:

```
warning: file is unlocked with modifications, skipping: sub-01/meg/run.fif
```

Lock or discard those files first before retrying.
