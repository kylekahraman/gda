# Spec: BoltDB Index Migration

**Replace JSON index (`index/index.go`) with BoltDB-backed index.**

## Why

The JSON index rewrites the entire file on every mutation. For repos with
10,000+ files, this becomes slow and wastes I/O. BoltDB gives O(log n)
writes, O(1) lookups, ACID transactions, and no CGo dependency.

## Interface to Preserve

The `Index` struct must keep these exact exported methods with the same
signatures — callers in `internal/annex/annex.go` depend on them:

```go
func Open(root string) (*Index, error)
func (idx *Index) Set(path, key string, size, mtime int64)
func (idx *Index) Remove(path string)
func (idx *Index) Get(path string) *Entry     // returns nil if not found
func (idx *Index) All() map[string]*Entry     // returns copy
func (idx *Index) Count() int
func (idx *Index) Save() error                // BoltDB: no-op, writes are immediate
func (idx *Index) Close() error
```

The `Entry` struct stays exactly the same:

```go
type Entry struct {
    Path      string `json:"path"`
    Key       string `json:"key"`
    Size      int64  `json:"size,omitempty"`
    MTime     int64  `json:"mtime,omitempty"`
    AddedTime int64  `json:"added_at,omitempty"`
}
```

Remove: `sync.RWMutex`, `ByPath map[string]*Entry`, `dirty bool` fields.

## BoltDB Schema

- File: `.gda/index.db`
- Bucket name: `"files"`
- Key: file path (string, UTF-8)
- Value: JSON-encoded `Entry` struct

## Migration from JSON

On `Open()`, if `.gda/index.json` exists:
1. Open/create BoltDB at `.gda/index.db`
2. Create `"files"` bucket
3. Check if BoltDB bucket is empty (`Stats().KeyN == 0`)
4. If empty AND `index.json` exists → read JSON, write all entries to BoltDB
5. Delete `index.json` after successful migration
6. If BoltDB already has data → skip migration, use BoltDB as-is

## Implementation Pseudocode

```go
package index

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
    bolt "go.etcd.io/bbolt"
)

type Index struct {
    db   *bolt.DB
    path string
}

func Open(root string) (*Index, error) {
    dbPath := filepath.Join(root, ".gda", "index.db")
    jsonPath := filepath.Join(root, ".gda", "index.json")

    db, err := bolt.Open(dbPath, 0600, nil)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    // Ensure bucket exists
    err = db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("files"))
        return err
    })
    if err != nil {
        db.Close()
        return nil, err
    }

    idx := &Index{db: db, path: dbPath}

    // Migrate from JSON if present
    if _, err := os.Stat(jsonPath); err == nil {
        var count int
        db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte("files"))
            count = b.Stats().KeyN
            return nil
        })
        if count == 0 {
            data, err := os.ReadFile(jsonPath)
            if err == nil {
                var old struct {
                    ByPath map[string]*Entry `json:"by_path"`
                }
                if json.Unmarshal(data, &old) == nil {
                    db.Update(func(tx *bolt.Tx) error {
                        b := tx.Bucket([]byte("files"))
                        for path, entry := range old.ByPath {
                            v, _ := json.Marshal(entry)
                            b.Put([]byte(path), v)
                        }
                        return nil
                    })
                }
            }
            os.Remove(jsonPath)
        }
    }

    return idx, nil
}

func (idx *Index) Set(path, key string, size, mtime int64) {
    entry := &Entry{
        Path:      path,
        Key:       key,
        Size:      size,
        MTime:     mtime,
        AddedTime: time.Now().Unix(),
    }
    value, _ := json.Marshal(entry)
    idx.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("files"))
        return b.Put([]byte(path), value)
    })
}

func (idx *Index) Remove(path string) {
    idx.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("files"))
        return b.Delete([]byte(path))
    })
}

func (idx *Index) Get(path string) *Entry {
    var entry *Entry
    idx.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("files"))
        v := b.Get([]byte(path))
        if v != nil {
            entry = new(Entry)
            json.Unmarshal(v, entry)
        }
        return nil
    })
    return entry
}

func (idx *Index) All() map[string]*Entry {
    result := make(map[string]*Entry)
    idx.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("files"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            var entry Entry
            if json.Unmarshal(v, &entry) == nil {
                result[string(k)] = &entry
            }
        }
        return nil
    })
    return result
}

func (idx *Index) Count() int {
    var count int
    idx.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("files"))
        count = b.Stats().KeyN
        return nil
    })
    return count
}

func (idx *Index) Save() error {
    return nil // BoltDB writes are immediate
}

func (idx *Index) Close() error {
    return idx.db.Close()
}
```

## Files to Change

- `internal/index/index.go` — full rewrite (keep same package, same exported API)
- `go.mod` — add `go.etcd.io/bbolt` dependency
- `go.sum` — auto-generated by `go mod tidy`
- `internal/index/snapshot.go` — no changes needed (uses `Entry` struct, not `Index`)
- `internal/annex/annex.go` — no changes (uses `Index` interface, not internal fields)

## Files to Remove

None. Old `.gda/index.json` is deleted during migration.

## Verification

1. `go mod tidy` must succeed
2. `go build ./cmd/gda/` must succeed
3. `cd /tmp && rm -rf test && mkdir test && cd test && echo data > f.txt && gda init && gda add f.txt && gda status` must work
4. `gda snapshot s1 && gda checkout s1` must work
5. `gda mv f.txt g.txt && gda status` must show `g.txt`
6. `gda fsck` must report 1 ok
7. Verify `.gda/index.db` exists and is non-empty
8. Verify `.gda/index.json` does NOT exist (was migrated and deleted)
