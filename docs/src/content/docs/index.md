---
title: GDA — Data Versioning for Research
description: Version control for research datasets — without git-annex
---

GDA versions research datasets. No git required. You download data, add it,
snapshot when you've done something useful, checkout when you mess up, and
push to share with collaborators.

---

## Getting Started

| Guide | What you'll learn |
|---|---|
| [Overview](/gda/overview/) | What GDA is and when it's useful |
| [Installation](/gda/installation/) | Download or build GDA |
| [Quick Start](/gda/quick-start/) | Track your first dataset in 5 minutes |

## Guides

| Guide | What you'll learn |
|---|---|
| [Adding Data](/gda/guides/adding-data/) | Track files and directories |
| [Snapshots & Checkout](/gda/guides/snapshots/) | Version your dataset state |
| [Moving & Removing](/gda/guides/moving-removing/) | Rename or untrack files |
| [Unlock & Lock](/gda/guides/unlock-lock/) | Edit tracked files safely |
| [Remote Storage](/gda/guides/remote-storage/) | Sync via rsync |
| [Housekeeping](/gda/guides/housekeeping/) | GC and integrity checks |

## Reference

| Page | Description |
|---|---|
| [Commands](/gda/reference/commands/) | Complete command reference |
| [Store Format](/gda/reference/store-format/) | How data is stored on disk |
| [Configuration](/gda/reference/configuration/) | Environment variables and remotes |

---

## Quick Start

```bash
gda init
gda add bids/
gda status
gda mv bids/ sourcedata/
gda snapshot experiment1
gda checkout experiment1
```

→ [Continue to Overview →](/gda/overview/)
