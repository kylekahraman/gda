package main

import (
	"fmt"
	"os"

	"github.com/kylekahraman/gda/internal/annex"
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
	"remote":   "Add, remove, or list remote repositories",
	"push":     "Upload objects and snapshots to a remote",
	"pull":     "Download objects and snapshots from a remote",
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gda <command> [args...]\n\nCommands:\n")
	for _, name := range []string{
		"init", "add", "status", "mv", "rm",
		"snapshot", "log", "checkout",
		"gc", "fsck", "unlock", "lock",
		"remote", "push", "pull",
	} {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, helpText[name])
	}
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
		err = annex.Init(args)
	case "add":
		err = annex.Add(args)
	case "status":
		err = annex.Status(args)
	case "mv":
		err = annex.Move(args)
	case "rm":
		err = annex.Remove(args)
	case "snapshot":
		err = annex.Snapshot(args)
	case "log":
		err = annex.Log(args)
	case "checkout":
		err = annex.Checkout(args)
	case "gc":
		err = annex.GC(args)
	case "fsck":
		err = annex.Fsck(args)
	case "unlock":
		err = annex.Unlock(args)
	case "lock":
		err = annex.Lock(args)
	case "remote":
		err = annex.Remote(args)
	case "push":
		err = annex.Push(args)
	case "pull":
		err = annex.Pull(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
