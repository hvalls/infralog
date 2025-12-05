---
sidebar_position: 4
---

# Configuration

Infralog supports two configuration methods: YAML configuration files and environment variables. Both methods are **optional** - if not provided, Infralog will output a simple summary with no filtering.

**Environment variables take precedence over config file values**, allowing you to:
- Use config files for defaults
- Override specific values via environment variables in CI/CD
- Run without a config file using only environment variables

## Configuration File Structure

Create a `config.yml` file with the following structure:

```yaml
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

filter:
  # Optional: List of resource types to monitor.
  # Omit to monitor all resources, or use [] to monitor none.
  resource_types:
    - "aws_instance"
    - "aws_s3_bucket"
    - "aws_rds_cluster"
    - "aws_lambda_function"

  # Optional: List of outputs to monitor.
  # Omit to monitor all outputs, or use [] to monitor none.
  outputs:
    - "instance_ip"
    - "database_endpoint"
```

## Environment Variables

All configuration options can be set via environment variables with the `INFRALOG_` prefix. The variable name follows the config file structure in uppercase with underscores.

**Format:** `INFRALOG_<SECTION>_<SUBSECTION>_<KEY>`

### Available Environment Variables

#### Webhook Target

```bash
# Basic webhook configuration
export INFRALOG_TARGET_WEBHOOK_URL="https://example.com/webhook"
export INFRALOG_TARGET_WEBHOOK_METHOD="POST"

# Webhook retry configuration
export INFRALOG_TARGET_WEBHOOK_RETRY_MAX_ATTEMPTS=3
export INFRALOG_TARGET_WEBHOOK_RETRY_INITIAL_DELAY_MS=1000
export INFRALOG_TARGET_WEBHOOK_RETRY_MAX_DELAY_MS=30000
export INFRALOG_TARGET_WEBHOOK_RETRY_RETRY_ON_STATUS="500,502,503,504"
```

#### Slack Target

```bash
export INFRALOG_TARGET_SLACK_WEBHOOK_URL="https://hooks.slack.com/services/xxx"
export INFRALOG_TARGET_SLACK_CHANNEL="#infrastructure"
export INFRALOG_TARGET_SLACK_USERNAME="infralog-bot"
export INFRALOG_TARGET_SLACK_ICON_EMOJI=":robot:"
```

#### Filters

```bash
# Comma-separated lists for array values
export INFRALOG_FILTER_RESOURCE_TYPES="aws_instance,aws_s3_bucket,aws_vpc"
export INFRALOG_FILTER_OUTPUTS="public_ip,vpc_id"
```

## Targets

You can configure multiple targets simultaneously. Infralog will send notifications to all configured targets.

### Webhook

The webhook target sends HTTP POST/PUT requests with JSON payloads containing change information.

```yaml
target:
  webhook:
    url: "https://your-api.example.com/webhooks/terraform"
    method: "POST"
    retry:
      max_attempts: 3
      initial_delay_ms: 1000
      max_delay_ms: 30000
      retry_on_status: [500, 502, 503, 504]
```

**Payload structure:**

```json
{
  "plan": {
    "resource_changes": [
      {
        "type": "aws_instance",
        "name": "web",
        "change": {
          "actions": ["update"],
          "before": {
            "instance_type": "t2.micro"
          },
          "after": {
            "instance_type": "t2.small"
          }
        }
      }
    ],
    "output_changes": {}
  },
  "datetime": "2025-11-27T10:30:45Z"
}
```

See [Webhook Target](./targets/webhook.md) for more details.

### Slack

The Slack target sends formatted messages to Slack channels using incoming webhooks.

```yaml
target:
  slack:
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#infrastructure"
    username: "Infralog Bot"
    icon_emoji: ":terraform:"
```

See [Slack Target](./targets/slack.md) for more details.

## Output Formats

Infralog provides different output modes based on configuration:

**Local usage (no notification targets):**
```
$ infralog -f plan.json
✓ Plan analyzed: 3 resource(s) changed, 2 output(s) changed
  [+] aws_instance.web_server
  [~] aws_s3_bucket.app_data
  [-] aws_security_group.old_sg
```

**With notification targets:**
```
$ infralog -f plan.json --config-file config.yml
✓ Plan analyzed: 3 resource(s) changed, 2 output(s) changed
✓ Webhook notification sent
✓ Slack notification sent
```

## Filters

Filters control which resource types and outputs trigger notifications.

### Resource Type Filtering

```yaml
filter:
  resource_types:
    - "aws_instance"
    - "aws_s3_bucket"
    - "aws_rds_instance"
```

**Filter behavior:**
- **Omit the field** (or set to `null`): Monitor all resource types
- **Empty list** (`[]`): Monitor no resource types
- **List of types**: Monitor only the specified resource types

### Output Filtering

```yaml
filter:
  outputs:
    - "public_ip"
    - "database_endpoint"
    - "api_gateway_url"
```

**Filter behavior:**
- **Omit the field** (or set to `null`): Monitor all outputs
- **Empty list** (`[]`): Monitor no outputs
- **List of names**: Monitor only the specified outputs

## Configuration Examples

### Minimal Configuration

No config file needed - shows simple summary:

```bash
infralog -f plan.json
```

### Environment Variables Only

Run without a config file using only environment variables:

```bash
# Configure via environment variables
export INFRALOG_TARGET_SLACK_WEBHOOK_URL="https://hooks.slack.com/services/XXX"
export INFRALOG_TARGET_SLACK_CHANNEL="#infrastructure"
export INFRALOG_FILTER_RESOURCE_TYPES="aws_instance,aws_s3_bucket"

# Run without config file
infralog -f plan.json
```

### Multiple Targets

Send to both Slack and Webhook:

```yaml
target:
  webhook:
    url: "https://api.example.com/terraform/changes"
    method: "POST"
  slack:
    webhook_url: "https://hooks.slack.com/services/XXX/YYY/ZZZ"
    channel: "#ops"
```

### Production-Only Resources

Only notify about critical production resources:

```yaml
target:
  slack:
    webhook_url: "https://hooks.slack.com/services/XXX/YYY/ZZZ"
    channel: "#production-changes"

filter:
  resource_types:
    - "aws_instance"
    - "aws_rds_instance"
    - "aws_elasticache_cluster"
    - "aws_elb"
    - "aws_alb"
  outputs:
    - "database_endpoint"
    - "api_url"
```

### Webhook for Monitoring

Send to webhook for ingestion into logging systems:

```yaml
target:
  webhook:
    url: "https://logs.example.com/api/events"
    method: "POST"
```