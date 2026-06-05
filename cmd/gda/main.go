package main

import (
	"fmt"
	"os"

	"github.com/kylekahraman/gda/internal/annex"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: gda <command> [args...]\n")
		fmt.Fprintf(os.Stderr, "Commands: init, add, status, mv, rm\n")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

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
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
