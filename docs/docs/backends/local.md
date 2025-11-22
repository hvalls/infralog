---
sidebar_position: 2
---

# Local Backend

Read state files from the local filesystem:

```yaml
tfstate:
  local:
    path: "/path/to/terraform.tfstate"
```

This is useful for development, testing, or when using Terraform's local backend.
