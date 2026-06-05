---
title: Moving & Removing
description: Rename or untrack files in a GDA repository
---

## Moving Files

```shell
gda mv sub-01/meg/run.fif sub-01/meg/run-01.fif
```

This renames the symlink in the working tree and updates the index. The object in the store stays unchanged (the hash is the same).

## Moving Directories

```shell
gda mv bids/raw/ bids/sourcedata/
```

All files under `bids/raw/` are moved to `bids/sourcedata/` in one operation. Instant — no content copies.

## Safety Check

GDA verifies the source is a symlink pointing to a valid object before moving. If a file has been modified or unlocked, the move is rejected:

```
safety abort: sub-01/meg/run.fif is not a symlink.
It may be modified or unlocked. Run 'gda add' first.
```

## Removing Files

```shell
gda rm sub-01/meg/run.fif
gda rm sub-02/beh/sub-02_beh.tsv
```

Removes the symlink from the working tree and removes the entry from the index. The object **remains in the store** — other snapshots may still reference it.

## Garbage Collection

After removing files, run `gda gc` to clean up unreferenced objects:

```shell
gda gc --dry-run    # See what would be removed
gda gc              # Actually remove unreferenced objects
```

Objects younger than 1 hour are never deleted (grace period).
