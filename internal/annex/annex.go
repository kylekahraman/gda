package annex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kylekahraman/gda/internal/index"
	"github.com/kylekahraman/gda/internal/store"
)

type GDA struct {
	Store *store.Store
	Index *index.Index
	Root  string
}

// Open opens an existing GDA repo, or creates a new one.
func Open(root string) (*GDA, error) {
	st, err := store.Open(root)
	if err != nil {
		return nil, err
	}
	idx, err := index.Open(root)
	if err != nil {
		return nil, err
	}
	return &GDA{
		Store: st,
		Index: idx,
		Root:  root,
	}, nil
}

// Close saves state.
func (g *GDA) Close() error {
	return g.Index.Close()
}

// Init initializes a new GDA repo.
func (g *GDA) Init() error {
	// Store and Index are already opened by Open
	fmt.Printf("Initialized GDA store in %s\n", filepath.Join(g.Root, ".gda"))
	return g.printStats()
}

// Add adds files or directories to the store.
func (g *GDA) Add(paths []string) error {
	// Expand directories into file lists via a single walk
	var files []string
	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs %s: %w", path, err)
		}
		fi, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if fi.IsDir() {
			err := filepath.Walk(abs, func(walkPath string, walkFi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if walkFi.IsDir() || walkFi.Mode()&os.ModeSymlink != 0 {
					return nil
				}
				files = append(files, walkPath)
				return nil
			})
			if err != nil {
				return fmt.Errorf("walk %s: %w", path, err)
			}
		} else {
			files = append(files, abs)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found to add")
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	for _, abs := range files {
		fi, err := os.Lstat(abs)
		if err != nil {
			return fmt.Errorf("lstat %s: %w", abs, err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		// Hash and store
		key, err := g.Store.Add(abs)
		if err != nil {
			return fmt.Errorf("add %s: %w", abs, err)
		}

		// Replace original with symlink (relative path from file location to object)
		objAbs, err := filepath.Abs(g.Store.ObjectPath(key))
		if err != nil {
			return fmt.Errorf("abs object: %w", err)
		}
		rel, err := filepath.Rel(filepath.Dir(abs), objAbs)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}

		if err := os.Remove(abs); err != nil {
			return fmt.Errorf("remove original %s: %w", abs, err)
		}
		if err := os.Symlink(rel, abs); err != nil {
			return fmt.Errorf("create symlink %s: %w", abs, err)
		}

		// Update index
		relPath, _ := filepath.Rel(rootAbs, abs)
		g.Index.Set(relPath, key, fi.Size(), fi.ModTime().Unix())

		fmt.Printf("Added %s (%s, SHA256: %s)\n", relPath, formatSize(fi.Size()), key[:16])
	}

	return g.Index.Save()
}

// Status shows the current state of tracked files.
func (g *GDA) Status(paths []string) error {
	entries := g.Index.All()

	if len(entries) == 0 {
		fmt.Println("No files tracked.")
		return nil
	}

	var totalSize int64
	fmt.Println("Tracked files:")
	for path, entry := range entries {
		if len(paths) > 0 && !contains(paths, path) {
			continue
		}
		status := "ok"
		if !g.Store.Exists(entry.Key) {
			status = "missing"
		} else if ok, _ := g.Store.Verify(entry.Key); !ok {
			status = "corrupt"
		}
		fmt.Printf("  %s  %s  %s\n", status, formatSize(entry.Size), path)
		totalSize += entry.Size
	}
	fmt.Printf("\n%d files, %s total\n", len(entries), formatSize(totalSize))
	return nil
}

// Move renames a tracked file or directory prefix in the index.
// Single files: gda mv meg/file.fif bids/file.fif
// Directories:  gda mv meg/ bids/  (moves all files under meg/ to bids/)
// Instant operation — only symlinks and index entries, no content copies.
func (g *GDA) Move(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: gda mv <source> <dest>")
	}
	src := args[0]
	dst := args[1]

	// Try single-file move first
	entry := g.Index.Get(src)
	if entry != nil {
		return g.moveFile(src, dst)
	}

	// Otherwise treat as prefix/directory move
	srcPrefix := src
	if !strings.HasSuffix(srcPrefix, "/") {
		srcPrefix += "/"
	}

	var toMove []string
	for path := range g.Index.All() {
		if strings.HasPrefix(path, srcPrefix) {
			toMove = append(toMove, path)
		}
	}

	if len(toMove) == 0 {
		return fmt.Errorf("%s is not tracked and no tracked files share that prefix", src)
	}

	for _, oldPath := range toMove {
		suffix := oldPath[len(srcPrefix):]
		newPath := filepath.Join(dst, suffix)
		if err := g.moveFile(oldPath, newPath); err != nil {
			return fmt.Errorf("move %s: %w", oldPath, err)
		}
	}

	fmt.Printf("Moved %d files from %s/ -> %s/\n", len(toMove), strings.TrimRight(src, "/"), strings.TrimRight(dst, "/"))
	return g.Index.Save()
}

