package store

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Store manages the content-addressed object store.
type Store struct {
	Root string // path to .gda/
}

// Open opens an existing store, or creates one.
func Open(root string) (*Store, error) {
	s := &Store{Root: filepath.Join(root, ".gda")}
	if err := os.MkdirAll(s.objectsDir(), 0755); err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}
	return s, nil
}

func (s *Store) objectsDir() string {
	return filepath.Join(s.Root, "objects")
}

// Add reads a file, computes SHA256, stores it in objects/, returns the key.
func (s *Store) Add(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	// Write to temp while hashing simultaneously — single read pass
	dest, err := os.CreateTemp(s.objectsDir(), ".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(dest.Name())

	h := sha256.New()
	writer := io.MultiWriter(dest, h)

	if _, err := io.Copy(writer, f); err != nil {
		dest.Close()
		return "", fmt.Errorf("copy: %w", err)
	}
	dest.Close()

	key := fmt.Sprintf("%x", h.Sum(nil))
	if s.Exists(key) {
		return key, nil
	}

	objPath := s.objectPath(key)

	if err := os.MkdirAll(filepath.Dir(objPath), 0755); err != nil {
		return "", fmt.Errorf("create object dir: %w", err)
	}

	if err := os.Rename(dest.Name(), objPath); err != nil {
		return "", fmt.Errorf("rename object: %w", err)
	}

	if err := os.Chmod(objPath, 0444); err != nil {
		return "", fmt.Errorf("protect object: %w", err)
	}

	return key, nil
}

// ObjectPath returns the filesystem path for a given key.
func (s *Store) ObjectPath(key string) string {
	return s.objectPath(key)
}

func (s *Store) objectPath(key string) string {
	return filepath.Join(s.objectsDir(), key[:2], key[2:])
}

// Verify checks that the stored object's content matches its key (hash).
func (s *Store) Verify(key string) (bool, error) {
	f, err := os.Open(s.objectPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	computed := fmt.Sprintf("%x", h.Sum(nil))
	return computed == key, nil
}

// Size returns the size of a stored object in bytes.
func (s *Store) Size(key string) (int64, error) {
	fi, err := os.Stat(s.objectPath(key))
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// Exists checks if an object exists locally.
func (s *Store) Exists(key string) bool {
	_, err := os.Stat(s.objectPath(key))
	return err == nil
}
