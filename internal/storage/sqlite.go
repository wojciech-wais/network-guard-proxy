package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/networkguard/proxy/internal/models"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.createTables(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) createTables() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS rules (
			id TEXT PRIMARY KEY,
			enabled INTEGER NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 0,
			action TEXT NOT NULL DEFAULT 'deny',
			match_json TEXT NOT NULL DEFAULT '{}',
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			client_ip TEXT NOT NULL,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			decision TEXT NOT NULL,
			rule_id TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			latency_ms REAL NOT NULL DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_decision ON logs(decision);
	`)
	return err
}

func (s *SQLiteStore) CreateRule(rule *models.Rule) error {
	matchJSON, err := json.Marshal(rule.Match)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO rules (id, enabled, priority, action, match_json, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Enabled, rule.Priority, rule.Action, string(matchJSON),
		rule.Description, rule.CreatedAt.Format(time.RFC3339), rule.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetRule(id string) (*models.Rule, error) {
	row := s.db.QueryRow(`SELECT id, enabled, priority, action, match_json, description, created_at, updated_at FROM rules WHERE id = ?`, id)
	return s.scanRule(row)
}

func (s *SQLiteStore) scanRule(row *sql.Row) (*models.Rule, error) {
	var r models.Rule
	var enabled int
	var matchJSON, createdAt, updatedAt string
	err := row.Scan(&r.ID, &enabled, &r.Priority, &r.Action, &matchJSON, &r.Description, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	r.Enabled = enabled != 0
	if err := json.Unmarshal([]byte(matchJSON), &r.Match); err != nil {
		return nil, err
	}
	r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &r, nil
}

func (s *SQLiteStore) ListRules() ([]models.Rule, error) {
	rows, err := s.db.Query(`SELECT id, enabled, priority, action, match_json, description, created_at, updated_at FROM rules ORDER BY priority ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var r models.Rule
		var enabled int
		var matchJSON, createdAt, updatedAt string
		if err := rows.Scan(&r.ID, &enabled, &r.Priority, &r.Action, &matchJSON, &r.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		r.Enabled = enabled != 0
		if err := json.Unmarshal([]byte(matchJSON), &r.Match); err != nil {
			return nil, err
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *SQLiteStore) UpdateRule(rule *models.Rule) error {
	matchJSON, err := json.Marshal(rule.Match)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`UPDATE rules SET enabled=?, priority=?, action=?, match_json=?, description=?, updated_at=? WHERE id=?`,
		rule.Enabled, rule.Priority, rule.Action, string(matchJSON),
		rule.Description, rule.UpdatedAt.Format(time.RFC3339), rule.ID,
	)
	return err
}

func (s *SQLiteStore) DeleteRule(id string) error {
	_, err := s.db.Exec(`DELETE FROM rules WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) CreateLog(entry *models.LogEntry) error {
	_, err := s.db.Exec(
		`INSERT INTO logs (timestamp, client_ip, method, path, status_code, decision, rule_id, reason, latency_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.Timestamp.Format(time.RFC3339Nano), entry.ClientIP, entry.Method, entry.Path,
		entry.StatusCode, entry.Decision, entry.RuleID, entry.Reason, entry.LatencyMs,
	)
	return err
}

func (s *SQLiteStore) QueryLogs(limit int, decision, ip, pathContains string) ([]models.LogEntry, error) {
	query := `SELECT id, timestamp, client_ip, method, path, status_code, decision, rule_id, reason, latency_ms FROM logs WHERE 1=1`
	var args []interface{}

	if decision != "" {
		query += ` AND decision = ?`
		args = append(args, decision)
	}
	if ip != "" {
		query += ` AND client_ip = ?`
		args = append(args, ip)
	}
	if pathContains != "" {
		query += ` AND path LIKE ?`
		args = append(args, "%"+pathContains+"%")
	}

	query += ` ORDER BY id DESC`
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var e models.LogEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.ClientIP, &e.Method, &e.Path, &e.StatusCode, &e.Decision, &e.RuleID, &e.Reason, &e.LatencyMs); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SQLiteStore) PruneLogs(maxLogs int) error {
	_, err := s.db.Exec(`DELETE FROM logs WHERE id NOT IN (SELECT id FROM logs ORDER BY id DESC LIMIT ?)`, maxLogs)
	return err
}

func (s *SQLiteStore) RuleCount() (total int, enabled int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM rules`).Scan(&total)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM rules WHERE enabled = 1`).Scan(&enabled)
	return
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