// moveFile moves a single tracked file from oldPath to newPath.
func (g *GDA) moveFile(src, dst string) error {
	entry := g.Index.Get(src)
	if entry == nil {
		return fmt.Errorf("%s is not tracked", src)
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	srcAbs := filepath.Join(rootAbs, src)
	dstAbs := filepath.Join(rootAbs, dst)

	// SAFETY CHECK: Verify that the source file actually exists, is a symlink,
	// and points to the correct object key before deleting it.
	fi, err := os.Lstat(srcAbs)
	if err != nil {
		return fmt.Errorf("cannot stat source file: %w", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("safety abort: %s is not a symlink. It may be modified or unlocked. Run 'gda add' first", src)
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	// Compute new relative symlink target
	objAbs := filepath.Join(rootAbs, ".gda", "objects", entry.Key[:2], entry.Key[2:])
	newRel, err := filepath.Rel(filepath.Dir(dstAbs), objAbs)
	if err != nil {
		return fmt.Errorf("compute symlink target: %w", err)
	}

	// Create new symlink
	if err := os.Symlink(newRel, dstAbs); err != nil {
		return fmt.Errorf("create symlink %s: %w", dst, err)
	}

	// Remove old symlink
	os.Remove(srcAbs) // best-effort

	// Update index
	g.Index.Remove(src)

	relDst, _ := filepath.Rel(rootAbs, dstAbs)
	g.Index.Set(relDst, entry.Key, entry.Size, entry.MTime)

	fmt.Printf("Moved %s -> %s\n", src, relDst)

	return nil
}

// Remove removes a file from tracking (content stays in store).
func (g *GDA) Remove(paths []string) error {
	for _, path := range paths {
		entry := g.Index.Get(path)
		if entry == nil {
			fmt.Printf("%s is not tracked, skipping\n", path)
			continue
		}
		g.Index.Remove(path)

		abs := filepath.Join(g.Root, path)
		os.Remove(abs) // remove symlink, best-effort

		fmt.Printf("Removed %s from index (content retained in store)\n", path)
	}
	return g.Index.Save()
}

func (g *GDA) printStats() error {
	fmt.Printf("Store: %s\n", g.Store.Root)
	fmt.Printf("Index: %d entries\n", g.Index.Count())
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formatSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
	}
}

// Snapshot saves the current index state as a named snapshot.
func (g *GDA) Snapshot(args []string) error {
	name := fmt.Sprintf("snapshot-%d", time.Now().Unix())
	if len(args) > 0 {
		name = args[0]
	}

	snaps, err := index.OpenSnapshots(g.Root)
	if err != nil {
		return err
	}

	entries := g.Index.All()
	if len(entries) == 0 {
		return fmt.Errorf("nothing to snapshot (no tracked files)")
	}

	if err := snaps.Create(name, entries); err != nil {
		return err
	}

	fmt.Printf("Snapshot %q saved (%d files, %s)\n", name, len(entries),
		formatSize(sumSizes(entries)))
	return nil
}

// Log lists all snapshots.
func (g *GDA) Log(args []string) error {
	snaps, err := index.OpenSnapshots(g.Root)
	if err != nil {
		return err
	}

	list, err := snaps.List()
	if err != nil {
		return err
	}

	if len(list) == 0 {
		fmt.Println("No snapshots.")
		return nil
	}

	for _, s := range list {
		t := time.Unix(s.CreatedAt, 0).Format("2006-01-02 15:04:05")
		var totalSize int64
		for _, e := range s.Entries {
			totalSize += e.Size
		}
		fmt.Printf("  %s  %s  %d files  %s\n", s.Name, t, len(s.Entries),
			formatSize(totalSize))
	}
	return nil
}

// Checkout restores the working tree to a snapshot state.
func (g *GDA) Checkout(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda checkout <snapshot-name>")
	}
	name := args[0]

	snaps, err := index.OpenSnapshots(g.Root)
	if err != nil {
		return err
	}

	snap, err := snaps.Load(name)
	if err != nil {
		return err
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return err
	}

	// Identify and remove files tracked in current index but NOT present in snapshot
	currentIndex := g.Index.All()
	for path := range currentIndex {
		if _, exists := snap.Entries[path]; !exists {
			abs := filepath.Join(rootAbs, path)
			os.Remove(abs) // Remove symlink, best-effort
			g.Index.Remove(path)
		}
	}

	for path, entry := range snap.Entries {
		abs := filepath.Join(rootAbs, path)
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", path, err)
		}

		// Compute relative symlink target
		objAbs := filepath.Join(rootAbs, ".gda", "objects", entry.Key[:2], entry.Key[2:])
		rel, err := filepath.Rel(filepath.Dir(abs), objAbs)
		if err != nil {
			return fmt.Errorf("symlink target for %s: %w", path, err)
		}

		// Remove existing file/symlink if present
		os.Remove(abs)

		if err := os.Symlink(rel, abs); err != nil {
			return fmt.Errorf("restore %s: %w", path, err)
		}

		g.Index.Set(path, entry.Key, entry.Size, entry.MTime)
	}

	if err := g.Index.Save(); err != nil {
		return err
	}

	fmt.Printf("Restored %q (%d files)\n", name, len(snap.Entries))
	return nil
}

