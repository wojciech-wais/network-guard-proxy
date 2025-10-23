package storage

import (
	"github.com/networkguard/proxy/internal/models"
)

type Store interface {
	// Rules
	CreateRule(rule *models.Rule) error
	GetRule(id string) (*models.Rule, error)
	ListRules() ([]models.Rule, error)
	UpdateRule(rule *models.Rule) error
	DeleteRule(id string) error

	// Logs
	CreateLog(entry *models.LogEntry) error
	QueryLogs(limit int, decision, ip, pathContains string) ([]models.LogEntry, error)
	PruneLogs(maxLogs int) error

	// Stats
	RuleCount() (total int, enabled int, err error)

	Close() error
}
