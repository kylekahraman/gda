---
title: Why GDA?
description: When GDA makes sense — and when it doesn't
---

Versioning large research datasets has no perfect solution today. Git can't handle multi-GB files. Git-annex wraps git to work around this, but brings significant complexity with adjusted branches, direct mode, special remotes, and git internals leaking into the user experience. DVC couples data versioning to ML pipelines and git. Rsync copies files but doesn't version them.

GDA takes a different approach: a standalone tool built specifically for versioning research data. No git dependency, no pipelines, no merge conflicts.

## Good fit

- You have large files (GB+) that don't belong in git
- You want to restore previous dataset states without merge conflicts
- You need rsync-based sync to a server or cold storage
- You want a tool where each command does one thing

## Bad fit

- You need two people editing the same dataset at the same time (not yet — planned)
- You need S3 or HTTP remotes right now (rsync only for now)
- You're on Windows (not supported yet)
- You need git integration — GDA is standalone by design

GDA is a focused tool. It versions your data and stays out of your way. If your workflow needs more than that, the features may arrive later. If they don't, other tools may be a better fit.
