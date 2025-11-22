---
sidebar_position: 4
---

# Configuration

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

metrics:
  enabled: true
  address: ":8080"  # Address to expose /metrics endpoint
```
