# Infralog

> **Warning**: This project is in early development. The API and features are subject to change.

Infralog monitors Terraform state files and emits resource-level events when changes are detected. It polls your state file at configurable intervals, detects additions, modifications, and deletions, and delivers change notifications to configured targets.

## Features

- Monitors Terraform state files from multiple backends (S3, local filesystem)
- Detects resource and output changes with detailed diffs
- Configurable polling intervals
- Webhook notifications with JSON payloads and automatic retries
- Slack notifications with formatted messages
- Stdout logging with text or JSON format (default when no targets configured)
- Optional filtering by resource type and output name
- State persistence to detect changes across restarts

## Installation

### Docker

```bash
docker pull ghcr.io/hvalls/infralog:latest
```

Or build locally:

```bash
docker build -t infralog:latest .
```

### Building from Source

```bash
cd src/
go build -o infralog main.go
```

## Usage

### Running with Docker

```bash
docker run -v /path/to/config.yml:/etc/infralog/config.yml:ro \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  infralog:latest
```

For local backend with Docker:

```bash
docker run -v /path/to/config.yml:/etc/infralog/config.yml:ro \
  -v /path/to/terraform.tfstate:/data/terraform.tfstate:ro \
  infralog:latest
```

Using docker-compose:

```bash
docker-compose up
```

### Running from Binary

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
  # Configure one backend source (S3 or local)
  s3:
    bucket: "my-terraform-state-bucket"
    key: "path/to/terraform.tfstate"
    region: "us-east-1"
  # Or use a local file:
  # local:
  #   path: "/path/to/terraform.tfstate"

target:
  # Webhook target (optional)
  webhook:
    url: "https://example.com/infralog"
    method: "POST"  # POST or PUT (default: POST)
    retry:
      max_attempts: 3        # Number of retry attempts (default: 3)
      initial_delay_ms: 1000 # Initial delay in milliseconds (default: 1000)
      max_delay_ms: 30000    # Maximum delay in milliseconds (default: 30000)
      retry_on_status:       # HTTP status codes that trigger a retry
        - 500                # (default: [500, 502, 503, 504])
        - 502
        - 503
        - 504

  # Slack target (optional)
  slack:
    webhook_url: "https://hooks.slack.com/services/T00/B00/XXX"
    channel: "#infrastructure"  # Optional: override default channel
    username: "Infralog"        # Optional: override bot username
    icon_emoji: ":terraform:"   # Optional: override bot icon

  # Stdout target (optional)
  # If no targets are configured, stdout is used automatically
  stdout:
    enabled: true
    format: "text"  # "text" or "json" (default: "text")

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

## Backends

Infralog supports reading Terraform state files from multiple backend sources. Configure one backend in the `tfstate` section.

### S3

Read state files from Amazon S3:

```yaml
tfstate:
  s3:
    bucket: "my-terraform-state-bucket"
    key: "path/to/terraform.tfstate"
    region: "us-east-1"
```

AWS credentials are loaded from the standard AWS credential chain (environment variables, shared credentials file, IAM role, etc.).

### Local

Read state files from the local filesystem:

```yaml
tfstate:
  local:
    path: "/path/to/terraform.tfstate"
```

This is useful for development, testing, or when using Terraform's local backend.

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
    "timestamp": "2024-01-15T10:30:00Z",
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

#### Retry Behavior

Webhook requests are automatically retried on transient failures:

- Network errors trigger a retry
- Configurable HTTP status codes trigger a retry (default: 500, 502, 503, 504)
- Retries use exponential backoff with jitter to prevent thundering herd
- Non-retryable errors (e.g., 400 Bad Request) fail immediately

### Slack

Sends formatted notifications to a Slack channel using incoming webhooks.

To set up:

1. Create a Slack app at https://api.slack.com/apps
2. Enable "Incoming Webhooks" and create a webhook for your channel
3. Copy the webhook URL to your configuration

Messages include:

- Header with "Terraform State Changes Detected"
- State file location (bucket, key, region)
- Resource changes with color-coded status indicators
- Output changes with before/after values

### Stdout

Logs changes to standard output. This is useful for debugging, testing, or piping output to other tools.

**Automatic fallback**: If no targets are configured, stdout is used automatically with text format.

Two formats are available:

**Text format** (default):
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  TERRAFORM STATE CHANGES DETECTED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Time:   2024-01-15 10:30:00 UTC
  Bucket: my-bucket
  Key:    terraform.tfstate
  Region: us-east-1
──────────────────────────────────────────────────

  RESOURCE CHANGES

  [+] aws_instance.web (added)
  [~] aws_rds_instance.db (changed)
      instance_class: db.t2.micro → db.t2.small
  [-] aws_s3_bucket.old (removed)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**JSON format**: Same structure as the webhook payload, useful for piping to `jq` or other tools.

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
