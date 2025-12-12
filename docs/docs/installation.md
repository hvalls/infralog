---
sidebar_position: 2
---

# Installation

## Docker

```bash
docker pull hvalls/infralog:latest
```

## Building from source

```bash
git clone https://github.com/hvalls/infralog
cd src/
go build -o infralog main.go
```

or 

```bash
docker build -t infralog:latest .
```