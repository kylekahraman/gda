# GDA

GDA tracks files in a content-addressed store and keeps your working
tree as symlinks. Renames and restructures only update the index and
symlink, so content stays in place and operations complete regardless
of file size.

## Quick start

    gda init
    gda add bids/
    gda status
    gda mv bids/ sourcedata/
    gda snapshot experiment1
    gda checkout experiment1

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
| `fsck`     | Scan working tree and repair symlinks  |
| `gc`       | Remove unreferenced blobs              |

## How it works

- Files are stored as SHA256-addressed blobs under `.gda/objects/`
- The working tree holds symlinks to those blobs
- A JSON index maps each tracked path to its content hash
- Snapshots are named copies of the index stored as separate JSON files
- Written in Go, statically linked

## Status

Still in active development. Core operations work. Platforms beyond
Linux (Windows, macOS) and remote backends (rsync, S3) are planned
but not yet implemented.

## License

MIT
