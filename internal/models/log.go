package models

import "time"

type LogEntry struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	ClientIP   string    `json:"client_ip"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	Decision   string    `json:"decision"` // "allow" or "deny"
	RuleID     string    `json:"rule_id,omitempty"`
	Reason     string    `json:"reason"`
	LatencyMs  float64   `json:"latency_ms"`
}
