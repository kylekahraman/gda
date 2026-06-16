package annex

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kylekahraman/gda/internal/devlog"
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
	fmt.Printf("Initialized GDA store in %s\n", filepath.Join(g.Root, ".gda"))

	// Create default .gdaignore
	//? should this be moved somewhere else?
	ignoreFile := filepath.Join(g.Root, ".gdaignore")
	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		defaultIgnore := `# GDA ignore patterns (same syntax as .gitignore)
		# Files matching these patterns are skipped by 'gda add'.
		.DS_Store
		*.swp
		*.swo
		*~
		Thumbs.db
		ehthumbs.db
		Desktop.ini
		.gda/
		`
		if err := os.WriteFile(ignoreFile, []byte(defaultIgnore), 0644); err != nil {
			return fmt.Errorf("create .gdaignore: %w", err)
		}
		fmt.Println("Created .gdaignore")
	}

	return g.printStats()
}

// loadIgnorePatterns reads .gdaignore from the repo root.
// Returns nil if no .gdaignore exists.
func (g *GDA) loadIgnorePatterns() ([]string, error) {
	ignoreFile := filepath.Join(g.Root, ".gdaignore")
	f, err := os.Open(ignoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// isIgnored checks if a path (relative to repo root) matches any .gdaignore pattern.
func (g *GDA) isIgnored(relPath string, patterns []string) bool {
	base := filepath.Base(relPath)
	// Normalize: ensure paths don't start with ./
	clean := strings.TrimPrefix(relPath, "./")
	for _, p := range patterns {
		// Strip trailing / for matching (they're directory-only hints)
		p = strings.TrimSuffix(p, "/")
		// Try matching basename
		if matched, _ := filepath.Match(p, base); matched {
			return true
		}
		// Try matching clean relative path
		if matched, _ := filepath.Match(p, clean); matched {
			return true
		}
		// Try matching any path component (like .gitignore's "foo" matches dir/foo)
		if !strings.Contains(p, "/") {
			if matched, _ := filepath.Match(p, clean); matched {
				return true
			}
		}
	}
	return false
}

// Add adds files or directories to the store.
func (g *GDA) Add(paths []string) error {
	// Load ignore patterns once
	ignorePatterns, err := g.loadIgnorePatterns()
	if err != nil {
		return fmt.Errorf("load .gdaignore: %w", err)
	}

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

				// Compute relative path from repo root once
				relPath, _ := filepath.Rel(g.Root, walkPath)

				// Skip .gda/ directory and contents entirely
				if relPath == ".gda" || strings.HasPrefix(relPath, ".gda/") {
					if walkFi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				if walkFi.IsDir() || walkFi.Mode()&os.ModeSymlink != 0 {
					return nil
				}

				// Skip files matching .gdaignore patterns
				if len(ignorePatterns) > 0 && g.isIgnored(relPath, ignorePatterns) {
					devlog.Printf("skipping (ignored): %s", relPath)
					return nil
				}

				files = append(files, walkPath)
				return nil
			})
			if err != nil {
				return fmt.Errorf("walk %s: %w", path, err)
			}
		} else {
			absPath, _ := filepath.Abs(path)
			relPath, _ := filepath.Rel(g.Root, absPath)
			if len(ignorePatterns) > 0 && g.isIgnored(relPath, ignorePatterns) {
				fmt.Printf("Skipped %s (matches .gdaignore)\n", relPath)
				continue
			}
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

	total := len(files)
	for i, abs := range files {
		prefix := fmt.Sprintf("[%d/%d] ", i+1, total)
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

		fmt.Printf("%sAdded %s (%s, SHA256: %s)\n", prefix, relPath, formatSize(fi.Size()), key[:16])
	}

	return g.Index.Save()
}

// Status shows the current state of tracked files.
// Checks both the working tree (missing/broken symlinks) and the store.
func (g *GDA) Status(paths []string) error {
	entries := g.Index.All()

	if len(entries) == 0 {
		fmt.Println("No files tracked.")
		return nil
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	var totalSize int64
	var failCount int
	fmt.Println("Tracked files:")
	for path, entry := range entries {
		if len(paths) > 0 && !contains(paths, path) {
			continue
		}

		status := g.fileStatus(rootAbs, path, entry.Key, entry)
		if status != "ok" && status != "unlocked" {
			failCount++
		}
		fmt.Printf("  %s  %s  %s\n", status, formatSize(entry.Size), path)
		totalSize += entry.Size
	}

	fmt.Printf("\n%d files, %s total", len(entries), formatSize(totalSize))
	if failCount > 0 {
		fmt.Printf(", %d with issues", failCount)
	}
	fmt.Println()
	return nil
}

// fileStatus checks the working-tree and store health of a tracked file.
func (g *GDA) fileStatus(rootAbs, path, key string, entry *index.Entry) string {
	abs := filepath.Join(rootAbs, path)

	// Check working tree
	fi, err := os.Lstat(abs)
	if os.IsNotExist(err) {
		return "missing"
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		// Regular file (or other non-symlink)
		if entry != nil && entry.Unlocked {
			// Check if content matches stored key
			if ok, _ := g.Store.VerifyAtPath(key, abs); ok {
				return "unlocked"
			}
			return "modified"
		}
		return "modified"
	}

	// Symlink exists — does it resolve?
	if _, err := os.Stat(abs); err != nil {
		return "broken"
	}

	// Symlink resolves — check store
	if !g.Store.Exists(key) {
		return "missing"
	}
	if ok, _ := g.Store.Verify(key); !ok {
		return "corrupt"
	}
	return "ok"
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
	for path, entry := range currentIndex {
		if _, exists := snap.Entries[path]; !exists {
			abs := filepath.Join(rootAbs, path)
			if entry.Unlocked {
				if ok, _ := g.Store.VerifyAtPath(entry.Key, abs); !ok {
					fmt.Printf("warning: file is unlocked with modifications, skipping: %s\n", path)
					continue
				}
			}
			os.Remove(abs) // Remove symlink/file, best-effort
			g.Index.Remove(path)
		}
	}

	for path, entry := range snap.Entries {
		abs := filepath.Join(rootAbs, path)

		// Check if current file is unlocked and has modifications
		currentEntry := g.Index.Get(path)
		if currentEntry != nil && currentEntry.Unlocked {
			if ok, _ := g.Store.VerifyAtPath(currentEntry.Key, abs); !ok {
				fmt.Printf("warning: file is unlocked with modifications, skipping: %s\n", path)
				continue
			}
		}

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

// Fsck scans the index against the working tree and repairs what it can.
func (g *GDA) Fsck(args []string) error {
	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	entries := g.Index.All()
	if len(entries) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	var okCount, brokenCount, fixedCount, modifiedCount, missingCount int

	for path, entry := range entries {
		abs := filepath.Join(rootAbs, path)
		status := g.fileStatus(rootAbs, path, entry.Key, entry)

		switch status {
		case "ok", "unlocked":
			okCount++
		case "modified":
			modifiedCount++
			fmt.Printf("  modified %s\n", path)
		case "corrupt":
			brokenCount++
			fmt.Printf("  corrupt  %s (object content mismatch)\n", path)
		case "missing":
			if g.Store.Exists(entry.Key) {
				repaired := false
				if entry.Unlocked {
					os.Remove(abs)
					repaired = g.Store.CopyTo(entry.Key, abs) == nil
				} else {
					repaired = g.repairSymlink(abs, entry.Key)
				}
				if repaired {
					fixedCount++
					fmt.Printf("  fixed    %s\n", path)
				} else {
					missingCount++
					fmt.Printf("  missing  %s (object exists but repair failed)\n", path)
				}
			} else {
				missingCount++
				fmt.Printf("  missing  %s (object also missing)\n", path)
			}
		case "broken":
			if !g.Store.Exists(entry.Key) {
				brokenCount++
				fmt.Printf("  broken   %s (object missing)\n", path)
			} else {
				repaired := false
				if entry.Unlocked {
					os.Remove(abs)
					repaired = g.Store.CopyTo(entry.Key, abs) == nil
				} else {
					repaired = g.repairSymlink(abs, entry.Key)
				}
				if repaired {
					fixedCount++
					fmt.Printf("  fixed    %s\n", path)
				} else {
					brokenCount++
					fmt.Printf("  broken   %s (repair failed)\n", path)
				}
			}
		}
	}

	fmt.Printf("\n%d ok, %d broken, %d fixed, %d modified, %d missing\n",
		okCount, brokenCount, fixedCount, modifiedCount, missingCount)
	return nil
}

// repairSymlink recreates a symlink at abs pointing to the stored object for key.
// Returns true on success, false on failure. Does not touch the index.
func (g *GDA) repairSymlink(abs string, key string) bool {
	objAbs, err := filepath.Abs(g.Store.ObjectPath(key))
	if err != nil {
		return false
	}
	newRel, err := filepath.Rel(filepath.Dir(abs), objAbs)
	if err != nil {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return false
	}
	os.Remove(abs)
	if err := os.Symlink(newRel, abs); err != nil {
		return false
	}
	return true
}

func sumSizes(entries map[string]*index.Entry) int64 {
	var total int64
	for _, e := range entries {
		total += e.Size
	}
	return total
}

// Unlock replaces symlinks with writable copies of the content.
func (g *GDA) Unlock(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no files specified to unlock")
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	targets, err := g.expandPaths(paths, rootAbs)
	if err != nil {
		return err
	}

	for _, relPath := range targets {
		entry := g.Index.Get(relPath)
		if entry == nil {
			return fmt.Errorf("%s is not tracked", relPath)
		}

		if entry.Unlocked {
			fmt.Printf("%s is already unlocked\n", relPath)
			continue
		}

		absPath := filepath.Join(rootAbs, relPath)

		fi, err := os.Lstat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", relPath)
			}
			return fmt.Errorf("stat %s: %w", relPath, err)
		}

		if fi.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("warning: %s is modified, not a symlink, skipping\n", relPath)
			continue
		}

		if !g.Store.Exists(entry.Key) {
			return fmt.Errorf("object not found for %s, run gda add first", relPath)
		}

		objPath := g.Store.ObjectPath(entry.Key)
		content, err := os.ReadFile(objPath)
		if err != nil {
			return fmt.Errorf("read object %s: %w", entry.Key, err)
		}

		if err := os.Remove(absPath); err != nil {
			return fmt.Errorf("remove symlink %s: %w", relPath, err)
		}

		if err := os.WriteFile(absPath, content, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", relPath, err)
		}

		entry.Unlocked = true
		g.Index.Put(entry)

		fmt.Printf("Unlocked %s\n", relPath)
	}

	return g.Index.Save()
}

// Lock re-hashes the file, stores it if changed, and replaces it with a symlink.
func (g *GDA) Lock(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no files specified to lock")
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	targets, err := g.expandPaths(paths, rootAbs)
	if err != nil {
		return err
	}

	for _, relPath := range targets {
		absPath := filepath.Join(rootAbs, relPath)

		entry := g.Index.Get(relPath)
		if entry == nil {
			return fmt.Errorf("%s is not tracked", relPath)
		}

		if !entry.Unlocked {
			fmt.Printf("%s is already locked\n", relPath)
			continue
		}

		fi, err := os.Lstat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", relPath)
			}
			return fmt.Errorf("stat %s: %w", relPath, err)
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			fmt.Printf("warning: %s is a symlink, already locked, skipping\n", relPath)
			continue
		}

		// Save undo state before modifying
		undo, err := loadUndoStore(g.Root)
		if err != nil {
			return fmt.Errorf("load undo: %w", err)
		}
		undo.record(relPath, entry.Key, entry.Size, entry.MTime)
		if err := undo.save(); err != nil {
			return fmt.Errorf("save undo: %w", err)
		}

		newKey, err := g.Store.Add(absPath)
		if err != nil {
			return fmt.Errorf("lock store add %s: %w", relPath, err)
		}

		fiFinal, err := os.Lstat(absPath)
		if err != nil {
			return fmt.Errorf("stat final %s: %w", relPath, err)
		}

		entry.Key = newKey
		entry.Size = fiFinal.Size()
		entry.MTime = fiFinal.ModTime().Unix()
		entry.Unlocked = false
		g.Index.Put(entry)

		objAbs := filepath.Join(rootAbs, ".gda", "objects", newKey[:2], newKey[2:])
		rel, err := filepath.Rel(filepath.Dir(absPath), objAbs)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}

		if err := os.Remove(absPath); err != nil {
			return fmt.Errorf("remove original %s: %w", relPath, err)
		}
		if err := os.Symlink(rel, absPath); err != nil {
			return fmt.Errorf("create symlink %s: %w", relPath, err)
		}

		fmt.Printf("Locked %s\n", relPath)
	}

	return g.Index.Save()
}

