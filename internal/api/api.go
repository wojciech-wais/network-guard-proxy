package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/networkguard/proxy/internal/models"
	"github.com/networkguard/proxy/internal/storage"
)

type Server struct {
	store     storage.Store
	startTime time.Time
	version   string
	mux       *http.ServeMux
	webDir    string
}

func NewServer(store storage.Store, version, webDir string) *Server {
	s := &Server{
		store:     store,
		startTime: time.Now(),
		version:   version,
		webDir:    webDir,
	}
	s.mux = http.NewServeMux()
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/rules", s.handleRules)
	s.mux.HandleFunc("/api/v1/rules/", s.handleRule)
	s.mux.HandleFunc("/api/v1/logs", s.handleLogs)
	s.mux.HandleFunc("/api/v1/status", s.handleStatus)

	if s.webDir != "" {
		fs := http.FileServer(http.Dir(s.webDir))
		s.mux.Handle("/", fs)
	}
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listRules(w, r)
	case http.MethodPost:
		s.createRule(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRule(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/rules/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		http.Error(w, "missing rule id", http.StatusBadRequest)
		return
	}

	// Check for enable/disable sub-path
	if len(parts) == 2 {
		switch parts[1] {
		case "enable":
			if r.Method == http.MethodPost {
				s.toggleRule(w, r, id, true)
				return
			}
		case "disable":
			if r.Method == http.MethodPost {
				s.toggleRule(w, r, id, false)
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getRule(w, r, id)
	case http.MethodPut:
		s.updateRule(w, r, id)
	case http.MethodDelete:
		s.deleteRule(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListRules()
	if err != nil {
		log.Printf("ERROR: list rules: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if rules == nil {
		rules = []models.Rule{}
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *Server) createRule(w http.ResponseWriter, r *http.Request) {
	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	rule.ID = uuid.New().String()
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	if rule.Action != "allow" && rule.Action != "deny" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'allow' or 'deny'"})
		return
	}

	if err := s.store.CreateRule(&rule); err != nil {
		log.Printf("ERROR: create rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) getRule(w http.ResponseWriter, r *http.Request, id string) {
	rule, err := s.store.GetRule(id)
	if err != nil {
		log.Printf("ERROR: get rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if rule == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) updateRule(w http.ResponseWriter, r *http.Request, id string) {
	existing, err := s.store.GetRule(id)
	if err != nil {
		log.Printf("ERROR: get rule for update: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if existing == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}

	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	rule.ID = id
	rule.CreatedAt = existing.CreatedAt
	rule.UpdatedAt = time.Now()

	if rule.Action != "allow" && rule.Action != "deny" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'allow' or 'deny'"})
		return
	}

	if err := s.store.UpdateRule(&rule); err != nil {
		log.Printf("ERROR: update rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) deleteRule(w http.ResponseWriter, r *http.Request, id string) {
	existing, err := s.store.GetRule(id)
	if err != nil {
		log.Printf("ERROR: delete rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if existing == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}

	if err := s.store.DeleteRule(id); err != nil {
		log.Printf("ERROR: delete rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) toggleRule(w http.ResponseWriter, r *http.Request, id string, enabled bool) {
	rule, err := s.store.GetRule(id)
	if err != nil {
		log.Printf("ERROR: toggle rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if rule == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}

	rule.Enabled = enabled
	rule.UpdatedAt = time.Now()
	if err := s.store.UpdateRule(rule); err != nil {
		log.Printf("ERROR: toggle rule: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	decision := r.URL.Query().Get("decision")
	ip := r.URL.Query().Get("ip")
	pathContains := r.URL.Query().Get("path_contains")

	entries, err := s.store.QueryLogs(limit, decision, ip, pathContains)
	if err != nil {
		log.Printf("ERROR: query logs: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if entries == nil {
		entries = []models.LogEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	total, enabled, err := s.store.RuleCount()
	if err != nil {
		log.Printf("ERROR: rule count: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"version":        s.version,
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
		"rules_total":    total,
		"rules_enabled":  enabled,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
