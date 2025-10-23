---
title: "API Reference"
---

All API endpoints are served on the admin port (default `:9090`). Base path: `/api/v1`.

## Rules

### List Rules

```
GET /api/v1/rules
```

Returns all rules sorted by priority (ascending).

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "enabled": true,
    "priority": 1,
    "action": "deny",
    "match": { "path_prefixes": ["/admin"] },
    "description": "Block admin",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Rule

```
POST /api/v1/rules
```

**Request body:**
```json
{
  "priority": 1,
  "action": "deny",
  "enabled": true,
  "description": "Block admin paths",
  "match": {
    "path_prefixes": ["/admin"]
  }
}
```

**Response:** `201 Created` with the created rule (including generated `id`).

### Get Rule

```
GET /api/v1/rules/{id}
```

**Response:** `200 OK` with rule object, or `404 Not Found`.

### Update Rule

```
PUT /api/v1/rules/{id}
```

**Request body:** Full rule object (same as create).

**Response:** `200 OK` with updated rule.

### Delete Rule

```
DELETE /api/v1/rules/{id}
```

**Response:** `204 No Content`, or `404 Not Found`.

### Enable Rule

```
POST /api/v1/rules/{id}/enable
```

**Response:** `200 OK` with updated rule.

### Disable Rule

```
POST /api/v1/rules/{id}/disable
```

**Response:** `200 OK` with updated rule.

## Logs

### Query Logs

```
GET /api/v1/logs
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 100 | Max entries to return |
| `decision` | string | - | Filter by `allow` or `deny` |
| `ip` | string | - | Filter by client IP |
| `path_contains` | string | - | Filter by path substring |

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "timestamp": "2024-01-01T00:00:00Z",
    "client_ip": "192.168.1.1",
    "method": "GET",
    "path": "/admin",
    "status_code": 403,
    "decision": "deny",
    "rule_id": "uuid",
    "reason": "Block admin paths",
    "latency_ms": 0.5
  }
]
```

## Status

### Get Status

```
GET /api/v1/status
```

**Response:** `200 OK`
```json
{
  "version": "0.1.0",
  "uptime_seconds": 3600,
  "rules_total": 5,
  "rules_enabled": 3
}
```
