---
sidebar_position: 3
---

# Usage

## Basic Usage

The typical workflow for using Infralog:

```bash
# 1. Generate a Terraform plan
terraform plan -out=plan.tfplan

# 2. Convert plan to JSON format
terraform show -json plan.tfplan > plan.json

# 3. Analyze the plan
infralog -f plan.json
```

With a custom configuration file:

```bash
infralog -f plan.json --config-file config.yml
```

Or using environment variables (recommended for CI/CD):

```bash
export INFRALOG_TARGET_SLACK_WEBHOOK_URL="https://hooks.slack.com/services/XXX"
export INFRALOG_FILTER_RESOURCE_TYPES="aws_instance,aws_s3_bucket"
infralog -f plan.json
```

## CLI Flags

- `--plan-file` or `-f` (required): Path to Terraform plan JSON file
- `--config-file` (optional): Path to configuration YAML file

## Exit Codes

Infralog uses standard exit codes:

- `0`: Success - analysis completed (changes may or may not exist)
- `1`: Error - invalid arguments, file not found, parse errors, or target failures

## Docker Usage

Run Infralog in a container:

```bash
docker run -v $(pwd)/plan.json:/plan.json:ro \
           -v $(pwd)/config.yml:/config.yml:ro \
           hvalls/infralog:latest \
           -f /plan.json --config-file /config.yml
```

Or use environment variables instead of a config file:

```bash
docker run -v $(pwd)/plan.json:/plan.json:ro \
           -e INFRALOG_TARGET_SLACK_WEBHOOK_URL="https://hooks.slack.com/..." \
           -e INFRALOG_FILTER_RESOURCE_TYPES="aws_instance,aws_s3_bucket" \
           hvalls/infralog:latest \
           -f /plan.json
```

## Output Formats

### Default Output (No notification targets)

Simple summary with resource list:

```
$ infralog -f plan.json
✓ Plan analyzed: 3 resource(s) changed, 1 output(s) changed
  [+] aws_instance.web_server
  [~] aws_s3_bucket.app_data
  [-] aws_security_group.old_sg
  [~] output.instance_type
```

### With Notification Targets

Minimal confirmation when targets are configured:

```
$ infralog -f plan.json --config-file config.yml
✓ Plan analyzed: 3 resource(s) changed, 1 output(s) changed
✓ Webhook notification sent
✓ Slack notification sent
```

## Filtering Changes

Use filters to only receive notifications for specific resources:

```yaml
filter:
  # Only report changes to these resource types
  resource_types:
    - "aws_instance"
    - "aws_s3_bucket"
    - "aws_rds_cluster"

  # Only report changes to these outputs
  outputs:
    - "public_ip"
    - "database_endpoint"
```

Filter behavior:
- `nil` (omit field) = match all resources/outputs
- `[]` (empty list) = match none
- `["type1", "type2"]` = match only listed items

## Examples

### Notify on Any Infrastructure Change

```bash
# No config file = default to stdout with no filtering
infralog -f plan.json
```

### Send to Slack for Production Changes

```yaml
# config.yml
target:
  slack:
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"
    channel: "#production-changes"

filter:
  resource_types:
    - "aws_instance"
    - "aws_rds_instance"
    - "aws_s3_bucket"
```

```bash
infralog -f plan.json --config-file config.yml
```

### Webhook Integration for Approval Workflow

```yaml
# config.yml
target:
  webhook:
    url: "https://your-approval-system.com/api/terraform/changes"
    method: "POST"
    retry:
      max_attempts: 3
```

```bash
infralog -f plan.json --config-file config.yml
```

### Using Environment Variables in CI/CD

```bash
# GitHub Actions, GitLab CI, etc.
export INFRALOG_TARGET_WEBHOOK_URL="https://api.example.com/webhook"
export INFRALOG_TARGET_SLACK_WEBHOOK_URL="https://hooks.slack.com/services/XXX"
export INFRALOG_FILTER_RESOURCE_TYPES="aws_instance,aws_rds_instance"

terraform show -json plan.tfplan > plan.json
infralog -f plan.json  # No config file needed
```

See [Configuration](./configuration.md#environment-variables) for all available environment variables.
