# Spec: Remote Push/Pull (rsync backend)

## Problem

GDA is local-only. Researchers need to sync datasets between:
- Workstation → lab server
- Lab server → HPC cluster
- Laptop → desktop

Without remotes, GDA is unusable for real workflows.

## Design

Add three commands for remote sync:

```
gda remote add <name> <url>    → register a remote
gda push [remote]              → upload missing local objects
gda pull [remote]              → download missing remote objects
```

### Remote URL format

```
user@host:/path/to/repo     → SSH rsync
/path/to/repo               → local path (USB drive, mounted NAS)
```

S3/HTTP planned but not for v1.

### Remote storage

In `index` package, add a BoltDB bucket `"remotes"`:

```
Bucket: remotes
Key: remote name (string, e.g. "origin")
Value: JSON Remote struct
```

```go
type Remote struct {
    Name string `json:"name"`
    URL  string `json:"url"`
}
```

Add methods to Index:

```go
func (idx *Index) RemoteAdd(name, url string) error
func (idx *Index) RemoteRemove(name string) error
func (idx *Index) RemoteList() ([]Remote, error)
func (idx *Index) RemoteGet(name string) (*Remote, error)
```

Same BoltDB `"files"` bucket pattern — just a new bucket `"remotes"`.

## Push Logic (simplest version)

```
gda push [remote]
```

1. Resolve remote name (default "origin" if not specified)
2. Build list of all keys referenced by current index + all snapshots
3. For each key, check if remote has it:
   - `ssh user@host "[ -f /path/.gda/objects/XX/hash ]"`
   - Or: `rsync --list-only remote:.gda/objects/XX/hash`
4. If missing on remote, rsync the object:
   - `rsync -aR .gda/objects/XX/hash user@host:/path/`
   - The `-aR` preserves the relative path so the XX/ hash structure works

Optimization: batch rsync calls. Instead of one rsync per file, collect
all missing objects and run one rsync:

```
rsync -aR .gda/objects/XX/hash1 .gda/objects/XX/hash2 ... user@host:/path/
```

But for v1, even checking one-by-one with `rsync --list-only` works for
prototype. Use a local key list.

### Better approach: manifest-based sync

1. Build a set of all referenced keys (index + snapshots)
2. Write a temp manifest file: one key per line
3. On remote, list all existing keys → compare
4. Rsync only missing ones

Actually simplest: just rsync the entire `.gda/objects/` directory.
Rsync is smart — it only transfers files that differ. First sync is
full, subsequent syncs are incremental.

```
rsync -a --info=progress2 .gda/objects/ user@host:/path/.gda/objects/
```

This is O(objects) for listing but the data transfer is minimal.
For v1, this is acceptable. For v2, add per-object tracking.

### Files to include in sync

Only sync `.gda/objects/` and `.gda/snapshots/`. Do NOT sync:
- `.gda/index.db` (working state, may conflict)
- `.gda/remotes/` (working state)

The remote constructs its own index from the snapshots on checkout.

## Pull Logic

```
gda pull [remote]
```

1. Same remote resolution
2. Rsync from remote to local:
   ```
   rsync -a --info=progress2 user@host:/path/.gda/objects/ .gda/objects/
   rsync -a user@host:/path/.gda/snapshots/ .gda/snapshots/
   ```
3. After objects arrive, `gda status` verifies integrity
4. User can `gda checkout <snapshot>` to materialize a working tree

## Files to Change

| File | Change |
|------|--------|
| `internal/index/index.go` | Add `"remotes"` bucket + Remote CRUD methods |
| `internal/annex/annex.go` | Add `RemoteAdd()`, `Push()`, `Pull()` methods |
| `internal/annex/commands.go` | Add `RemoteAdd()`, `Push()`, `Pull()` wrappers |
| `cmd/gda/main.go` | Add `remote`, `push`, `pull` commands |

## Verify

```
cd /tmp && rm -rf local remote && mkdir local remote

# Init local repo
cd local && echo "data" > f.txt && gda init && gda add f.txt && gda snapshot snap1

# Add remote pointing to /tmp/remote
gda remote add origin /tmp/remote

# Push
gda push origin
ls /tmp/remote/.gda/objects/  ← should have objects

# Pull on fresh remote
cd /tmp/remote && gda init
gda remote add origin /tmp/local
gda pull origin
gda checkout snap1
cat f.txt  ← "data"
```
