---
title: "Web UI Guide"
---

## Accessing the Web UI

The Web UI is served at the root of the admin port:

```
http://localhost:9090/
```

No authentication is required (can be added behind a reverse proxy).

## Dashboard

The dashboard shows:

- **Version**: Current proxy version
- **Uptime**: How long the proxy has been running
- **Total Rules**: Number of configured rules
- **Enabled Rules**: Number of currently active rules

## Rules View

The rules view provides a table of all configured rules with:

- Priority, action, description, match summary, and enabled status
- **Add Rule**: Opens a form to create a new rule
- **Edit**: Modify an existing rule
- **Enable/Disable**: Toggle a rule without deleting it
- **Delete**: Permanently remove a rule

### Creating a Rule

Click "Add Rule" to open the rule form. Fill in:

1. **Priority**: Lower numbers are evaluated first
2. **Action**: Allow or Deny
3. **Description**: Human-readable explanation
4. **Match conditions**: Any combination of IP CIDRs, methods, path prefixes, exact paths, and header conditions

Use comma-separated values for lists. Header conditions use `key:value` format.

## Logs View

The logs view shows recent proxy decisions with:

- Timestamp, client IP, method, path, status code, decision, and matching rule ID
- **Decision filter**: Show only allowed or denied requests
- **Path filter**: Search logs by path substring
- **Auto-refresh**: Toggle automatic 5-second refresh
