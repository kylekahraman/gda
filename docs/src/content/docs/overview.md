---
title: GDA — Data Versioning for Research
description: Content-addressed data versioning for research datasets
---

GDA is a lightweight, standalone tool for versioning research data. It uses content-addressed storage — every file is hashed (SHA256), stored in an object store, and referenced by its hash. The working tree keeps symlinks to objects, making datasets navigable without hydrating the full store.

## Why GDA?

Research data pipelines suffer from a fundamental problem: **how do you version large, binary datasets without duplicating them on every change?**

Git can't handle large files. Git-annex can, but its complexity is overwhelming for most researchers. GDA fills the gap — simpler than git-annex, purpose-built for research data.

## Key Features

- **Content-addressed storage** — files stored once, referenced by SHA256 hash
- **Symlink working tree** — browse and access files without special tools
- **Snapshots** — lightweight, named checkpoints of your dataset state
- **Unlock/Lock** — edit files safely when pipelines need write access
- **Remote sync** — push/pull to any rsync target (server, cloud, cold storage)
- **No git dependency** — GDA manages everything itself

## When to Use GDA

- Versioning fMRI, MEG, iEEG, or behavioral datasets
- Sharing large research data across lab members
- Archiving raw and processed data with change tracking
- Lightweight alternative to git-annex for neuroscience data
