<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" style="height:64px;margin-right:32px"/>

# Ten network guard proxy jest lepszy. Przygotuj mi po angielsku specyfikację tego projektu dla Claude code w pliku md makrdown żeby mógł podać ten plik do Claude code żeby mi go zaimplementował.

```markdown
# Network Guard Proxy – Project Specification

## 1. Overview

Network Guard Proxy is a lightweight **reverse proxy and application‑level firewall written in Go**.  
It sits between clients and backend services, **inspecting and filtering inbound HTTP traffic** based on configurable security rules, while exposing a small **web UI (JavaScript + HTML + CSS)** for rule management and live log viewing.

The goal is to demonstrate:

- Strong Go backend skills (concurrency, networking, clean architecture).
- Understanding of HTTP, reverse proxies, and basic security controls.
- Ability to build a small full‑stack system (Go API + JS UI).
- Experience working with configuration, persistence, and observability.

---

## 2. High‑Level Architecture

Components:

1. **Proxy Engine (Go)**  
   - Listens for incoming HTTP requests from clients.  
   - Applies a chain of **filter rules** (IP, path, method, rate limit, header checks, etc.).  
   - For allowed requests, forwards traffic to the configured **upstream backend** (one or more targets).  
   - For blocked requests, returns an HTTP error (e.g., 403 Forbidden) with a JSON error body.  
   - Logs each request decision (allowed/blocked + reason) to the log store.

2. **Control API (Go, REST/JSON)**  
   - Exposes a REST API to manage rules and view logs.  
   - Used by the web UI and can also be called via curl/postman.

3. **Web UI (JavaScript)**  
   - Single‑page style app served by the Go backend (static files).  
   - Allows basic CRUD for rules and live view of recent logs.

4. **Persistence Layer**  
   - Simple storage for:
     - Rule configurations.
     - Request logs (only recent N entries or time‑bound).  
   - Can be implemented with:
     - Option A (preferred for simplicity): embedded database (SQLite).  
     - Option B: JSON/YAML files on disk (if we want zero external dependencies).

5. **Configuration / Runtime**  
   - The proxy configuration (listen addresses, upstream target, DB path, log retention, etc.) is read from a **YAML/JSON config file** and/or environment variables.

---

## 3. Detailed Functional Requirements

### 3.1 Proxy Engine

- Runs as a single binary, e.g. `network-guard-proxy`.
- Exposes two listeners:
  - **Public listener** (e.g., `:8080`) – used by clients; all requests go through the filter chain and then to the upstream.  
  - **Admin listener** (e.g., `:9090`) – exposes the Control API and serves the Web UI.
- Supports HTTP/1.1 initially (no HTTPS termination needed for this MVP; can be behind nginx in real use).

**Request flow:**

1. Client sends `HTTP` request to `:8080`.
2. Proxy builds a context object with:
   - Timestamp, client IP, method, path, headers, body size.
3. Proxy applies rules in a defined order:
   - IP rules (allow/deny specific addresses / CIDR).
   - Path rules (allow/deny specific paths or prefixes).
   - Method rules (e.g., deny `DELETE` globally).
   - Rate limit rules (optional basic implementation).
   - Header checks (block if certain header present/missing or matches regex).
4. If any rule **blocks**:
   - Proxy returns `403 Forbidden` (or configured status) with JSON:
     ```json
     ```
     { "error": "blocked_by_rule", "rule_id": "<id>", "reason": "<short reason>" }
     ```
     ```
   - Request is **not** forwarded.
5. If request is **allowed**:
   - Proxy forwards it to the configured upstream (e.g., `http://localhost:9000`).
   - Proxy copies response status, headers, and body back to the client.

**Forwarding details:**

- Use Go’s `net/http` standard library (`ReverseProxy` or custom implementation).
- Add an `X-Network-Guard-Id` header so that upstream can see it came via the proxy.
- Preserve client IP via `X-Forwarded-For`.

### 3.2 Rule Model

Each rule has (minimum):

- `id` (string/UUID).
- `enabled` (bool).
- `priority` (int; lower = evaluated earlier).
- `action` (`"allow"` or `"deny"`).
- `match` object with fields (all optional – rule matches if **all provided** conditions are true):
  - `ip_cidr` – list of CIDR ranges (e.g., `"192.168.0.0/24"`).
  - `methods` – list of HTTP methods (e.g., `["GET", "POST"]`).
  - `path_prefixes` – list of path prefixes (e.g., `["/admin", "/api/private"]`).
  - `exact_paths` – list of exact paths.
  - `header_equals` – map of header name → exact value.
  - `header_contains` – map of header name → substring.
