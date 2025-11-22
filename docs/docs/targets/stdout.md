---
sidebar_position: 3
---

# Stdout Target

Logs changes to standard output. This is useful for debugging, testing, or piping output to other tools.

**Automatic fallback**: If no targets are configured, stdout is used automatically with text format.

## Configuration

```yaml
target:
  stdout:
    enabled: true
    format: "text"  # "text" or "json" (default: "text")
```

## Text Format

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

## JSON Format

Emits one JSON line per change, optimized for log aggregation tools like Loki, Elasticsearch, or Splunk:

```json
{"timestamp":"2024-01-15T10:30:00Z","level":"info","msg":"resource added","event_type":"resource_change","source":"s3://my-bucket/terraform.tfstate","resource_type":"aws_instance","resource_name":"web","status":"added"}
{"timestamp":"2024-01-15T10:30:00Z","level":"info","msg":"resource changed","event_type":"resource_change","source":"s3://my-bucket/terraform.tfstate","resource_type":"aws_rds_instance","resource_name":"db","status":"changed","changes":{"instance_class":{"before":"db.t2.micro","after":"db.t2.small"}}}
{"timestamp":"2024-01-15T10:30:00Z","level":"info","msg":"output changed","event_type":"output_change","source":"s3://my-bucket/terraform.tfstate","output_name":"endpoint","status":"changed","changes":{"value":{"before":"old.example.com","after":"new.example.com"}}}
```

Each line is a valid JSON object with the following fields:

| Field | Description |
|-------|-------------|
| `timestamp` | ISO 8601 UTC timestamp |
| `level` | Log level (always "info") |
| `msg` | Human-readable message (e.g., "resource added") |
| `event_type` | Either "resource_change" or "output_change" |
| `source` | State file location (e.g., "s3://bucket/key" or "file://path") |
| `resource_type` | Terraform resource type (for resource changes) |
| `resource_name` | Terraform resource name (for resource changes) |
| `output_name` | Terraform output name (for output changes) |
| `status` | Change status: "added", "changed", or "removed" |
| `changes` | Attribute changes with `before`/`after` values (only for "changed" status) |

When using JSON format, operational messages (like "Polling...") are suppressed to keep the output clean for log ingestion.
