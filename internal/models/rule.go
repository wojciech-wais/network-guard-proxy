package models

import "time"

type MatchConditions struct {
	IPCIDRs        []string          `json:"ip_cidr,omitempty"`
	Methods        []string          `json:"methods,omitempty"`
	PathPrefixes   []string          `json:"path_prefixes,omitempty"`
	ExactPaths     []string          `json:"exact_paths,omitempty"`
	HeaderEquals   map[string]string `json:"header_equals,omitempty"`
	HeaderContains map[string]string `json:"header_contains,omitempty"`
}

type Rule struct {
	ID          string          `json:"id"`
	Enabled     bool            `json:"enabled"`
	Priority    int             `json:"priority"`
	Action      string          `json:"action"` // "allow" or "deny"
	Match       MatchConditions `json:"match"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
