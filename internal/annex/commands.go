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
	if _, err := os.Stat(root + "/.gda/index.db"); err == nil {
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

// Fsck scans and repairs the working tree.
func Fsck(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Fsck(args)
}

// Unlock unlocks the given tracked files.
func Unlock(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda unlock <file> [file...]")
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Unlock(args)
}

// Lock locks the given tracked files.
func Lock(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda lock <file> [file...]")
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Lock(args)
}

// Undo reverts the last lock for the given files.
func Undo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gda undo <file> [file...]")
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Undo(args)
}

// Reindex rebuilds the index from existing symlinks in the working tree.
func Reindex(args []string) error {
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Reindex(args)
}

func Remote(args []string) error {
	if len(args) == 0 {
		g, err := Open(".")
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer g.Close()
		remotes, err := g.Index.RemoteList()
		if err != nil {
			return err
		}
		for _, r := range remotes {
			fmt.Printf("%s\t%s\n", r.Name, r.URL)
		}
		return nil
	}

	subCmd := args[0]
	switch subCmd {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: gda remote add <name> <url>")
		}
		g, err := Open(".")
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer g.Close()
		err = g.RemoteAdd(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Printf("Added remote %s (%s)\n", args[1], args[2])
		return nil
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: gda remote remove <name>")
		}
		g, err := Open(".")
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer g.Close()
		return g.Index.RemoteRemove(args[1])
	case "list":
		g, err := Open(".")
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer g.Close()
		remotes, err := g.Index.RemoteList()
		if err != nil {
			return err
		}
		for _, r := range remotes {
			fmt.Printf("%s\t%s\n", r.Name, r.URL)
		}
		return nil
	default:
		return fmt.Errorf("unknown remote command: %s", subCmd)
	}
}

func Push(args []string) error {
	remoteName := ""
	if len(args) > 0 {
		remoteName = args[0]
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Push(remoteName)
}

func Pull(args []string) error {
	remoteName := ""
	if len(args) > 0 {
		remoteName = args[0]
	}
	g, err := Open(".")
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer g.Close()
	return g.Pull(remoteName)
}
