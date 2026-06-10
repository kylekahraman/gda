package annex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Reindex walks the working tree, finds symlinks pointing to .gda/objects/,
// verifies the objects exist, and rebuilds the index from them.
func (g *GDA) Reindex(args []string) error {
	rootAbs, err := filepath.Abs(g.Root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}

	objectsDir := filepath.Join(rootAbs, ".gda", "objects")

	var found, skipped, added int

	err = filepath.Walk(rootAbs, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}

		relPath, _ := filepath.Rel(rootAbs, walkPath)

		// Skip .gda/ directory entirely
		if relPath == ".gda" || strings.HasPrefix(relPath, ".gda/") {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		// Only process symlinks
		if fi.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		found++

		// Resolve the symlink target
		target, err := os.Readlink(walkPath)
		if err != nil {
			fmt.Printf("  skip    %s (cannot readlink: %v)\n", relPath, err)
			skipped++
			return nil
		}

		// Check if target points to .gda/objects/
		// Target is relative (e.g., "../.gda/objects/ab/cdef1234")
		// Resolve it against the symlink's directory
		symlinkDir := filepath.Dir(walkPath)
		absTarget := filepath.Join(symlinkDir, target)
		absTarget, err = filepath.Abs(absTarget)
		if err != nil {
			skipped++
			return nil
		}

		if !strings.HasPrefix(absTarget, objectsDir) {
			// Not a GDA object symlink — could be a git-annex remnant or other
			skipped++
			return nil
		}

		// Extract key from the object path: <objectsDir>/AB/CDEF... → ABCDEF...
		relObj, _ := filepath.Rel(objectsDir, absTarget)
		parts := strings.SplitN(relObj, string(filepath.Separator), 2)
		if len(parts) != 2 {
			skipped++
			return nil
		}
		key := parts[0] + parts[1]

		// Verify the object file exists and is a regular file
		objFi, err := os.Stat(absTarget)
		if err != nil {
			fmt.Printf("  skip    %s (object missing: %s)\n", relPath, key[:16])
			skipped++
			return nil
		}
		if objFi.IsDir() {
			skipped++
			return nil
		}

		// Also check if the object content matches the symlink path
		size := objFi.Size()
		mtime := objFi.ModTime().Unix()

		// Check if already indexed (skip)
		if g.Index.Get(relPath) != nil {
			fmt.Printf("  ok      %s\n", relPath)
			return nil
		}

		// Add to index
		g.Index.Set(relPath, key, size, mtime)
		added++
		fmt.Printf("  added   %s (%s, SHA256: %s)\n", relPath, formatSize(size), key[:16])

		return nil
	})

	if err != nil {
		return fmt.Errorf("reindex walk: %w", err)
	}

	if err := g.Index.Save(); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	fmt.Printf("\nFound %d symlinks, skipped %d, added %d to index\n", found, skipped, added)
	if added == 0 && found > 0 {
		fmt.Println("All symlinks already in index or not GDA objects.")
	}
	return nil
}
