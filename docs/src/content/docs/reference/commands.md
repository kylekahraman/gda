---
title: Commands
description: Complete GDA command reference
---

## `gda init`

Initialize a GDA repository in the current directory.

## `gda add <path> [<path> ...]`

Track files or directories. Directories are walked recursively.

## `gda status [<path> ...]`

Show tracked files with integrity status: `ok`, `unlocked`, `modified`, `missing`, `broken`, `corrupt`.

## `gda mv <source> <dest>`

Rename a tracked file or directory prefix.

## `gda rm <path> [<path> ...]`

Untrack files. Content stays in the store.

## `gda snapshot [<name>]`

Create a named checkpoint of the current index. Auto-names with timestamp if no name given.

## `gda log`

List all snapshots with creation time, file count, and total size.

## `gda checkout <snapshot-name>`

Restore working tree to a snapshot state.

## `gda unlock <path> [<path> ...]`

Materialize symlinks into writable copies for editing.

## `gda lock <path> [<path> ...]`

Re-hash unlocked files, store new content, restore symlinks.

## `gda gc [--dry-run]`

Remove unreferenced objects from the store.

## `gda fsck`

Check and repair working tree integrity.

## `gda remote add <name> <url>`

Add a remote rsync target.

## `gda remote list`

List configured remotes.

## `gda push <remote-name>`

Sync local objects to remote.

## `gda pull <remote-name>`

Sync remote objects to local.

## `gda help [<command>]`

Show help for a specific command.

## Global Flags

| Flag | Description |
|------|-------------|
| `GDA_DEV=1` | Enable debug logging to stderr |
