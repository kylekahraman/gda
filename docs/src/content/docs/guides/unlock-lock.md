---
title: Unlock & Lock
description: Edit tracked files without breaking the symlink system
---

By default, GDA tracks files as **read-only symlinks** to content-addressed blobs. Some data processing pipelines need write access to files.

## Unlock

Materializes a symlink into a writable copy:

```shell
gda unlock sub-01/meg/run.fif
```

After unlock, you can edit the file normally.

## Lock

Re-hashes the unlocked file, stores a new object if changed, and replaces it with a symlink:

```shell
gda lock sub-01/meg/run.fif
```

If the content changed, GDA creates a new object (new hash). The old object remains in the store for previous snapshots.

## Checking Status

`gda status` shows unlocked files:

```
  unlocked  496.0 KB  sub-01/meg/run.fif
  modified  512.0 KB  sub-02/beh/data.csv
```

- `unlocked` = writable copy, content matches stored hash
- `modified` = writable copy, content differs from stored hash

## Locking All Unlocked Files

```shell
gda lock $(gda status | grep unlocked | awk '{print $3}')
```
