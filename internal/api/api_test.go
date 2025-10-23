package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/networkguard/proxy/internal/models"
	"github.com/networkguard/proxy/internal/storage"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	f, err := os.CreateTemp("", "ngp-api-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	store, err := storage.NewSQLiteStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	return NewServer(store, "test", "")
}

func TestCreateAndListRules(t *testing.T) {
	srv := newTestServer(t)

	// Create a rule
	body := `{"priority":10,"action":"deny","enabled":true,"description":"block admin","match":{"path_prefixes":["/admin"]}}`
	req := httptest.NewRequest("POST", "/api/v1/rules", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.Rule
	json.NewDecoder(w.Body).Decode(&created)
	if created.ID == "" {
		t.Fatal("expected rule ID to be set")
	}
	if created.Action != "deny" {
		t.Errorf("expected deny, got %s", created.Action)
	}

	// List rules
	req = httptest.NewRequest("GET", "/api/v1/rules", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var rules []models.Rule
	json.NewDecoder(w.Body).Decode(&rules)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

func TestGetUpdateDeleteRule(t *testing.T) {
	srv := newTestServer(t)

	// Create
	body := `{"priority":5,"action":"allow","enabled":true,"description":"allow GET","match":{"methods":["GET"]}}`
	req := httptest.NewRequest("POST", "/api/v1/rules", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var created models.Rule
	json.NewDecoder(w.Body).Decode(&created)

	// Get
	req = httptest.NewRequest("GET", "/api/v1/rules/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// Update
	updateBody := `{"priority":1,"action":"deny","enabled":true,"description":"updated","match":{"methods":["POST"]}}`
	req = httptest.NewRequest("PUT", "/api/v1/rules/"+created.ID, bytes.NewBufferString(updateBody))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Rule
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Priority != 1 || updated.Description != "updated" {
		t.Errorf("unexpected update result: %+v", updated)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/api/v1/rules/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", w.Code)
	}

	// Verify deleted
	req = httptest.NewRequest("GET", "/api/v1/rules/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

func TestEnableDisableRule(t *testing.T) {
	srv := newTestServer(t)

	body := `{"priority":1,"action":"deny","enabled":true,"description":"test","match":{}}`
	req := httptest.NewRequest("POST", "/api/v1/rules", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var created models.Rule
	json.NewDecoder(w.Body).Decode(&created)

	// Disable
	req = httptest.NewRequest("POST", "/api/v1/rules/"+created.ID+"/disable", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("disable: expected 200, got %d", w.Code)
	}

	var disabled models.Rule
	json.NewDecoder(w.Body).Decode(&disabled)
	if disabled.Enabled {
		t.Error("expected disabled")
	}

	// Enable
	req = httptest.NewRequest("POST", "/api/v1/rules/"+created.ID+"/enable", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("enable: expected 200, got %d", w.Code)
	}

	var enabled models.Rule
	json.NewDecoder(w.Body).Decode(&enabled)
	if !enabled.Enabled {
		t.Error("expected enabled")
	}
}

func TestStatusEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var status map[string]interface{}
	json.NewDecoder(w.Body).Decode(&status)
	if status["version"] != "test" {
		t.Errorf("expected version 'test', got %v", status["version"])
	}
}

func TestLogsEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/logs", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var logs []models.LogEntry
	json.NewDecoder(w.Body).Decode(&logs)
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}
