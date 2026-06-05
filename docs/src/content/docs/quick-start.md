---
title: Quick Start
description: Get started with GDA in 5 minutes
---

Follow along in your terminal. Every command below shows what you'll see.

## Create a test dataset

```shell
mkdir -p ~/gda-test/sub-01/meg ~/gda-test/sub-01/beh
cd ~/gda-test
echo "simulated MEG recording" > sub-01/meg/run.fif
echo "trial,rt,accuracy" > sub-01/beh/data.csv
echo "1,450,1" >> sub-01/beh/data.csv
echo "2,520,0" >> sub-01/beh/data.csv
```

## 1. Initialize

```shell
gda init
```

Output:
```
Initialized GDA store in .gda
Store: .gda
Index: 0 entries
```

A `.gda/` directory is created. This is where GDA stores everything.

## 2. Add data

```shell
gda add sub-01/
```

Output:
```
Added sub-01/meg/run.fif (21 B, SHA256: a1b2c3d4e5f6...)
Added sub-01/beh/data.csv (39 B, SHA256: f6e5d4c3b2a1...)
```

Each file is hashed (SHA256), stored in `.gda/objects/`, and the original is replaced with a symlink pointing to the object.

## 3. Check status

```shell
gda status
```

Output:
```
Tracked files:
  ok  21 B   sub-01/meg/run.fif
  ok  39 B   sub-01/beh/data.csv

2 files, 60 B total
```

Every file shows `ok` if everything is intact.

## 4. Create a snapshot

```shell
gda snapshot raw-v1
```

Output:
```
Snapshot "raw-v1" saved (2 files, 60 B)
```

A snapshot records every tracked file's path and hash. It's like a bookmark you can return to.

## 5. View history

```shell
gda log
```

Output:
```
  raw-v1  2026-06-05 12:30:00  2 files  60 B
```

All your snapshots, with timestamps and sizes.

## 6. Modify and snapshot again

```shell
gda unlock sub-01/beh/data.csv
echo "3,380,1" >> sub-01/beh/data.csv
gda lock sub-01/beh/data.csv
gda snapshot raw-v2
gda log
```

Output:
```
Unlocked sub-01/beh/data.csv
Locked sub-01/beh/data.csv
Snapshot "raw-v2" saved (2 files, 70 B)
  raw-v2  2026-06-05 12:31:00  2 files  70 B
  raw-v1  2026-06-05 12:30:00  2 files  60 B
```

## 7. Restore a previous state

```shell
gda checkout raw-v1
gda status
```

Output:
```
Restored "raw-v1" (2 files)
Tracked files:
  ok  21 B   sub-01/meg/run.fif
  ok  39 B   sub-01/beh/data.csv
```

Your working tree is back to exactly how it was after `raw-v1`. The data from `raw-v2` is still in the store — you can switch back any time.

## What's next?

→ [Adding Data](/gda/guides/adding-data/) — track real datasets, handle directories
→ [Snapshots & Checkout](/gda/guides/snapshots/) — version management workflow
→ [Commands](/gda/reference/commands/) — full command reference
