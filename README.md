# Infralog

Analyze Terraform plan files and send notifications about infrastructure changes. Perfect for CI/CD pipelines.

## Quick Start

```bash
# Generate and convert plan
terraform plan -out=plan.tfplan
terraform show -json plan.tfplan > plan.json

# Analyze the plan
infralog -f plan.json
```

## Features

- Analyzes Terraform plan JSON files
- Sends notifications to Webhook and Slack
- Filters by resource type and output name
- Lightweight single binary

## Documentation

**[hvalls.github.io/infralog](https://hvalls.github.io/infralog/)**

## Installation

```bash
# Build from source
cd src/
go build -o infralog main.go
```

## Usage

```bash
# Basic usage
infralog -f plan.json

# With configuration
infralog -f plan.json --config-file config.yml
```

## License

MIT License
