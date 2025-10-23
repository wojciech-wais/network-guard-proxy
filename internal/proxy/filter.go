package proxy

import (
	"net"
	"strings"

	"github.com/networkguard/proxy/internal/models"
)

type FilterResult struct {
	Decision string // "allow" or "deny"
	RuleID   string
	Reason   string
}

func EvaluateRules(rules []models.Rule, clientIP, method, path string, headers map[string]string) FilterResult {
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if matchesRule(rule, clientIP, method, path, headers) {
			return FilterResult{
				Decision: rule.Action,
				RuleID:   rule.ID,
				Reason:   rule.Description,
			}
		}
	}
	// Default: allow
	return FilterResult{Decision: "allow"}
}

func matchesRule(rule models.Rule, clientIP, method, path string, headers map[string]string) bool {
	m := rule.Match

	if len(m.IPCIDRs) > 0 && !matchIPCIDR(clientIP, m.IPCIDRs) {
		return false
	}
	if len(m.Methods) > 0 && !matchMethods(method, m.Methods) {
		return false
	}
	if len(m.PathPrefixes) > 0 && !matchPathPrefixes(path, m.PathPrefixes) {
		return false
	}
	if len(m.ExactPaths) > 0 && !matchExactPaths(path, m.ExactPaths) {
		return false
	}
	if len(m.HeaderEquals) > 0 && !matchHeaderEquals(headers, m.HeaderEquals) {
		return false
	}
	if len(m.HeaderContains) > 0 && !matchHeaderContains(headers, m.HeaderContains) {
		return false
	}

	return true
}

func matchIPCIDR(clientIP string, cidrs []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func matchMethods(method string, methods []string) bool {
	for _, m := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func matchPathPrefixes(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func matchExactPaths(path string, exactPaths []string) bool {
	for _, p := range exactPaths {
		if path == p {
			return true
		}
	}
	return false
}

func matchHeaderEquals(headers map[string]string, expected map[string]string) bool {
	for key, val := range expected {
		headerVal, ok := headers[key]
		if !ok || !strings.EqualFold(headerVal, val) {
			return false
		}
	}
	return true
}

func matchHeaderContains(headers map[string]string, substrings map[string]string) bool {
	for key, substr := range substrings {
		headerVal, ok := headers[key]
		if !ok || !strings.Contains(strings.ToLower(headerVal), strings.ToLower(substr)) {
			return false
		}
	}
	return true
}
