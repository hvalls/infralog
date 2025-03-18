# Infralog

⚠ **Warning**: This project is in very early development stage and is not ready for production use. The API and features are subject to change.

Infralog monitors your Terraform state files and emits resource-level events when changes are detected.

## Use Cases
- Trigger post-deployment verification scripts for specific resources
- Implement custom compliance checks when sensitive resources are modified
- Execute resource-specific cleanup operations after terraform destroy
- Build audit trails of infrastructure changes with resource-level granularity
- Automate cross-account or cross-region synchronization based on infrastructure changes
- Integrate with existing monitoring and alerting systems

## config.yml

```yaml
polling:
  interval: 10 # in seconds
tfstate:
  s3:
    bucket: ""
    key: ""
    region: ""
```
