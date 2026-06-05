package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Snapshot struct {
	Name      string           `json:"name"`
	CreatedAt int64            `json:"created_at"`
	Entries   map[string]*Entry `json:"entries"`
}

type snapshots struct {
	dir string
}

func SnapshotsDir(root string) string {
	return filepath.Join(root, ".gda", "snapshots")
}

func OpenSnapshots(root string) (*snapshots, error) {
	dir := SnapshotsDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create snapshots dir: %w", err)
	}
	return &snapshots{dir: dir}, nil
}

func (s *snapshots) Create(name string, entries map[string]*Entry) error {
	path := filepath.Join(s.dir, name+".json")
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("snapshot %q already exists", name)
	}

	snap := Snapshot{
		Name:      name,
		CreatedAt: time.Now().Unix(),
		Entries:   entries,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	return nil
}

func (s *snapshots) List() ([]Snapshot, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snaps []Snapshot
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		snap, err := s.Load(e.Name()[:len(e.Name())-5])
		if err != nil {
			continue
		}
		snaps = append(snaps, *snap)
	}

	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].CreatedAt < snaps[j].CreatedAt
	})
	return snaps, nil
}

func (s *snapshots) Load(name string) (*Snapshot, error) {
	path := filepath.Join(s.dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot %q: %w", name, err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("parse snapshot %q: %w", name, err)
	}
	return &snap, nil
}
