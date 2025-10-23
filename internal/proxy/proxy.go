package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/networkguard/proxy/internal/models"
	"github.com/networkguard/proxy/internal/storage"
)

type ProxyServer struct {
	upstream *url.URL
	proxy    *httputil.ReverseProxy
	store    storage.Store
	maxLogs  int
}

func NewProxyServer(upstreamURL string, store storage.Store, maxLogs int) (*ProxyServer, error) {
	u, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("parse upstream URL: %w", err)
	}

	rp := httputil.NewSingleHostReverseProxy(u)

	// Customize the director to preserve X-Forwarded-For
	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		originalDirector(req)
		// X-Forwarded-For is handled by ReverseProxy, but ensure it's set
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
				req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
			} else {
				req.Header.Set("X-Forwarded-For", clientIP)
			}
		}
	}

	return &ProxyServer{
		upstream: u,
		proxy:    rp,
		store:    store,
		maxLogs:  maxLogs,
	}, nil
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.New().String()

	clientIP := extractClientIP(r)
	headers := flattenHeaders(r.Header)

	rules, err := p.store.ListRules()
	if err != nil {
		log.Printf("ERROR: failed to load rules: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	result := EvaluateRules(rules, clientIP, r.Method, r.URL.Path, headers)

	if result.Decision == "deny" {
		latency := float64(time.Since(start).Microseconds()) / 1000.0
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Network-Guard-Id", requestID)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "blocked_by_rule",
			"rule_id": result.RuleID,
			"reason":  result.Reason,
		})

		p.logRequest(clientIP, r.Method, r.URL.Path, http.StatusForbidden, "deny", result.RuleID, result.Reason, latency)
		return
	}

	// Forward to upstream
	r.Header.Set("X-Network-Guard-Id", requestID)
	recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	p.proxy.ServeHTTP(recorder, r)

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	reason := result.Reason
	if reason == "" {
		reason = "no rule matched"
	}
	p.logRequest(clientIP, r.Method, r.URL.Path, recorder.statusCode, "allow", result.RuleID, reason, latency)
}

func (p *ProxyServer) logRequest(clientIP, method, path string, statusCode int, decision, ruleID, reason string, latencyMs float64) {
	entry := &models.LogEntry{
		Timestamp:  time.Now(),
		ClientIP:   clientIP,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Decision:   decision,
		RuleID:     ruleID,
		Reason:     reason,
		LatencyMs:  latencyMs,
	}
	if err := p.store.CreateLog(entry); err != nil {
		log.Printf("ERROR: failed to write log: %v", err)
	}

	// Periodic pruning
	if p.maxLogs > 0 {
		_ = p.store.PruneLogs(p.maxLogs)
	}
}

func extractClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func flattenHeaders(h http.Header) map[string]string {
	result := make(map[string]string, len(h))
	for key, vals := range h {
		result[key] = strings.Join(vals, ", ")
	}
	return result
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}
