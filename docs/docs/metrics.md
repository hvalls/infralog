---
sidebar_position: 8
---

# Metrics

Infralog exposes Prometheus metrics at `/metrics` when enabled. This allows you to monitor polling status, change detection, and notification delivery.

## Configuration

```yaml
metrics:
  enabled: true
  address: ":8080"  # Default address
```

## Available Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `infralog_changes_total` | Counter | `type`, `resource_type` | Total infrastructure changes detected |
| `infralog_poll_errors_total` | Counter | `stage` | Polling errors (stage: fetch, parse, compare) |
| `infralog_last_successful_poll_timestamp` | Gauge | - | Unix timestamp of last successful poll |
| `infralog_notifications_sent_total` | Counter | `target` | Successful notifications by target |
| `infralog_notification_errors_total` | Counter | `target` | Failed notifications by target |

## Running with Docker

When running with Docker, expose the metrics port:

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yml:/etc/infralog/config.yml:ro \
  infralog:latest
```

Then access metrics at `http://localhost:8080/metrics`.
