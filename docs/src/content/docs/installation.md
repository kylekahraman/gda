---
title: Installation
description: How to install GDA
---

## Download a Binary

Download the latest release from [GitHub](https://github.com/kylekahraman/gda/releases).

| Platform | Binary |
|----------|--------|
| Linux (x86_64) | `gda-linux-amd64` |
| macOS (Apple Silicon) | `gda-darwin-arm64` |
| macOS (Intel) | `gda-darwin-amd64` |

```shell
# Download and make executable
chmod +x gda-darwin-arm64

# Move to your PATH
mv gda-darwin-arm64 /usr/local/bin/gda
```

## Build from Source

Requires Go 1.21+.

```shell
git clone https://github.com/kylekahraman/gda.git
cd gda
go build -o gda ./cmd/gda/
./gda help
```

## Verify Installation

```shell
gda help
gda version
```
