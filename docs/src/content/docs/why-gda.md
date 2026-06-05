---
title: Why GDA?
description: The honest truth about when GDA makes sense
---

You probably got here because git-annex made you want to throw your laptop off a roof. I get it.

It's powerful but it's a beast. Adjusted branches, direct mode, special remotes, git internals leaking everywhere. You don't need any of that. You just need to version some files and get on with your research.

## Good fit

- You have large files (GB+) that don't belong in git
- You want to restore previous dataset states without merge conflicts
- You need rsync-based sync to a server or cold storage
- You value "one command = one thing" over reading the manual for the 15th time

## Bad fit

- You need two people editing the same dataset at the same time (not yet — planned)
- You need S3/HTTP remotes right now (rsync only for now)
- You're on Windows (not supported yet)
- You need git integration — GDA is standalone by design

GDA is a focused tool. It versions your data and tries not to get in your way. If your workflow is more complex than that, fair enough. Come back when the features land.
