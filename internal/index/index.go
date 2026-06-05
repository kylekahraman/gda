package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single tracked file.
type Entry struct {
	Path      string `json:"path"`
	Key       string `json:"key"`
	Size      int64  `json:"size,omitempty"`
	MTime     int64  `json:"mtime,omitempty"` // unix timestamp
	AddedTime int64  `json:"added_at,omitempty"`
}

// Index maintains the path → object key mapping.
type Index struct {
	mu     sync.RWMutex
	path   string
	ByPath map[string]*Entry `json:"by_path"`
	dirty  bool
}

// Open loads the index from disk, or creates a new one.
func Open(root string) (*Index, error) {
	path := filepath.Join(root, ".gda", "index.json")
	idx := &Index{
		path:   path,
		ByPath: make(map[string]*Entry),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, fmt.Errorf("read index: %w", err)
	}

	if err := json.Unmarshal(data, idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	return idx, nil
}

// Set adds or updates an entry in the index.
func (idx *Index) Set(path, key string, size, mtime int64) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.ByPath[path] = &Entry{
		Path:      path,
		Key:       key,
		Size:      size,
		MTime:     mtime,
		AddedTime: time.Now().Unix(),
	}
	idx.dirty = true
}

// Remove deletes an entry from the index.
func (idx *Index) Remove(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.ByPath, path)
	idx.dirty = true
}

// Get returns the entry for a path, or nil if not found.
func (idx *Index) Get(path string) *Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.ByPath[path]
}

// All returns all entries.
func (idx *Index) All() map[string]*Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	cp := make(map[string]*Entry, len(idx.ByPath))
	for k, v := range idx.ByPath {
		cp[k] = v
	}
	return cp
}

// Count returns the number of tracked entries.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.ByPath)
}

// Save writes the index to disk if dirty.
func (idx *Index) Save() error {
	idx.mu.RLock()
	if !idx.dirty {
		idx.mu.RUnlock()
		return nil
	}
	idx.mu.RUnlock()

	idx.mu.Lock()
	defer idx.mu.Unlock()

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	dir := filepath.Dir(idx.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create index dir: %w", err)
	}

	tmpPath := idx.path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create temp index: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(tmpPath)
	}()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write temp index: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync temp index: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp index: %w", err)
	}

	if err := os.Rename(tmpPath, idx.path); err != nil {
		return fmt.Errorf("rename index: %w", err)
	}

	idx.dirty = false
	return nil
}

// Close saves the index.
func (idx *Index) Close() error {
	return idx.Save()
}
