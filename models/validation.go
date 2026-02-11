package models

import (
	"time"

	"github.com/araddon/dateparse"
)

var (
	ValidStatuses   = []string{"pending", "in_progress", "completed", "cancelled"}
	ValidPriorities = []string{"low", "medium", "high"}
)

func IsValidStatus(status string) bool {
	for _, s := range ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
}

func IsValidPriority(priority string) bool {
	for _, p := range ValidPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

func IsValidDate(due_date time.Time) bool {
	now := time.Now().UTC()
	due := due_date.UTC()
	// Trunca o horário das duas datas para comparação apenas de dias
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)

	// Allow today or future dates
	return !dueDate.Before(nowDate)
}

// Usa o ParseDate araddon/dateparse para vários formatos automaticamente(YYYY-MM-DD, RFC3339, etc.)
func ParseDate(s string) (time.Time, error) {
	return dateparse.ParseAny(s)
}
