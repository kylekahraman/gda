package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kylekahraman/gda/internal/annex"
	"github.com/kylekahraman/gda/internal/devlog"
)

var helpText = map[string]string{
	"init":     "Initialize a GDA repo in the current directory",
	"add":      "Track files or directories (recursive)",
	"status":   "Show tracked files and working-tree integrity",
	"mv":       "Rename a tracked path or directory prefix",
	"rm":       "Untrack a file (content stays in store)",
	"snapshot": "Save a named checkpoint of the current index",
	"log":      "List named snapshots",
	"checkout": "Restore working tree from a snapshot",
	"gc":       "Remove unreferenced blobs from the store",
	"fsck":     "Scan index vs working tree and repair broken symlinks",
	"unlock":   "Replace a symlink with a writable copy of the content",
	"lock":     "Re-hash a file and restore its symlink",
	"undo":     "Revert the last lock operation",
	"reindex":  "Rebuild index from existing symlinks",
	"remote":   "Add, remove, or list remote repositories",
	"push":     "Upload objects and snapshots to a remote",
	"pull":     "Download objects and snapshots from a remote",
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gda <command> [args...]\n\nCommands:\n")
	for _, name := range []string{
		"init", "add", "status", "mv", "rm",
		"snapshot", "log", "checkout",
		"gc", "fsck", "unlock", "lock", "undo",
		"reindex",
		"remote", "push", "pull",
	} {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, helpText[name])
	}
}

func runCmd(name string, args []string, fn func([]string) error) error {
	if devlog.Enabled() {
		devlog.Printf("CMD: gda %s %v", name, args)
	}
	start := time.Now()
	err := fn(args)
	elapsed := time.Since(start)
	if devlog.Enabled() {
		if err != nil {
			devlog.Printf("ERROR: gda %s: %v (%v)", name, err, elapsed)
		} else {
			devlog.Printf("OK: gda %s (%v)", name, elapsed)
		}
	}
	return err
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd == "help" || cmd == "--help" {
		if len(args) > 0 {
			if desc, ok := helpText[args[0]]; ok {
				fmt.Printf("%s: %s\n", args[0], desc)
				return
			}
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
			os.Exit(1)
		}
		usage()
		return
	}

	var err error
	switch cmd {
	case "init":
		err = runCmd("init", args, annex.Init)
	case "add":
		err = runCmd("add", args, annex.Add)
	case "status":
		err = runCmd("status", args, annex.Status)
	case "mv":
		err = runCmd("mv", args, annex.Move)
	case "rm":
		err = runCmd("rm", args, annex.Remove)
	case "snapshot":
		err = runCmd("snapshot", args, annex.Snapshot)
	case "log":
		err = runCmd("log", args, annex.Log)
	case "checkout":
		err = runCmd("checkout", args, annex.Checkout)
	case "gc":
		err = runCmd("gc", args, annex.GC)
	case "fsck":
		err = runCmd("fsck", args, annex.Fsck)
	case "unlock":
		err = runCmd("unlock", args, annex.Unlock)
	case "lock":
		err = runCmd("lock", args, annex.Lock)
	case "undo":
		err = runCmd("undo", args, annex.Undo)
	case "reindex":
		err = runCmd("reindex", args, annex.Reindex)
	case "remote":
		err = runCmd("remote", args, annex.Remote)
	case "push":
		err = runCmd("push", args, annex.Push)
	case "pull":
		err = runCmd("pull", args, annex.Pull)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
