---
title: "Architecture"
---

## System Components

Network Guard Proxy consists of four main components:

### 1. Proxy Engine

The proxy engine listens on the public port (default `:8080`) and handles all incoming HTTP requests. For each request it:

1. Builds a request context (client IP, method, path, headers)
2. Evaluates the filter chain against configured rules
3. Either blocks the request (403 Forbidden) or forwards it to the upstream backend
4. Logs the decision with latency information

The engine is built on Go's `net/http/httputil.ReverseProxy`.

### 2. Control API

A REST API running on the admin port (default `:9090`) provides CRUD operations for rules and query access to request logs. All endpoints return JSON.

### 3. Web UI

A vanilla JavaScript single-page application served as static files from the admin port. It provides a dashboard, rule management, and log viewing interface.

### 4. Storage Layer

An abstracted storage interface backed by SQLite (using `modernc.org/sqlite`, a pure-Go implementation requiring no CGO). Stores rules and request logs with automatic pruning.

## Package Structure

```
cmd/networkguard/main.go    - CLI entry point, server lifecycle
internal/
  config/config.go          - YAML configuration loading
  models/rule.go            - Rule data model
  models/log.go             - Log entry data model
  storage/storage.go        - Storage interface
  storage/sqlite.go         - SQLite implementation
  proxy/proxy.go            - Reverse proxy engine
  proxy/filter.go           - Rule matching / filter chain
  api/api.go                - REST API handlers
web/                        - Static UI assets
```

## Request Flow

```
Client -> :8080 -> Filter Chain -> [DENY] -> 403 JSON response
                                -> [ALLOW] -> Upstream Backend -> Response to Client
```

Every request is logged with its decision, matched rule ID, and latency.
