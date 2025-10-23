---
title: "Rule System"
---

## Overview

Network Guard Proxy uses a priority-based rule system. Rules are evaluated in ascending priority order (lower number = higher priority). The first matching rule decides the action. If no rule matches, the default action is **allow**.

## Rule Structure

Each rule has:

- **priority** (int): Evaluation order. Lower values are checked first.
- **action** (`allow` or `deny`): What to do when the rule matches.
- **enabled** (bool): Disabled rules are skipped during evaluation.
- **match**: Conditions that must ALL be true for the rule to match.

## Match Conditions

All conditions within a rule are combined with AND logic. If a condition field is empty/omitted, it matches everything.

### IP CIDR (`ip_cidr`)

Match client IP against CIDR ranges.

```json
{ "ip_cidr": ["10.0.0.0/8", "172.16.0.0/12"] }
```

### HTTP Methods (`methods`)

Match request HTTP method (case-insensitive).

```json
{ "methods": ["GET", "POST"] }
```

### Path Prefixes (`path_prefixes`)

Match if the request path starts with any of the given prefixes.

```json
{ "path_prefixes": ["/admin", "/api/private"] }
```

### Exact Paths (`exact_paths`)

Match if the request path exactly equals one of the given paths.

```json
{ "exact_paths": ["/health", "/ready"] }
```

### Header Equals (`header_equals`)

Match if the specified headers have exact values (case-insensitive comparison).

```json
{ "header_equals": { "X-API-Key": "secret123" } }
```

### Header Contains (`header_contains`)

Match if the specified headers contain the given substrings (case-insensitive).

```json
{ "header_contains": { "User-Agent": "sqlmap" } }
```

## Examples

### Block all traffic from internal network

```json
{
  "priority": 1,
  "action": "deny",
  "enabled": true,
  "description": "Block internal IPs",
  "match": { "ip_cidr": ["10.0.0.0/8"] }
}
```

### Allow only GET and POST, deny everything else

Create two rules:

1. Allow rule (priority 1):
```json
{
  "priority": 1,
  "action": "allow",
  "match": { "methods": ["GET", "POST"] }
}
```

2. Deny-all rule (priority 2):
```json
{
  "priority": 2,
  "action": "deny",
  "match": {},
  "description": "Deny all other methods"
}
```

### Block suspicious user agents

```json
{
  "priority": 5,
  "action": "deny",
  "description": "Block SQL injection tools",
  "match": { "header_contains": { "User-Agent": "sqlmap" } }
}
```
