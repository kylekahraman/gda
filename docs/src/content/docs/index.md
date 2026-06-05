---
title: GDA — Data Versioning for Research
description: Content-addressed data versioning for research datasets
---

import { Card, CardGrid } from '@astrojs/starlight/components';

GDA is a lightweight, standalone tool for versioning research data. Content-addressed storage, symlink working tree, snapshots, and remote sync — without the complexity of git-annex.

<CardGrid>
  <Card title="Quick Start" icon="rocket">
    Get started in 5 minutes — init, add, snapshot, done.
  </Card>
  <Card title="Installation" icon="download">
    Download a binary or build from source.
  </Card>
  <Card title="Commands" icon="reference">
    Full command reference for every GDA operation.
  </Card>
</CardGrid>

## Key Features

- **Content-addressed storage** — files stored once by SHA256 hash
- **Symlink working tree** — browse without special tools
- **Snapshots** — lightweight, named dataset checkpoints
- **Unlock/Lock** — edit files when pipelines need write access
- **Remote sync** — push/pull via rsync
- **No git dependency** — standalone binary
