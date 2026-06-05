---
title: Quick Start
description: Get started with GDA in 5 minutes
---

Initialize a repository, add data, and create a snapshot — all in a few commands.

## 1. Initialize

```shell
cd /path/to/your/dataset
gda init
```

Creates a `.gda/` directory with the object store and index.

## 2. Add Data

Add individual files or entire directories:

```shell
gda add sub-01/
gda add sub-02/meg/run-01.fif
```

Files are hashed, stored in `.gda/objects/`, and replaced with symlinks.

## 3. Check Status

```shell
gda status
```

Shows all tracked files with their integrity status (`ok`, `missing`, `modified`, `broken`, `corrupt`).

## 4. Create a Snapshot

```shell
gda snapshot raw-v1
```

Saves the current index as a named checkpoint.

## 5. View History

```shell
gda log
```

Lists all snapshots.

## 6. Restore

```shell
gda checkout raw-v1
```

Restores the working tree to the snapshot state.

## Full Example

```shell
cd /tmp && mkdir test-dataset && cd test-dataset
mkdir -p sub-01/meg sub-01/beh
echo "test data" > sub-01/meg/run.fif
gda init
gda add sub-01/
gda status
gda snapshot initial
gda log
```
