---
sidebar_position: 3
---

# Usage

## Basic usage

```bash
# 1. Generate a Terraform plan
terraform plan -out=plan.tfplan

# 2. Convert plan to JSON format
terraform show -json plan.tfplan > plan.json

# 3. Analyze the plan
infralog -f plan.json --config-file config.yml
```

## Run with Docker

```bash
docker run -v $(pwd)/plan.json:/plan.json:ro \
           -v $(pwd)/config.yml:/config.yml:ro \
           hvalls/infralog:latest \
           -f /plan.json --config-file /config.yml
```

## CLI flags

- `--plan-file` or `-f` (required): Path to Terraform plan JSON file
- `--config-file` (optional): Path to configuration YAML file

For configuration options, see the [Configuration](./configuration.md) page.
