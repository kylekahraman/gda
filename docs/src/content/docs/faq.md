---
title: FAQ & Troubleshooting
description: Common questions and solutions
---

## General

### How is GDA different from git-annex?

Git-annex is powerful but complex — it hooks into git internals, uses multiple branches (master, synced/master, adjusted branches), and has a steep learning curve. GDA is a standalone tool with no git dependency. No adjusted branches, no merge conflicts, no special index. One working tree, one command per operation.

### How is GDA different from DVC?

DVC (Data Version Control) is designed for ML pipelines — it tracks data alongside code in a git repo. GDA is designed for research datasets directly. No git integration required. No pipelines. Just data versioning.

### How is GDA different from Git LFS?

Git LFS replaces large files with pointer files in git, storing blobs on a remote server. It still requires git and has limitations (file size caps, remote requirements). GDA works entirely standalone — no git, no server required for local use.

### Does GDA use git?

No. GDA manages its own object store, index, and snapshots. There is no git dependency.

### Can I use GDA alongside git?

Yes. GDA stores everything in `.gda/`. You can have a git repo in the same directory — they won't interfere. Add `.gda/` to your `.gitignore`.

## Data

### Where is my data stored?

In `.gda/objects/XX/YYY...` where `XX` is the first 2 characters of the SHA256 hash and `YYY...` is the rest. Files are read-only (0444 permissions) and content-addressed — the filename IS the hash.

### Is data duplicated if I add the same file twice?

No. Identical content produces the same SHA256 hash. The second add finds the existing object and only creates an index entry. Zero storage overhead for duplicates.

### Can I recover deleted files?

If the file is still referenced by a snapshot, use `gda checkout <snapshot>` to restore the working tree. If you removed the file from tracking (`gda rm`) and ran `gda gc`, the object may be deleted. Run `gda gc --dry-run` first to see what would be removed.

## Snapshots

### What's the difference between a snapshot and a commit?

A git commit records changes (diff) from the previous commit. A GDA snapshot records the complete state of every tracked file. This makes checkout instant and avoids merge conflicts, but snapshots are slightly larger (though only hashes — the actual file content is stored once).

### How many snapshots can I have?

As many as you want. Each snapshot is a JSON manifest of path→hash mappings. A snapshot with 10,000 files is roughly 1-2 MB.

## Performance

### How fast is GDA?

Hashing is the bottleneck — GDA reads every file once (hash + copy in a single pass). On a modern SSD, expect ~500 MB/s for adds. Operations like `mv`, `rm`, `status` are instant (index-only).

### How large can a repository be?

The store can hold any number of objects limited only by disk space. The BoltDB index handles millions of entries without issue.

## Errors

### "safety abort: file is not a symlink"

You're trying to move a file that was unlocked or modified. Run `gda lock <file>` first to restore the symlink, then retry the move.

### "no files found to add"

GDA skips symlinks and directories automatically. If all files in the path are symlinks or the directory is empty, nothing is added.

### Object store integrity

Run `gda fsck` to scan all tracked files and verify their hashes match the stored objects. Broken symlinks are repaired automatically.

## Getting help

- [GitHub Issues](https://github.com/kylekahraman/gda/issues) — bug reports and feature requests
- [GitHub Discussions](https://github.com/kylekahraman/gda/discussions) — questions and community support
