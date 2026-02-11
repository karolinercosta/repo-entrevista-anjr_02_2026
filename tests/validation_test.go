package tests

import (
	"testing"
	"time"

	"example.com/tasksapi/models"
)

func TestIsValidStatus(t *testing.T) {
	valid := []string{"pending", "in_progress", "completed", "cancelled"}
	for _, s := range valid {
		if !models.IsValidStatus(s) {
			t.Errorf("expected status %s to be valid", s)
		}
	}
	invalid := []string{"foo", "done", "waiting"}
	for _, s := range invalid {
		if models.IsValidStatus(s) {
			t.Errorf("expected status %s to be invalid", s)
		}
	}
}

func TestIsValidPriority(t *testing.T) {
	valid := []string{"low", "medium", "high"}
	for _, p := range valid {
		if !models.IsValidPriority(p) {
			t.Errorf("expected priority %s to be valid", p)
		}
	}
	invalid := []string{"urgent", "none", "foo"}
	for _, p := range invalid {
		if models.IsValidPriority(p) {
			t.Errorf("expected priority %s to be invalid", p)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	today := time.Now().UTC()
	if !models.IsValidDate(today) {
		t.Errorf("expected today to be valid")
	}
	yesterday := today.AddDate(0, 0, -1)
	if models.IsValidDate(yesterday) {
		t.Errorf("expected yesterday to be invalid")
	}
	tomorrow := today.AddDate(0, 0, 1)
	if !models.IsValidDate(tomorrow) {
		t.Errorf("expected tomorrow to be valid")
	}
}

func TestIsCompletedTask(t *testing.T) {
	if !models.IsCompletedTask("completed") {
		t.Errorf("expected completed to be blocked")
	}
	if models.IsCompletedTask("pending") {
		t.Errorf("expected pending to not be blocked")
	}
}
