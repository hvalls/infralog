---
sidebar_position: 2
---

# Slack Target

Sends formatted notifications to a Slack channel using incoming webhooks.

## Setup

1. Create a Slack app at https://api.slack.com/apps
2. Enable "Incoming Webhooks" and create a webhook for your channel
3. Copy the webhook URL to your configuration

## Configuration

```yaml
target:
  slack:
    webhook_url: "https://hooks.slack.com/services/T00/B00/XXX"
    channel: "#infrastructure"  # Optional: override default channel
    username: "Infralog"        # Optional: override bot username
    icon_emoji: ":terraform:"   # Optional: override bot icon
```