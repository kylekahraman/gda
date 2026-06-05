---
title: GDA — Data Versioning for Research
description: Version control for research datasets — one command, no git
---

> ⚠️ Pre-alpha (v0.0.1). Designed for individual researchers who want to version, restore, and share datasets. Sharing via push/pull works today. Simultaneous editing is not supported.

GDA versions research datasets. No git required. You download data, add it,
snapshot when you've done something useful, checkout when you mess up, and
push to share with collaborators.

## Who this is for

- **You download datasets from OpenNeuro** and want to track preprocessing steps
- **You share data with labmates** and want everyone on the same file tree
- **You archive data** and want to know it's not silently corrupting
- **You restructure files constantly** and hate waiting for copies
- **You want to version large datasets without using git-annex**

→ [Why GDA?](/gda/why-gda/) — honest comparison
→ [Use Cases](/gda/use-cases/) — real workflows
→ [Installation](/gda/installation/) — get it now
→ [Quick Start](/gda/quick-start/) — 5 minute tutorial

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

→ [Continue to Installation →](/gda/installation/)

---

## Documentation

| Section | What's in it |
|---|---|
| [Overview](/gda/overview/) | What GDA does and why it's different |
| [Installation](/gda/installation/) | Download or build GDA |
| [Quick Start](/gda/quick-start/) | Track your first dataset in 5 minutes |
| [Concepts](/gda/concepts/) | How hashing, symlinks, and snapshots work |
| [Guides](/gda/guides/adding-data/) | Adding, snapshotting, syncing data |
| [FAQ](/gda/faq/) | Common questions and troubleshooting |
| [Commands](/gda/reference/commands/) | Full reference |
