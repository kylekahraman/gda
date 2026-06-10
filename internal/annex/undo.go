package annex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UndoEntry stores the previous state of a file before a destructive operation.
type UndoEntry struct {
	Key   string `json:"key"`
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime"`
}

// UndoStore holds undo state in .gda/undo.json.
type UndoStore struct {
	Entries map[string]*UndoEntry `json:"entries"`
	path    string
}

func loadUndoStore(root string) (*UndoStore, error) {
	path := filepath.Join(root, ".gda", "undo.json")
	u := &UndoStore{
		Entries: make(map[string]*UndoEntry),
		path:    path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return u, nil
		}
		return nil, fmt.Errorf("read undo: %w", err)
	}
	if err := json.Unmarshal(data, u); err != nil {
		return nil, fmt.Errorf("parse undo: %w", err)
	}
	return u, nil
}

func (u *UndoStore) save() error {
	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal undo: %w", err)
	}
	// Atomic write
	tmp := u.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write undo tmp: %w", err)
	}
	if err := os.Rename(tmp, u.path); err != nil {
		return fmt.Errorf("rename undo: %w", err)
	}
	return nil
}

func (u *UndoStore) record(relPath, key string, size, mtime int64) {
	u.Entries[relPath] = &UndoEntry{Key: key, Size: size, MTime: mtime}
}

func (u *UndoStore) pop(relPath string) *UndoEntry {
	entry := u.Entries[relPath]
	delete(u.Entries, relPath)
	return entry
}

func (u *UndoStore) isEmpty() bool {
	return len(u.Entries) == 0
}

// removeFile deletes the undo.json file entirely.
func (u *UndoStore) removeFile() error {
	return os.Remove(u.path)
}
