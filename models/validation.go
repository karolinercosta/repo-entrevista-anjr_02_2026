package models

import (
	"time"
)

const StatusCompleted = "completed"

var (
	ValidStatuses   = map[string]struct{}{"pending": {}, "in_progress": {}, "completed": {}, "cancelled": {}}
	ValidPriorities = map[string]struct{}{"low": {}, "medium": {}, "high": {}}
)

func IsValidStatus(status string) bool {
	_, ok := ValidStatuses[status]
	return ok
}

func IsValidPriority(priority string) bool {
	_, ok := ValidPriorities[priority]
	return ok
}

func IsValidDate(due_date Date) bool {
	now := time.Now().UTC()
	due := due_date.UTC()
	// Trunca o horário das duas datas para comparação apenas de dias
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)

	return !dueDate.Before(nowDate)
}

func IsCompletedTask(status string) bool {
	return status == StatusCompleted
}

func IsValidTitle(title string) bool {
	return len(title) >= 3 && len(title) <= 100
}
