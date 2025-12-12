---
sidebar_position: 1
slug: /
---

# Introduction

Infralog analyzes Terraform plan files and notifies you about infrastructure changes before they're applied. Perfect for integrating into CI/CD pipelines to provide visibility and approval workflows for infrastructure changes.

## How it works

1. Terraform generates a plan file
2. Convert the plan to JSON format using `terraform show -json`
3. Infralog parses the plan and extracts resource/output changes
4. Changes are filtered based on your configuration
5. Notifications are sent to all configured targets


## Quick start

```bash
# Generate a Terraform plan
terraform plan -out=plan.tfplan

# Convert plan to JSON
terraform show -json plan.tfplan > plan.json

# Analyze the plan
infralog -f plan.json --config-file config.yml
```

