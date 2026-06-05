---
title: Use Cases
description: Who GDA is for and what problems it solves
---

## You download datasets from OpenNeuro

You find a cool dataset, download it, start preprocessing. A month later you realize a file got corrupted somewhere, or you can't remember which preprocessing version produced those stats. You wish you'd snapped a bookmark.

```shell
gda init
gda add bids/
gda snapshot downloaded-v1
# ... run preprocessing ...
gda snapshot preprocessed-v2
gda log
# Months later:
gda fsck    # verify nothing is corrupted
gda checkout downloaded-v1  # go back to raw data
```

## You share data with collaborators

Your labmate needs the latest preprocessing output. You could zip it, upload it,
send a link, wait for "hey the file structure is different." Or:

```shell
gda snapshot analysis-v3
gda push origin
```

They pull the same snapshot. Your file tree is their file tree. No "which version do you have?" discussions.

## You archive data with integrity

Hard drives fail. Files corrupt silently. With GDA:

```shell
gda add archive/
gda snapshot archive-v1
gda fsck    # anytime: check every file against its hash
```

If a file changes or gets corrupted, `gda fsck` catches it. You don't find out by getting a cryptic error in the middle of a conference talk.

## You restructure your data constantly

Your BIDS layout changes. Subject folders get renamed. Task labels get fixed.
With raw files, renaming a 500 GB directory means waiting for a copy. With GDA:

```shell
gda mv bids/wrong-name/ bids/right-name/
# Instant. No data copied. Just symlinks.
```

## You want to share a dataset publicly

You've collected a dataset and want to share it on OSF, GIN, or your lab server.
GDA's remote push lets you sync to any rsync target. Collaborators pull your
snapshots and get exactly the same file tree you have.

---

**GDA is not for you if:**
- You only work alone on small files (< 1 GB) and never share data
- You need concurrent multi-user editing (not yet supported)
- You need S3 or HTTP remotes (rsync only for now)
