---
title: Configuration
description: GDA configuration reference
---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GDA_DEV=1` | Enable debug logging to stderr with timestamps and operation durations |

## Remote Configuration

Remotes are stored in `.gdarc` at the repository root in YAML format:

```yaml
remotes:
  origin: rsync://server/path/to/store
```

Use `gda remote add` to manage remotes.

## No Other Configuration

GDA is designed to be zero-config. No config files, no gitignore equivalents. The `.gda/` directory is automatically excluded from operations.
