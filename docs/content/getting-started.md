---
title: "Getting Started"
---

## Prerequisites

- Go 1.24 or later

## Installation

Clone the repository and build:

```bash
git clone https://github.com/networkguard/proxy.git
cd proxy
go build ./cmd/networkguard/
```

## Configuration

Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to set your upstream backend URL:

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

## Running

Start the proxy:

```bash
./networkguard --config config.yaml
```

The proxy will listen on two ports:
- **:8080** - Public proxy port (client traffic)
- **:9090** - Admin port (API + Web UI)

## Quick Test

You can set up a simple upstream server for testing:

```bash
python3 -m http.server 9000 &
./networkguard --config config.yaml
```

Then access the Web UI at `http://localhost:9090` and send requests through the proxy at `http://localhost:8080`.

## Creating Your First Rule

Use the API to create a rule that blocks access to `/admin` paths:

```bash
curl -X POST http://localhost:9090/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "priority": 1,
    "action": "deny",
    "enabled": true,
    "description": "Block admin paths",
    "match": {
      "path_prefixes": ["/admin"]
    }
  }'
```

Test the rule:

```bash
curl -i http://localhost:8080/admin/secret
# Returns 403 Forbidden
```
