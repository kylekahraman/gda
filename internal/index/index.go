package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Entry struct {
	Path      string `json:"path"`
	Key       string `json:"key"`
	Size      int64  `json:"size,omitempty"`
	MTime     int64  `json:"mtime,omitempty"`
	AddedTime int64  `json:"added_at,omitempty"`
}

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

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("files"))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	idx := &Index{db: db, path: dbPath}

	if _, err := os.Stat(jsonPath); err == nil {
		var count int
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("files"))
			if b != nil {
				count = b.Stats().KeyN
			}
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
	return nil
}

func (idx *Index) Close() error {
	return idx.db.Close()
}
