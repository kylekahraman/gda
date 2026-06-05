# GDA — Agent Guide

Go CLI tool for content-addressed data versioning. Designed for research
datasets where git-annex is overkill.

## Hard Rules

- **One concern per change.** Small, focused commits.
- **Read the code before writing anything.** Understand the full picture
  before touching a file.
- **No AI-generated comments** in Go source. Doc comments on exported
  functions only, and only when the function name isn't self-explanatory.
- **No "we" anywhere.** Single-developer project.
- **No mocked tests.** Test against real directory structures and files.
- **Verify with `go build ./cmd/gda/` before committing.**

## Project Structure

```
cmd/gda/main.go              # CLI entry
internal/
  annex/
    annex.go                 # Core logic: add, status, mv, rm, snapshot, log, checkout, gc, fsck
    commands.go              # Thin wrappers: Init(), Add(), Status(), ...
  store/store.go             # Content-addressed blob store (SHA256, 0444, atomic rename)
  index/
    index.go                 # JSON index (path → key mapping)
    snapshot.go              # Named JSON-manifest snapshots
docs/
  decisions/ADR-001-json-vs-boltdb.md
  tickets/                   # Deprecated — use GitHub Issues
```

## Key Design Decisions

- **No git dependency.** GDA manages its own object store and index.
  Git can be layered on top if users want it for metadata.
- **Symlink working tree.** Tracked files are symlinks to read-only
  (0444) blobs in `.gda/objects/XX/fullhash`.
- **Single-read `Store.Add`.** Uses `io.MultiWriter` to hash + copy in
  one pass. No double-read for new files.
- **Directory `mv` is prefix-based.** `gda mv bids/ sourcedata/` moves
  all files whose index path starts with `bids/`. Same cost as single file.
- **JSON index for prototype.** Abstracted behind an interface for future
  BoltDB swap. JSON is fine up to ~10K files.
- **Grace period on GC.** Objects younger than 1 hour are never deleted,
  preventing races with concurrent adds.

## Commands

| Command    | Description                                    |
|------------|------------------------------------------------|
| `init`     | Initialize repo in current directory            |
| `add`      | Track files or directories (recursive)          |
| `status`   | Show tracked files + working-tree integrity     |
| `mv`       | Rename tracked path or directory prefix         |
| `rm`       | Untrack (content stays in store)                |
| `snapshot` | Named checkpoint of current index               |
| `log`      | List snapshots                                  |
| `checkout` | Restore working tree from snapshot              |
| `gc`       | Prune unreferenced objects                      |
| `fsck`     | Scan index vs working tree, repair broken links |

## Testing

Always test against a real directory:

```
cd /tmp && rm -rf test && mkdir test && cd test
mkdir -p sub-01/meg sub-01/beh
echo data > sub-01/meg/run.fif
/path/to/gda init
/path/to/gda add sub-01/
/path/to/gda status
/path/to/gda mv sub-01/ bids/
/path/to/gda snapshot test1
/path/to/gda checkout test1
/path/to/gda gc --dry-run
```

## Next Up

- `fsck` for broken symlink repair
- `status` working-tree integrity checks (broken, missing, modified states)
- Real dataset testing: MEG → fMRI → freesurfer
- rsync remote backend for push/pull
