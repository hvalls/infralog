# Infralog

⚠ **Warning**: This project is in very early development stage and is not ready for production use. The API and features are subject to change.

Infralog monitors your Terraform state files and emits resource-level events when changes are detected.

## Use Cases
- Trigger post-deployment verification scripts for specific resources
- Implement custom compliance checks when sensitive resources are modified
- Execute resource-specific cleanup operations after terraform destroy
- Build audit trails of infrastructure changes with resource-level granularity
- Automate cross-account or cross-region synchronization based on infrastructure changes
- Integrate with existing monitoring and alerting systems

## config.yml

```yaml
polling:
  interval: 10 # in seconds
tfstate:
  s3:
    bucket: ""
    key: ""
    region: ""
target:
  webhook:
    url: "http://localhost:8080/infralog"
filter:
  resource_types: # If not specified, all resources will be monitored. Use [] to not monitor any resource.
    - "aws_instance"
    - "aws_s3_bucket"
```

## Usage
```bash
$ infralog --config-file config.yml
```

## Build from sources
```bash
$ cd src/
$ go build -o infralog main.go
```

## Targets

### Webhook target

- `POST` request will be made to the specified URL
- The request body will contain the JSON payload with the following structure:
```json
{
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
}
```