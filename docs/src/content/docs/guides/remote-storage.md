---
title: Remote Storage
description: Sync your data with remote storage
---

GDA uses rsync for remote storage. Remotes are defined in `.gdarc`.

## Adding a Remote

```shell
gda remote add origin rsync://server/path/to/store
```

The remote path must point to an empty or existing GDA object store.

## Pushing Objects

```shell
gda push origin
```

Copies all local objects not present on the remote to the remote store.

## Pulling Objects

```shell
gda pull origin
```

Copies all remote objects not present locally to the local store. Also syncs the index.

## Checking Remote Status

```shell
gda remote list
```

Lists configured remotes.
