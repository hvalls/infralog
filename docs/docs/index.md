---
sidebar_position: 1
slug: /
---

# Introduction

:::warning
This project is in early development. The API and features are subject to change.
:::

Infralog monitors Terraform state files and emits resource-level events when changes are detected. It polls your state file at configurable intervals, detects additions, modifications, and deletions, and delivers change notifications to configured targets.

## Features

- Monitors Terraform state files from multiple backends (S3, local filesystem)
- Detects resource and output changes with detailed diffs
- Configurable polling intervals
- Webhook notifications with JSON payloads and automatic retries
- Slack notifications with formatted messages
- Stdout logging with text or JSON format (default when no targets configured)
- Optional filtering by resource type and output name
- State persistence to detect changes across restarts
- Prometheus metrics endpoint for monitoring
