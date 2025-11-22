---
sidebar_position: 7
---

# Persistence

By default, Infralog stores the last-seen state in memory. This means that if Infralog restarts, any changes that occurred during downtime will not be detected.

To enable change detection across restarts, configure the `persistence.state_file` option:

```yaml
persistence:
  state_file: "/var/lib/infralog/state.json"
```

## How It Works

- The last-seen state is saved to disk after each detected change
- On startup, Infralog loads the persisted state and compares it against the current state
- Changes that occurred while Infralog was stopped are detected and emitted

State files are written atomically to prevent corruption.