// expandPaths resolves each path argument. Files are returned as-is.
// Directories are expanded to all tracked entries under that prefix.
func (g *GDA) expandPaths(paths []string, rootAbs string) ([]string, error) {
	var targets []string
	seen := make(map[string]bool)

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("abs %s: %w", path, err)
		}

		relPath, err := filepath.Rel(rootAbs, absPath)
		if err != nil {
			return nil, fmt.Errorf("rel %s: %w", path, err)
		}
		if strings.HasPrefix(relPath, "..") {
			return nil, fmt.Errorf("%s is outside the repo", path)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", relPath, err)
		}

		if !fi.IsDir() {
			if !seen[relPath] {
				targets = append(targets, relPath)
				seen[relPath] = true
			}
			continue
		}

		// Directory — collect all tracked entries under this prefix
		for _, entry := range g.Index.All() {
			if relPath == "." || entry.Path == relPath || strings.HasPrefix(entry.Path, relPath+"/") {
				if !seen[entry.Path] {
					targets = append(targets, entry.Path)
					seen[entry.Path] = true
				}
			}
		}
	}

	return targets, nil
}

// Undo reverts the last lock operation for the given files.
func (g *GDA) Undo(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("usage: gda undo <file> [file...]")
	}

	undo, err := loadUndoStore(g.Root)
	if err != nil {
		return fmt.Errorf("load undo: %w", err)
	}
	if undo.isEmpty() {
		return fmt.Errorf("nothing to undo")
	}

	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	targets, err := g.expandPaths(paths, rootAbs)
	if err != nil {
		return err
	}

	var restored int
	for _, relPath := range targets {
		absPath := filepath.Join(rootAbs, relPath)

		entry := undo.pop(relPath)
		if entry == nil {
			continue
		}

		if !g.Store.Exists(entry.Key) {
			return fmt.Errorf("previous object for %s not found in store (may have been GC'd)", relPath)
		}

		// Restore the symlink pointing to the old object
		objAbs := filepath.Join(rootAbs, ".gda", "objects", entry.Key[:2], entry.Key[2:])
		rel, err := filepath.Rel(filepath.Dir(absPath), objAbs)
		if err != nil {
			return fmt.Errorf("symlink target for %s: %w", relPath, err)
		}

		if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", relPath, err)
		}
		if err := os.Symlink(rel, absPath); err != nil {
			return fmt.Errorf("create symlink %s: %w", relPath, err)
		}

		// Update index with old key
		g.Index.Set(relPath, entry.Key, entry.Size, entry.MTime)

		fmt.Printf("Undid %s\n", relPath)
		restored++
	}

	if err := g.Index.Save(); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	// Save undo store (entries were popped)
	if undo.isEmpty() {
		if err := undo.removeFile(); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove undo: %w", err)
		}
	} else {
		if err := undo.save(); err != nil {
			return fmt.Errorf("save undo: %w", err)
		}
	}

	if restored == 0 {
		return fmt.Errorf("no files were undone")
	}
	return nil
}

