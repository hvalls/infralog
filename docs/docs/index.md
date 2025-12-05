---
sidebar_position: 1
slug: /
---

# Introduction

Infralog analyzes Terraform plan files and notifies you about infrastructure changes **before** they're applied. Perfect for integrating into CI/CD pipelines to provide visibility and approval workflows for infrastructure changes.

## Features

- **Plan Analysis**: Analyzes Terraform plan JSON output to detect changes
- **Change Detection**: Identifies resource additions, modifications, and deletions with detailed attribute diffs
- **Multiple Targets**: Webhook and Slack notifications
- **Flexible Output**: Simple summary, minimal notifications, or JSON format
- **Filtering**: Optional filtering by resource type and output name
- **CI/CD Native**: One-time execution designed for pipeline integration
- **Lightweight**: Single binary with no external dependencies

## Quick Start

```bash
# Generate a Terraform plan
terraform plan -out=plan.tfplan

# Convert plan to JSON
terraform show -json plan.tfplan > plan.json

# Analyze the plan
infralog -f plan.json
```

## Use Cases

- **PR Comments**: Post infrastructure changes in pull request comments
- **Approval Workflows**: Trigger manual approvals for specific resource types
- **Audit Trail**: Send changes to webhook endpoints for logging
- **Team Notifications**: Alert teams in Slack about planned infrastructure changes
- **Cost Estimation**: Integrate with cost estimation tools by analyzing resource changes

## How It Works

1. Terraform generates a plan file
2. Convert the plan to JSON format using `terraform show -json`
3. Infralog parses the plan and extracts resource/output changes
4. Changes are filtered based on your configuration
5. Notifications are sent to all configured targets
6. Exit code indicates success or failure

## Next Steps

- [Installation](./installation.md) - How to install Infralog
- [Usage](./usage.md) - Basic and advanced usage patterns
- [Configuration](./configuration.md) - Configure targets and filters
- [Targets](./targets/webhook.md) - Available notification targets
