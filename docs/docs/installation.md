---
sidebar_position: 2
---

# Installation

## Docker

```bash
docker pull ghcr.io/hvalls/infralog:latest
```

Or build locally:

```bash
docker build -t infralog:latest .
```

## Building from Source

```bash
cd src/
go build -o infralog main.go
```