func (g *GDA) RemoteAdd(name, url string) error {
	return g.Index.RemoteAdd(name, url)
}

func (g *GDA) Push(remoteName string) error {
	if remoteName == "" {
		remoteName = "origin"
	}
	remote, err := g.Index.RemoteGet(remoteName)
	if err != nil {
		return err
	}

	src := filepath.Join(g.Root, ".gda") + "/"
	var dest string
	if strings.Contains(remote.URL, ":") {
		dest = remote.URL + "/.gda/"
	} else {
		dest = filepath.Join(remote.URL, ".gda") + "/"
	}

	fmt.Printf("Pushing to remote %s (%s)...\n", remote.Name, remote.URL)
	cmd := exec.Command("rsync", "-a", "--info=progress2",
		"--include=objects/", "--include=objects/**",
		"--include=snapshots/", "--include=snapshots/**",
		"--exclude=*", src, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync push: %w", err)
	}

	fmt.Println("Push complete.")
	return nil
}

func (g *GDA) Pull(remoteName string) error {
	if remoteName == "" {
		remoteName = "origin"
	}
	remote, err := g.Index.RemoteGet(remoteName)
	if err != nil {
		return err
	}

	var src string
	if strings.Contains(remote.URL, ":") {
		src = remote.URL + "/.gda/"
	} else {
		src = filepath.Join(remote.URL, ".gda") + "/"
	}
	dest := filepath.Join(g.Root, ".gda") + "/"

	fmt.Printf("Pulling from remote %s (%s)...\n", remote.Name, remote.URL)
	cmd := exec.Command("rsync", "-a", "--info=progress2",
		"--include=objects/", "--include=objects/**",
		"--include=snapshots/", "--include=snapshots/**",
		"--exclude=*", src, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync pull: %w", err)
	}

	fmt.Println("Pull complete. Verifying integrity...")
	return g.Status(nil)
}