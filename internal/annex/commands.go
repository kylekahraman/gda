package annex

import (
	"fmt"
	"os"
)

// Init initializes a GDA repo in the current directory.
func Init(args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// Check if already initialized
	if _, err := os.Stat(root + "/.gda"); err == nil {
		return fmt.Errorf("already initialized (remove .gda/ to re-init)")
	}

	g, err := Open(root)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Init()
}

// Add adds files to the GDA store.
func Add(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda add <file> [file...]")
	}

	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Add(args)
}

// Status shows the current state.
func Status(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Status(args)
}

// Move moves a tracked file.
func Move(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Move(args)
}

// Remove removes files from tracking.
func Remove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda rm <file> [file...]")
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Remove(args)
}

// Snapshot creates a named snapshot.
func Snapshot(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Snapshot(args)
}

// Log lists snapshots.
func Log(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Log(args)
}

// Checkout restores from a snapshot.
func Checkout(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Checkout(args)
}

// GC removes unreferenced objects.
func GC(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.GC(args)
}
