---
title: "Configuration Reference"
---

## Config File

Network Guard Proxy reads its configuration from a YAML file specified via the `--config` flag:

```bash
./networkguard --config config.yaml
```

## Full Reference

```yaml
proxy:
  listen_address: ":8080"       # Public proxy listener address
  upstream_url: "http://localhost:9000"  # Backend to forward allowed requests

admin:
  listen_address: ":9090"       # Admin API and Web UI listener

storage:
  driver: "sqlite"              # Storage backend (currently only "sqlite")
  dsn: "./network_guard.db"     # Database file path

logging:
  level: "info"                 # Log level
  max_logs: 10000               # Maximum number of log entries to retain
```

## Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `proxy.listen_address` | string | `:8080` | Address for the public proxy listener |
| `proxy.upstream_url` | string | `http://localhost:9000` | URL of the upstream backend service |
| `admin.listen_address` | string | `:9090` | Address for the admin API and Web UI |
| `storage.driver` | string | `sqlite` | Storage driver |
| `storage.dsn` | string | `./network_guard.db` | Database connection string |
| `logging.level` | string | `info` | Logging verbosity |
| `logging.max_logs` | int | `10000` | Max log entries before pruning |

## Log Pruning

When `max_logs` is set, the proxy automatically prunes older log entries to keep the total count under the configured limit. Pruning occurs after each request is logged.
