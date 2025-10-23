package proxy

import (
	"testing"

	"github.com/networkguard/proxy/internal/models"
)

func TestEvaluateRules_DefaultAllow(t *testing.T) {
	result := EvaluateRules(nil, "1.2.3.4", "GET", "/", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
}

func TestEvaluateRules_IPCIDRDeny(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{IPCIDRs: []string{"10.0.0.0/8"}},
			Description: "block internal",
		},
	}

	result := EvaluateRules(rules, "10.1.2.3", "GET", "/", nil)
	if result.Decision != "deny" {
		t.Errorf("expected deny, got %s", result.Decision)
	}
	if result.RuleID != "r1" {
		t.Errorf("expected rule r1, got %s", result.RuleID)
	}

	result = EvaluateRules(rules, "192.168.1.1", "GET", "/", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow for non-matching IP, got %s", result.Decision)
	}
}

func TestEvaluateRules_MethodDeny(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{Methods: []string{"DELETE", "PATCH"}},
			Description: "block destructive methods",
		},
	}

	result := EvaluateRules(rules, "1.2.3.4", "DELETE", "/resource", nil)
	if result.Decision != "deny" {
		t.Errorf("expected deny for DELETE, got %s", result.Decision)
	}

	result = EvaluateRules(rules, "1.2.3.4", "GET", "/resource", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow for GET, got %s", result.Decision)
	}
}

func TestEvaluateRules_PathPrefix(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{PathPrefixes: []string{"/admin", "/api/private"}},
			Description: "block admin paths",
		},
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"/admin/users", "deny"},
		{"/admin", "deny"},
		{"/api/private/keys", "deny"},
		{"/api/public", "allow"},
		{"/", "allow"},
	}

	for _, tt := range tests {
		result := EvaluateRules(rules, "1.2.3.4", "GET", tt.path, nil)
		if result.Decision != tt.expected {
			t.Errorf("path %s: expected %s, got %s", tt.path, tt.expected, result.Decision)
		}
	}
}

func TestEvaluateRules_ExactPath(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{ExactPaths: []string{"/secret"}},
			Description: "block secret",
		},
	}

	result := EvaluateRules(rules, "1.2.3.4", "GET", "/secret", nil)
	if result.Decision != "deny" {
		t.Errorf("expected deny for /secret, got %s", result.Decision)
	}

	result = EvaluateRules(rules, "1.2.3.4", "GET", "/secret/more", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow for /secret/more (not exact), got %s", result.Decision)
	}
}

func TestEvaluateRules_HeaderEquals(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{HeaderEquals: map[string]string{"X-Bad": "yes"}},
			Description: "block bad header",
		},
	}

	headers := map[string]string{"X-Bad": "yes"}
	result := EvaluateRules(rules, "1.2.3.4", "GET", "/", headers)
	if result.Decision != "deny" {
		t.Errorf("expected deny with matching header, got %s", result.Decision)
	}

	headers = map[string]string{"X-Bad": "no"}
	result = EvaluateRules(rules, "1.2.3.4", "GET", "/", headers)
	if result.Decision != "allow" {
		t.Errorf("expected allow with non-matching header, got %s", result.Decision)
	}
}

func TestEvaluateRules_HeaderContains(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{HeaderContains: map[string]string{"User-Agent": "sqlmap"}},
			Description: "block sqlmap",
		},
	}

	headers := map[string]string{"User-Agent": "Mozilla sqlmap/1.0"}
	result := EvaluateRules(rules, "1.2.3.4", "GET", "/", headers)
	if result.Decision != "deny" {
		t.Errorf("expected deny with sqlmap UA, got %s", result.Decision)
	}

	headers = map[string]string{"User-Agent": "Mozilla/5.0"}
	result = EvaluateRules(rules, "1.2.3.4", "GET", "/", headers)
	if result.Decision != "allow" {
		t.Errorf("expected allow with normal UA, got %s", result.Decision)
	}
}

func TestEvaluateRules_PriorityOrder(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "allow-get", Enabled: true, Priority: 1, Action: "allow",
			Match:       models.MatchConditions{Methods: []string{"GET"}},
			Description: "allow GET",
		},
		{
			ID: "deny-all", Enabled: true, Priority: 2, Action: "deny",
			Match:       models.MatchConditions{},
			Description: "deny everything else",
		},
	}

	result := EvaluateRules(rules, "1.2.3.4", "GET", "/", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow for GET (higher priority), got %s", result.Decision)
	}
	if result.RuleID != "allow-get" {
		t.Errorf("expected rule allow-get, got %s", result.RuleID)
	}

	result = EvaluateRules(rules, "1.2.3.4", "POST", "/", nil)
	if result.Decision != "deny" {
		t.Errorf("expected deny for POST, got %s", result.Decision)
	}
	if result.RuleID != "deny-all" {
		t.Errorf("expected rule deny-all, got %s", result.RuleID)
	}
}

func TestEvaluateRules_DisabledRuleSkipped(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: false, Priority: 1, Action: "deny",
			Match:       models.MatchConditions{},
			Description: "disabled deny-all",
		},
	}

	result := EvaluateRules(rules, "1.2.3.4", "GET", "/", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow when only rule is disabled, got %s", result.Decision)
	}
}

func TestEvaluateRules_MultipleConditionsAND(t *testing.T) {
	rules := []models.Rule{
		{
			ID: "r1", Enabled: true, Priority: 1, Action: "deny",
			Match: models.MatchConditions{
				Methods:      []string{"DELETE"},
				PathPrefixes: []string{"/admin"},
			},
			Description: "block DELETE on admin",
		},
	}

	// Both conditions match -> deny
	result := EvaluateRules(rules, "1.2.3.4", "DELETE", "/admin/users", nil)
	if result.Decision != "deny" {
		t.Errorf("expected deny when both conditions match, got %s", result.Decision)
	}

	// Only method matches -> allow (path doesn't match)
	result = EvaluateRules(rules, "1.2.3.4", "DELETE", "/public", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow when only method matches, got %s", result.Decision)
	}

	// Only path matches -> allow (method doesn't match)
	result = EvaluateRules(rules, "1.2.3.4", "GET", "/admin/users", nil)
	if result.Decision != "allow" {
		t.Errorf("expected allow when only path matches, got %s", result.Decision)
	}
}
