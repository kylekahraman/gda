# Spec: Unlock/Lock — Hybrid Working Tree Mode

## Problem

GDA's default mode replaces tracked files with symlinks to read-only
objects (0444). This breaks:
- Pipelines that write through to input files (FreeSurfer, FSL)
- HPC filesystems that reject symlinks
- Users who want to edit a file directly

A FUSE mount solves this but requires kernel modules, root/admin
access, and thousands of lines of Go.

## Solution: Unlock/Lock

Two commands that toggle between symlink and materialized modes per file:

```
gda unlock file.txt   → replace symlink with writable content copy
gda lock file.txt     → re-hash, store if changed, replace with symlink
```

Default state is **locked** (symlink — current behavior, instant ops).
Users unlock only the files they need to modify. Checkout and status
handle both modes transparently.

## Files to Change

### 1. `internal/store/store.go` — Add CopyTo method

Needed to materialize object content to a working-tree path:

```go
func (s *Store) CopyTo(key, dest string) error
```

Reads from `s.objectPath(key)`, writes to `dest`, preserves the content.
Not hashing, not storing — pure copy.

### 2. `internal/index/index.go` — Add Unlocked flag to Entry

```go
type Entry struct {
    Path      string `json:"path"`
    Key       string `json:"key"`
    Size      int64  `json:"size,omitempty"`
    MTime     int64  `json:"mtime,omitempty"`
    AddedTime int64  `json:"added_at,omitempty"`
    Unlocked  bool   `json:"unlocked,omitempty"` // true after gda unlock
}
```

No migration needed — missing field defaults to false (locked) for
existing entries.

### 3. `internal/annex/annex.go` — Add Unlock/Lock + checkout and status updates

### 4. `internal/annex/commands.go` — Add Unlock/Lock wrappers

### 5. `cmd/gda/main.go` — Register unlock/lock commands

## Unlock Logic

```
gda unlock <file> [file...]
```

For each file:
1. Look up entry in index. Error if not tracked.
2. Check if already unlocked (Entry.Unlocked == true). Skip if so.
3. Lstat the path. If not a symlink, warn and skip (user already modified it).
4. Copy object content from `.gda/objects/XX/hash` to tracked path.
5. Set Entry.Unlocked = true in index.
6. Print confirmation.

Important: do NOT remove the symlink first and then copy — that would
leave a window where the file is missing. Instead:
- Read object content into memory
- os.Remove(symlink)
- os.WriteFile(path, content, 0644)
- (Small files only — if this is a concern, use temp file + rename)

### Edge cases

- File already unlocked → skip with info message
- File was already modified (not a symlink) → warn "file is modified, not
  a symlink" and skip
- Object missing from store → error "object not found, run gda add first"
- Path is a directory → error
- Multiple files in one command → process each independently

## Lock Logic

```
gda lock <file> [file...]
```

For each file:
1. Look up entry in index. Error if not tracked.
2. Check if already locked (Entry.Unlocked == false). Skip if so.
3. Lstat the path. If it's a symlink, warn and skip (already locked).
4. Re-hash and store the file via Store.Add(). If content changed,
   Store.Add stores the new blob and returns a new key.
5. Update index: set new key/size/mtime, set Unlocked = false.
6. Replace file with symlink to the (possibly new) object.
7. Print confirmation.

### Edge cases

- File already locked → skip with info message
- File was deleted between unlock and lock → error "file not found"
  (user has to `gda rm` to clean up)

## Status Display Updates

`gda status` currently shows `modified` when a tracked file is not a
symlink. With unlock/lock, this changes:

- File is unlocked (Entry.Unlocked == true) AND is a regular file → `unlocked`
- File is unlocked AND content hash changed → `modified`
- File is unlocked BUT re-locked (is a symlink) → treat as locked, `ok`

In `fileStatus()`:

```go
func (g *GDA) fileStatus(rootAbs, path, key string, entry *index.Entry) string {
    abs := filepath.Join(rootAbs, path)

    fi, err := os.Lstat(abs)
    if os.IsNotExist(err) {
        return "missing"
    }

    if fi.Mode()&os.ModeSymlink == 0 {
        // Regular file (or other non-symlink)
        if entry != nil && entry.Unlocked {
            // Check if content matches stored key
            if g.Store.VerifyAtPath(key, abs) {
                return "unlocked"
            }
            return "modified"
        }
        return "modified"
    }

    // Symlink
    if _, err := os.Stat(abs); err != nil {
        return "broken"
    }

    if !g.Store.Exists(key) {
        return "missing"
    }
    if ok, _ := g.Store.Verify(key); !ok {
        return "corrupt"
    }
    return "ok"
}
```

Note: `fileStatus` signature changes to accept `*index.Entry` — update
callers in Status/Fsck.

## Checkout Updates

Checkout currently removes files that exist in current index but not
in the target snapshot. With unlock/lock:

1. If current file is unlocked and user has made changes, warn before
   overwriting: "file is unlocked with modifications, skipping".
   The user must `gda lock` or `gda rm` first.
2. If current file is unlocked but content matches stored key (no
   modifications), remove the regular file and restore symlink from
   snapshot normally.
3. If current file is locked (symlink), existing behavior applies.

### Determining "has modifications"

Compare `sha256(file)` against `entry.Key`. Two options:
- Add `Store.VerifyAtPath(key, absPath)` that hashes a given path and
  compares against the key. Reuse existing Verify logic.
- Compute hash inline.

Add to `store.go`:

```go
// VerifyAtPath checks if the content at a given path matches the key.
func (s *Store) VerifyAtPath(key, path string) (bool, error) {
    f, err := os.Open(path)
    if err != nil {
        return false, err
    }
    defer f.Close()
    h := sha256.New()
    io.Copy(h, f)
    return fmt.Sprintf("%x", h.Sum(nil)) == key, nil
}
```

## Summary of Changes

| File | Change |
|------|--------|
| `internal/store/store.go` | Add `CopyTo(key, dest)` and `VerifyAtPath(key, path)` |
| `internal/index/index.go` | Add `Unlocked bool` to Entry |
| `internal/annex/annex.go` | Add `Unlock()`, `Lock()` methods. Update `fileStatus()` to show unlocked/modified. Update `Checkout()` to handle unlocked files safely. Update `Fsck()` to use new status logic. |
| `internal/annex/commands.go` | Add `Unlock()`, `Lock()` wrappers |
| `cmd/gda/main.go` | Add `unlock`, `lock` cases |

## Verification

```
cd /tmp && rm -rf test && mkdir test && cd test
echo "data" > f.txt
gda init && gda add f.txt
gda status                     → "ok"
gda unlock f.txt
gda status                     → "unlocked"
echo "modified" >> f.txt
gda status                     → "modified"
gda lock f.txt
gda status                     → "ok"
cat f.txt                      → "data\nmodified\n"
# Symlink behavior preserved for restructure
gda mv f.txt g.txt
gda status                     → "ok" for g.txt
gda checkout snap1             → restores f.txt as symlink
gda status                     → "ok" for f.txt
```
