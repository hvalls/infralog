# Infralog

> **Warning**: This project is in early development. The API and features are subject to change.

Infralog monitors Terraform state files and emits resource-level events when changes are detected. It polls your state file at configurable intervals, detects additions, modifications, and deletions, and delivers change notifications to configured targets.

## Features

- Monitors Terraform state files from multiple backends (S3, local filesystem)
- Detects resource and output changes with detailed diffs
- Webhook and Slack notifications
- Optional filtering by resource type and output name
- State persistence to detect changes across restarts
- Prometheus metrics endpoint

## Documentation

Full documentation is available at **[hvalls.github.io/infralog](https://hvalls.github.io/infralog/)**

## Quick Start

```bash
# Build
cd src/
go build -o infralog main.go

# Run
./infralog --config-file config.yml
```

## License

MIT License - see [LICENSE](LICENSE) for details.
