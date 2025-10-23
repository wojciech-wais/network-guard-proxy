# Network Guard Proxy

A lightweight reverse proxy and application-level firewall written in Go. It sits between clients and backend services, inspecting and filtering inbound HTTP traffic based on configurable security rules, with a web UI for rule management and live log viewing.

## Features

- **Rule-based filtering** – IP CIDR, HTTP method, path prefix/exact, header equals/contains matching
- **Priority-based evaluation** – first matching rule wins; default action is allow
- **REST API** – full CRUD for rules, log querying, and status endpoint
- **Web UI** – vanilla JS dashboard with rule management and auto-refreshing log view
- **SQLite persistence** – pure Go driver (no CGO), automatic log pruning
- **Graceful shutdown** – clean termination on SIGINT/SIGTERM

## Prerequisites

- Go 1.24+

## Build

```bash
go build -o networkguard ./cmd/networkguard/
```

## Run

```bash
cp config.example.yaml config.yaml
# Edit config.yaml to set your upstream_url
./networkguard --config config.yaml
```

The proxy starts two listeners:

| Port | Purpose |
|------|---------|
| `:8080` | Public proxy – client traffic goes through the filter chain to the upstream |
| `:9090` | Admin – REST API + Web UI |

## Quick Smoke Test

Start a simple upstream server and the proxy:

```bash
# Terminal 1 – upstream
python3 -m http.server 9000

# Terminal 2 – proxy
./networkguard --config config.yaml
```

Open the Web UI at `http://localhost:9090`.

Create a deny rule via the API:

```bash
curl -s -X POST http://localhost:9090/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "priority": 1,
    "action": "deny",
    "enabled": true,
    "description": "Block admin paths",
    "match": {"path_prefixes": ["/admin"]}
  }' | jq .
```

Test it:

```bash
# Allowed
curl -i http://localhost:8080/

# Blocked (403)
curl -i http://localhost:8080/admin/secret
```

View logs:

```bash
curl -s http://localhost:9090/api/v1/logs?limit=5 | jq .
```

## Configuration

See `config.example.yaml`:

```yaml
proxy:
  listen_address: ":8080"
  upstream_url: "http://localhost:9000"
admin:
  listen_address: ":9090"
storage:
  driver: "sqlite"
  dsn: "./network_guard.db"
logging:
  level: "info"
  max_logs: 10000
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/rules` | List all rules |
| `POST` | `/api/v1/rules` | Create a rule |
| `GET` | `/api/v1/rules/{id}` | Get a rule |
| `PUT` | `/api/v1/rules/{id}` | Update a rule |
| `DELETE` | `/api/v1/rules/{id}` | Delete a rule |
| `POST` | `/api/v1/rules/{id}/enable` | Enable a rule |
| `POST` | `/api/v1/rules/{id}/disable` | Disable a rule |
| `GET` | `/api/v1/logs` | Query logs (params: `limit`, `decision`, `ip`, `path_contains`) |
| `GET` | `/api/v1/status` | Proxy status (version, uptime, rule counts) |

## Tests

```bash
go test ./...
```

## Project Structure

```
cmd/networkguard/main.go        CLI entry point
internal/
  config/config.go              YAML config loading
  models/rule.go                Rule model
  models/log.go                 Log entry model
  storage/storage.go            Storage interface
  storage/sqlite.go             SQLite implementation
  proxy/proxy.go                Reverse proxy engine
  proxy/filter.go               Rule matching / filter chain
  api/api.go                    REST API handlers
web/                            Static UI assets (HTML/JS/CSS)
docs/                           Hugo documentation site
config.example.yaml             Example configuration
```

## Documentation

The `docs/` directory contains a Hugo site. Build it with:

```bash
cd docs && hugo --minify
```

## License

MIT
