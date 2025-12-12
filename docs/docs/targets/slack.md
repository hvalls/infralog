---
sidebar_position: 2
---

# Slack target

Sends formatted notifications to a Slack channel using incoming webhooks.

For configuration options, see the [Configuration](../configuration.md) page.

## Message format

```
Terraform Plan Changes
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Time: 2025-12-12 10:30:45 UTC

Git Context
👤 Committer: John Doe
🌿 Branch: feature/add-vpc
📝 Commit: abc123de
🔗 Repository: git@github.com:company/infrastructure.git

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Resource Changes
🟢 aws_instance.web_server - added
🟡 aws_s3_bucket.app_data - changed
    • instance_type: t2.micro → t2.small
🔴 aws_security_group.old_sg - removed
```