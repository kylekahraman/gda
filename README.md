# GDA

[![Docs](https://img.shields.io/badge/docs-starlight-blue?style=flat-square)](https://kylekahraman.github.io/gda/)
[![Release](https://img.shields.io/github/v/release/kylekahraman/gda?style=flat-square&color=green)](https://github.com/kylekahraman/gda/releases)

Content-addressed data versioning for research datasets.

GDA tracks files in a content-addressed store and keeps your working tree as symlinks. Renames and restructures only update the index and symlink — content stays in place regardless of file size.

## Quick start

```
gda init
gda add bids/
gda status
gda mv bids/ sourcedata/
gda snapshot experiment1
gda checkout experiment1
```

## Commands

| Command    | Description                            |
|------------|----------------------------------------|
| `init`     | Initialize a repo in the current dir   |
| `add`      | Track files or directories (recursive) |
| `status`   | Show tracked files and integrity       |
| `mv`       | Rename a tracked path or directory     |
| `rm`       | Untrack a file (content stays stored)  |
| `snapshot` | Save a named checkpoint                |
| `log`      | List snapshots                         |
| `checkout` | Restore working tree from a snapshot   |
| `unlock`   | Materialize symlink for editing        |
| `lock`     | Re-hash and restore symlink            |
| `fsck`     | Scan working tree and repair symlinks  |
| `gc`       | Remove unreferenced blobs              |
| `remote`   | Manage rsync remotes                   |
| `push`     | Sync objects to remote                 |
| `pull`     | Sync objects from remote               |

## How it works

- Files stored as SHA256-addressed blobs under `.gda/objects/`
- Working tree holds symlinks to those blobs
- BoltDB index maps each tracked path to its content hash
- Snapshots are named copies of the index
- Written in Go, statically linked, single binary

## Documentation

Full documentation is available at [kylekahraman.github.io/gda/](https://kylekahraman.github.io/gda/)

## Status

Pre-alpha (v0.0.1). Core operations work on Linux and macOS. Remote via rsync. Remote backends (S3, SSH) and Windows support are planned.

## License

MIT
