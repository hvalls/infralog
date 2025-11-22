---
sidebar_position: 1
---

# Webhook Target

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
            "before": "t2.small",
            "after": "t2.medium"
          }
        }
      }
    ],
    "output_diffs": [
      {
        "output_name": "instance_ip",
        "status": "changed",
        "value_diff": {
          "before": "10.0.1.20",
          "after": "10.0.1.28"
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

## Retry Behavior

Webhook requests are automatically retried on transient failures:

- Network errors trigger a retry
- Configurable HTTP status codes trigger a retry (default: 500, 502, 503, 504)
- Retries use exponential backoff with jitter to prevent thundering herd
- Non-retryable errors (e.g., 400 Bad Request) fail immediately
