# Infralog

> **Warning**: This project is in early development. The API and features are subject to change.

Infralog monitors Terraform state files and emits resource-level events when changes are detected. It polls your state file at configurable intervals, detects additions, modifications, and deletions, and delivers change notifications to configured targets.

## Features

- Monitors Terraform state files stored in S3
- Detects resource and output changes with detailed diffs
- Configurable polling intervals
- Webhook notifications with JSON payloads
- Optional filtering by resource type and output name
- State persistence to detect changes across restarts

## Installation

### Building from Source

```bash
cd src/
go build -o infralog main.go
```

## Usage

Specify the configuration file using either the `--config-file` flag or the `INFRALOG_CONFIG_FILE` environment variable:

```bash
infralog --config-file config.yml
```

```bash
export INFRALOG_CONFIG_FILE=config.yml
infralog
```

## Configuration

Create a `config.yml` file with the following structure:

```yaml
polling:
  interval: 600  # Polling interval in seconds

tfstate:
  s3:
    bucket: "my-terraform-state-bucket"
    key: "path/to/terraform.tfstate"
    region: "us-east-1"

target:
  webhook:
    url: "https://example.com/infralog"
    method: "POST"  # POST or PUT (default: POST)

filter:
  # Optional: List of resource types to monitor.
  # Omit to monitor all resources, or use [] to monitor none.
  resource_types:
    - "aws_instance"
    - "aws_s3_bucket"

  # Optional: List of outputs to monitor.
  # Omit to monitor all outputs, or use [] to monitor none.
  outputs:
    - "instance_ip"

persistence:
  # Optional: Path to persist the last-seen state.
  # Enables change detection across restarts.
  state_file: "/var/lib/infralog/state.json"
```

## Targets

### Webhook

When changes are detected, Infralog sends an HTTP request to the configured URL with a JSON payload:

```json
{
  "diffs": {
    "resource_diffs": [
      {
        "resource_type": "aws_instance",
        "resource_name": "web_server",
        "status": "changed",
        "attribute_diffs": {
          "instance_type": {
            "old_value": "t2.small",
            "new_value": "t2.medium"
          }
        }
      }
    ],
    "output_diffs": [
      {
        "output_name": "instance_ip",
        "status": "changed",
        "value_diff": {
          "old_value": "10.0.1.20",
          "new_value": "10.0.1.28"
        }
      }
    ]
  },
  "metadata": {
    "tfstate": {
      "s3": {
        "bucket": "my-terraform-state-bucket",
        "key": "path/to/terraform.tfstate",
        "region": "us-east-1"
      }
    }
  }
}
```

The `status` field indicates the type of change: `added`, `changed`, or `removed`.

## Persistence

By default, Infralog stores the last-seen state in memory. This means that if Infralog restarts, any changes that occurred during downtime will not be detected.

To enable change detection across restarts, configure the `persistence.state_file` option:

- The last-seen state is saved to disk after each detected change
- On startup, Infralog loads the persisted state and compares it against the current state
- Changes that occurred while Infralog was stopped are detected and emitted

State files are written atomically to prevent corruption.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add my feature'`)
6. Push to the branch (`git push origin feature/my-feature`)
7. Open a pull request

Before submitting:

- Ensure all tests pass
- Add tests for new functionality
- Update documentation as needed

For major changes, please open an issue first to discuss your proposal.

## License

[Add your license here]
