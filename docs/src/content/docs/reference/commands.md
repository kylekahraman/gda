---
title: Commands
description: Complete GDA command reference
---

## Core Commands

### `gda init`

Initialize a GDA repository in the current directory.

```shell
gda init
```

Creates `.gda/` with the object store and index. Safe to run on an existing repo (no-op if already initialized).

---

### `gda add <path> [<path> ...]`

Track files or directories in the store. Directories are walked recursively.

```shell
gda add sub-01/meg/run.fif
gda add sub-02/
gda add bids/ derivatives/
```

**Behavior:**
- Computes SHA256 hash while copying to store (single pass)
- Replaces each file with a relative symlink to the object
- Skips existing symlinks and directories automatically
- Deduplicates: identical files share one object

---

### `gda status [<path> ...]`

Show tracked files and their working tree integrity.

```shell
$ gda status
Tracked files:
  ok  512.0 KB  sub-01/meg/run.fif
  ok  1.2 KB    sub-01/beh/data.csv
  unlocked  496.0 KB  sub-02/meg/run.fif
2 files, 1.0 MB total
```

**Status values:**

| Status | Meaning |
|--------|---------|
| `ok` | Symlink exists, resolves, and content matches hash |
| `unlocked` | Symlink replaced with writable copy, content matches stored hash |
| `modified` | Unlocked file, content differs from stored hash |
| `missing` | No file or symlink at tracked path |
| `broken` | Symlink exists but target is missing |
| `corrupt` | Symlink resolves but stored content doesn't match hash |

---

## Snapshots

### `gda snapshot [<name>]`

Create a named checkpoint of the current index.

```shell
$ gda snapshot
Snapshot "snapshot-1749134400" saved (143 files, 12.3 GB)

$ gda snapshot preprocessing-v3
Snapshot "preprocessing-v3" saved (158 files, 12.5 GB)
```

Auto-generates a timestamp-based name if none is provided.

---

### `gda log`

List all snapshots with details.

```shell
$ gda log
  preprocessing-v3  2026-06-05 12:40:00  158 files  12.5 GB
  preprocessing-v2  2026-06-03 09:15:22  143 files  12.3 GB
  preprocessing-v1  2026-06-01 14:32:05  127 files  11.8 GB
  raw-v1            2026-05-28 11:00:00  112 files  11.2 GB
```

Most recent snapshot appears first.

---

### `gda checkout <snapshot-name>`

Restore the working tree to a snapshot state.

```shell
$ gda checkout raw-v1
Restored "raw-v1" (112 files)
```

**Safety:**
- Removes files tracked in the current index but absent in the snapshot
- Warns about unlocked modified files without overwriting them
- Creates parent directories automatically

---

## Maintenance

### `gda gc [--dry-run]`

Remove objects not referenced by any snapshot or the current index.

```shell
$ gda gc --dry-run
Would remove 12 objects (45.2 MB), keep 158 objects

$ gda gc
Removed 12 objects (45.2 MB), keeping 158 objects
```

Objects younger than 1 hour are never removed (grace period prevents race conditions).

---

### `gda fsck`

Scan the index and repair working tree issues.

```shell
$ gda fsck
158 ok, 0 broken, 0 fixed, 0 modified, 0 missing
```

Automatically repairs broken symlinks when the object exists in the store.

---

## File Editing

### `gda unlock <path> [<path> ...]`

Materialize a symlink into a writable copy for editing.

```shell
$ gda unlock sub-01/beh/data.csv
Unlocked sub-01/beh/data.csv
$ gda status
  unlocked  1.2 KB  sub-01/beh/data.csv
```

---

### `gda lock <path> [<path> ...]`

Re-hash an unlocked file, store new content if changed, restore symlink.

```shell
$ gda lock sub-01/beh/data.csv
Locked sub-01/beh/data.csv
```

If content changed, a new object is created. The old object persists for previous snapshots.

---

## File Management

### `gda mv <source> <dest>`

Rename a tracked file or directory prefix.

```shell
$ gda mv sub-01/meg/run.fif sub-01/meg/run-01.fif
Moved sub-01/meg/run.fif -> sub-01/meg/run-01.fif

$ gda mv bids/raw/ bids/sourcedata/
Moved 12 files from bids/raw/ -> bids/sourcedata/
```

**Safety:** Aborts if the source is not a symlink (prevents data loss with modified files).

---

### `gda rm <path> [<path> ...]`

Untrack files. Content stays in the store (use `gc` to reclaim space).

```shell
$ gda rm sub-02/temp/scratch.nii.gz
Removed sub-02/temp/scratch.nii.gz from index (content retained in store)
```

---

## Remote Operations

### `gda remote add <name> <url>`

Add a remote rsync target.

```shell
gda remote add origin rsync://server/path/to/store
```

### `gda remote list`

Show configured remotes.

```shell
$ gda remote list
origin: rsync://server/path/to/store
```

### `gda push <remote-name>`

Sync local objects to remote.

```shell
gda push origin
```

Copies only objects not present on the remote.

### `gda pull <remote-name>`

Sync remote objects to local.

```shell
gda pull origin
```

---

## Help

### `gda help [<command>]`

Show help for a specific command.

```shell
gda help add
gda help snapshot
```

Without arguments, shows the full command list.

## Global Flags

| Flag | Description |
|------|-------------|
| `GDA_DEV=1` | Enable debug logging to stderr with timestamps and operation durations |