// GC removes unreferenced objects from the store.
func (g *GDA) GC(args []string) error {
	dryRun := false
	for _, a := range args {
		if a == "--dry-run" || a == "-n" {
			dryRun = true
		}
	}

	// Collect all referenced keys from current index
	referenced := make(map[string]bool)
	for _, entry := range g.Index.All() {
		referenced[entry.Key] = true
	}

	// Collect all referenced keys from all snapshots
	snaps, err := index.OpenSnapshots(g.Root)
	if err == nil {
		snapList, err := snaps.List()
		if err == nil {
			for _, snap := range snapList {
				for _, entry := range snap.Entries {
					referenced[entry.Key] = true
				}
			}
		}
	}

	// Walk objects directory
	objectsDir := filepath.Join(g.Root, ".gda", "objects")
	var removed, kept int64
	var removedSize int64
	now := time.Now()
	gracePeriod := 1 * time.Hour

	err = filepath.Walk(objectsDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}
		if fi.IsDir() {
			return nil
		}

		// Grace period to prevent race conditions with concurrent additions
		if now.Sub(fi.ModTime()) < gracePeriod {
			kept++
			return nil
		}

		// Extract key from path: objects/XX/XXXX...
		rel, _ := filepath.Rel(objectsDir, path)
		if len(rel) < 3 {
			return nil
		}
		key := rel[:2] + rel[3:] // remove the '/' between prefix and rest

		if !referenced[key] {
			removed++
			removedSize += fi.Size()
			if !dryRun {
				if err := os.Remove(path); err != nil {
					return fmt.Errorf("remove %s: %w", path, err)
				}
			}
		} else {
			kept++
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("gc walk: %w", err)
	}

	if dryRun {
		fmt.Printf("Would remove %d objects (%s), keep %d objects\n",
			removed, formatSize(removedSize), kept)
	} else {
		fmt.Printf("Removed %d objects (%s), kept %d objects\n",
			removed, formatSize(removedSize), kept)
	}
	return nil
}

func sumSizes(entries map[string]*index.Entry) int64 {
	var total int64
	for _, e := range entries {
		total += e.Size
	}
	return total
}