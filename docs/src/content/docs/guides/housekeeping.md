---
title: Housekeeping
description: Garbage collection and integrity checking
---

## Garbage Collection

Remove objects that are no longer referenced by any snapshot or the current index:

```shell
# Preview
gda gc --dry-run

# Execute
gda gc
```

Objects younger than 1 hour are never removed. This grace period prevents race conditions with concurrent operations.

## Integrity Check

Scan your entire project and verify everything is intact:

```shell
gda fsck
```

Checks for:
- Broken symlinks (symlink exists but target is missing)
- Missing files (no symlink or file at all)
- Corrupt objects (stored content doesn't match hash)
- Modified unlocked files (content changed since lock)

```
17 ok, 0 broken, 0 fixed, 0 modified, 0 missing
```

## Repair

`gda fsck` automatically repairs broken symlinks when the object still exists in the store.
