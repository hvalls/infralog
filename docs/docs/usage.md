---
sidebar_position: 3
---

# Usage

## Running with Docker

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yml:/etc/infralog/config.yml:ro \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  infralog:latest
```

For local backend with Docker:

```bash
docker run -p 8080:8080 \
  -v /path/to/config.yml:/etc/infralog/config.yml:ro \
  -v /path/to/terraform.tfstate:/data/terraform.tfstate:ro \
  infralog:latest
```

The `-p 8080:8080` flag exposes the Prometheus metrics endpoint (when enabled in config).

Using docker-compose:

```bash
docker-compose up
```

## Running from Binary

Specify the configuration file using either the `--config-file` flag or the `INFRALOG_CONFIG_FILE` environment variable:

```bash
infralog --config-file config.yml
```

```bash
export INFRALOG_CONFIG_FILE=config.yml
infralog
```
