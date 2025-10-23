package storage

import (
	"os"
	"testing"
	"time"

	"github.com/networkguard/proxy/internal/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	f, err := os.CreateTemp("", "ngp-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	store, err := NewSQLiteStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestRuleCRUD(t *testing.T) {
	store := newTestStore(t)
	now := time.Now().Truncate(time.Second)

	rule := &models.Rule{
		ID:          "test-1",
		Enabled:     true,
		Priority:    10,
		Action:      "deny",
		Match:       models.MatchConditions{Methods: []string{"DELETE"}},
		Description: "block DELETE",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create
	if err := store.CreateRule(rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}

	// Get
	got, err := store.GetRule("test-1")
	if err != nil {
		t.Fatalf("get rule: %v", err)
	}
	if got == nil {
		t.Fatal("expected rule, got nil")
	}
	if got.Action != "deny" || got.Priority != 10 {
		t.Errorf("unexpected rule: %+v", got)
	}
	if len(got.Match.Methods) != 1 || got.Match.Methods[0] != "DELETE" {
		t.Errorf("unexpected match: %+v", got.Match)
	}

	// List
	rules, err := store.ListRules()
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	// Update
	rule.Priority = 5
	rule.UpdatedAt = time.Now().Truncate(time.Second)
	if err := store.UpdateRule(rule); err != nil {
		t.Fatalf("update rule: %v", err)
	}
	got, _ = store.GetRule("test-1")
	if got.Priority != 5 {
		t.Errorf("expected priority 5, got %d", got.Priority)
	}

	// Delete
	if err := store.DeleteRule("test-1"); err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	got, _ = store.GetRule("test-1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestLogCRUD(t *testing.T) {
	store := newTestStore(t)

	for i := 0; i < 5; i++ {
		entry := &models.LogEntry{
			Timestamp:  time.Now(),
			ClientIP:   "192.168.1.1",
			Method:     "GET",
			Path:       "/test",
			StatusCode: 200,
			Decision:   "allow",
			Reason:     "ok",
			LatencyMs:  1.5,
		}
		if i == 3 {
			entry.Decision = "deny"
			entry.StatusCode = 403
			entry.ClientIP = "10.0.0.1"
		}
		if err := store.CreateLog(entry); err != nil {
			t.Fatalf("create log %d: %v", i, err)
		}
	}

	// Query all
	entries, err := store.QueryLogs(100, "", "", "")
	if err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 logs, got %d", len(entries))
	}

	// Query by decision
	entries, err = store.QueryLogs(100, "deny", "", "")
	if err != nil {
		t.Fatalf("query deny logs: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 deny log, got %d", len(entries))
	}

	// Query by IP
	entries, err = store.QueryLogs(100, "", "10.0.0.1", "")
	if err != nil {
		t.Fatalf("query by IP: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 log for 10.0.0.1, got %d", len(entries))
	}

	// Query by path
	entries, err = store.QueryLogs(100, "", "", "test")
	if err != nil {
		t.Fatalf("query by path: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 logs matching 'test', got %d", len(entries))
	}

	// Limit
	entries, err = store.QueryLogs(2, "", "", "")
	if err != nil {
		t.Fatalf("query with limit: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 logs with limit, got %d", len(entries))
	}
}

func TestPruneLogs(t *testing.T) {
	store := newTestStore(t)

	for i := 0; i < 10; i++ {
		entry := &models.LogEntry{
			Timestamp:  time.Now(),
			ClientIP:   "1.1.1.1",
			Method:     "GET",
			Path:       "/",
			StatusCode: 200,
			Decision:   "allow",
			LatencyMs:  1.0,
		}
		if err := store.CreateLog(entry); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}

	if err := store.PruneLogs(5); err != nil {
		t.Fatalf("prune: %v", err)
	}

	entries, _ := store.QueryLogs(100, "", "", "")
	if len(entries) != 5 {
		t.Errorf("expected 5 after prune, got %d", len(entries))
	}
}

func TestRuleCount(t *testing.T) {
	store := newTestStore(t)
	now := time.Now()

	store.CreateRule(&models.Rule{ID: "a", Enabled: true, Action: "deny", CreatedAt: now, UpdatedAt: now})
	store.CreateRule(&models.Rule{ID: "b", Enabled: true, Action: "allow", CreatedAt: now, UpdatedAt: now})
	store.CreateRule(&models.Rule{ID: "c", Enabled: false, Action: "deny", CreatedAt: now, UpdatedAt: now})

	total, enabled, err := store.RuleCount()
	if err != nil {
		t.Fatalf("rule count: %v", err)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
	if enabled != 2 {
		t.Errorf("expected 2 enabled, got %d", enabled)
	}
}
