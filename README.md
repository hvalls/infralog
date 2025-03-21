# Infralog

⚠ **Warning**: This project is in very early development stage and is not ready for production use. The API and features are subject to change.

Infralog monitors your Terraform state files and emits resource-level events when changes are detected.

## Build from sources
```bash
$ cd src/
$ go build -o infralog main.go
```

## config.yml

```yaml
polling:
  interval: 600 # in seconds
tfstate:
  s3:
    bucket: ""
    key: ""
    region: ""
target:
  webhook:
    url: "http://localhost:8080/infralog"
    method: "POST" # Only POST and PUT supported. Default is POST.
filter:
  resource_types: # If not specified, all resources will be monitored. Use [] to not monitor any resource.
    - "aws_instance"
    - "aws_s3_bucket"
  outputs: # If not specified, all outputs will be monitored. Use [] to not monitor any output.
    - "instance_ip"
```

## Usage
Using `--config-file` flag:
```bash
$ infralog --config-file config.yml
```
Using `INFRALOG_CONFIG_FILE` environment variable:
```bash
$ INFRALOG_CONFIG_FILE=config.yml infralog
```

## Targets

### Webhook target

- `POST` request will be made to the specified URL
- The request body will contain the JSON payload with the following structure:
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
                "bucket": "",
                "key": "",
                "region": ""
            }
        }
    }
}
```

## Contributions

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a new branch for your feature
3. Make your changes
4. Run tests if applicable
5. Commit your changes
6. Push to the branch
7. Open a pull request

Please make sure to:
- Update documentation as needed
- Add tests for new features when applicable
- Ensure all tests pass before submitting

For major changes, please open an issue first to discuss what you would like to change.