- `description` – human‑readable text.
- `created_at`, `updated_at`.

Rules are evaluated in ascending `priority`.  
First matching rule decides; if **no rule matches**, default action is **allow**.

### 3.3 Control API

Base URL: `http://admin-host:9090/api/v1`

Endpoints:

- `GET /rules`
  - Returns list of all rules.
- `POST /rules`
  - Creates a new rule.
  - Request body: JSON with rule fields (except id, timestamps).
- `GET /rules/{id}`
  - Returns rule by id.
- `PUT /rules/{id}`
  - Updates an existing rule.
- `DELETE /rules/{id}`
  - Deletes rule.
- `POST /rules/{id}/enable`
- `POST /rules/{id}/disable`

- `GET /logs`
  - Query parameters:
    - `limit` (int, default 100).
    - `decision` (`allow` or `deny`).
    - `ip` (optional filter by client IP).
    - `path_contains` (optional filter by substring).
  - Returns list of recent log entries.

Optional small endpoint:

- `GET /status`
  - Returns health info (uptime, rule count, version).

### 3.4 Log Model

Each log entry contains:

- `id`.
- `timestamp`.
- `client_ip`.
- `method`.
- `path`.
- `status_code` (returned to client).
- `decision` (`allow`/`deny`).
- `rule_id` (nullable if no rule matched).
- `reason` (short string).
- `latency_ms` (time spent processing + upstream call for allowed).

Log retention:

- Keep configurable, e.g. `max_logs` or `max_days`.
- On startup / periodically, prune older entries.

---

## 4. Web UI Requirements

The UI is a single‑page style app, implemented with **vanilla JavaScript** (no heavy framework needed).  
Served from `/` on the admin port (e.g., `http://localhost:9090/`) as static files:

- `index.html`
- `app.js`
- `styles.css`

### 4.1 Screens

1. **Dashboard / Status**
   - Shows basic information:
     - Proxy status (from `/status` API).
     - Count of rules (enabled/disabled).
     - Link to Rules and Logs views.

2. **Rules View**
   - Table of rules:
     - Columns: ID, Priority, Action, Summary of match conditions, Enabled.
   - Buttons:
     - Add new rule.
     - Edit existing rule (inline or modal).
     - Enable/disable.
     - Delete.
   - Uses the Control API for CRUD operations.

3. **Logs View**
   - Paginated or scrollable list of recent logs (most recent first).
   - Columns: Time, IP, Method, Path, Decision, Rule ID, Status Code.
   - Simple filters:
     - Decision (allow/deny).
     - Text filter on path.
   - Periodic auto‑refresh (e.g., every 5 seconds).

UI does not need authentication for this project (can be added later).

---

## 5. Configuration and CLI

- Configuration file (YAML or JSON), e.g. `config.yaml`:

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

- Binary accepts a `--config` flag:

```bash
network-guard-proxy --config ./config.yaml
```

- On startup:
    - Load config.
    - Initialize storage.
    - Start admin server, then proxy server.
    - Log startup information.

---

## 6. Non‑Functional Requirements

- Written in **Go** (latest stable, e.g. Go 1.22+).
- Code structured in clear packages, e.g.:
    - `cmd/networkguard/` – main package \& CLI.
    - `internal/proxy/` – proxy engine \& rule evaluation.
    - `internal/api/` – REST API handlers.
    - `internal/storage/` – DB/file storage abstraction.
    - `internal/models/` – data structures for rules and logs.
    - `web/` – static assets for UI.
- Use `context.Context` for all request handling.
- Graceful shutdown on SIGINT/SIGTERM (finish in‑flight requests).
- Minimal unit tests for:
    - Rule matching logic.
    - Basic API handlers (happy path).

---

## 7. Example Use Cases

1. **Block all traffic from a given IP range**
    - Rule with `action: "deny"`, `ip_cidr: ["10.0.0.0/8"]`.
2. **Allow only GET and POST, deny everything else**
    - High‑priority allow rule for `methods: ["GET", "POST"]`.
    - Lower‑priority deny rule for any method.
3. **Deny access to `/admin` paths**
    - Rule with `action: "deny"`, `path_prefixes: ["/admin"]`.
4. **Deny requests with suspicious header**
    - Rule with `header_contains: { "User-Agent": "sqlmap" }`.

These can be pre‑seeded as sample rules in development mode.

---

## 8. Deliverables

- Go module with full source code.
- Config example (`config.example.yaml`).
- Minimal README:
    - How to build.
    - How to run with a sample upstream service.
    - How to access the admin UI.
- Optional: simple mock upstream server (in Go) for local testing.

```
```

