package tests

import (
	"encoding/json"
	"strings"
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
	now := time.Now().UTC()
	today := models.NewDate(now.Year(), now.Month(), now.Day())
	if !models.IsValidDate(today) {
		t.Errorf("expected today to be valid")
	}

	yesterdayTime := now.AddDate(0, 0, -1)
	yesterday := models.NewDate(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day())
	if models.IsValidDate(yesterday) {
		t.Errorf("expected yesterday to be invalid")
	}

	tomorrowTime := now.AddDate(0, 0, 1)
	tomorrow := models.NewDate(tomorrowTime.Year(), tomorrowTime.Month(), tomorrowTime.Day())
	if !models.IsValidDate(tomorrow) {
		t.Errorf("expected tomorrow to be valid")
	}
}

func TestParseDateOnly(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{"2026-02-15", false},
		{"2026-12-31", false},
		{"invalid", true},
		{"2026/02/15", true},
		{"15-02-2026", true},
	}

	for _, tt := range tests {
		date, err := models.ParseDateOnly(tt.input)
		if tt.shouldErr && err == nil {
			t.Errorf("expected error for input %s", tt.input)
		}
		if !tt.shouldErr && err != nil {
			t.Errorf("unexpected error for input %s: %v", tt.input, err)
		}
		if !tt.shouldErr && date.String() != tt.input {
			t.Errorf("expected %s, got %s", tt.input, date.String())
		}
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

func TestDateJSONMarshaling(t *testing.T) {
	type TaskWithDate struct {
		DueDate *models.Date `json:"due_date"`
	}

	t.Run("marshal date to JSON", func(t *testing.T) {
		date := models.NewDate(2026, 12, 31)
		task := TaskWithDate{DueDate: &date}

		jsonBytes, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		expected := `{"due_date":"2026-12-31"}`
		if string(jsonBytes) != expected {
			t.Errorf("expected %s, got %s", expected, string(jsonBytes))
		}
	})

	t.Run("unmarshal JSON to date", func(t *testing.T) {
		jsonStr := `{"due_date":"2026-12-31"}`
		var task TaskWithDate

		err := json.Unmarshal([]byte(jsonStr), &task)
		if err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if task.DueDate == nil {
			t.Fatal("expected due_date to be set")
		}
		if task.DueDate.String() != "2026-12-31" {
			t.Errorf("expected 2026-12-31, got %s", task.DueDate.String())
		}
	})

	t.Run("marshal null date", func(t *testing.T) {
		task := TaskWithDate{DueDate: nil}

		jsonBytes, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		expected := `{"due_date":null}`
		if string(jsonBytes) != expected {
			t.Errorf("expected %s, got %s", expected, string(jsonBytes))
		}
	})
}

func TestIsValidTitle(t *testing.T) {
	t.Run("valid titles", func(t *testing.T) {
		valid := []string{"abc", "Test", "A Valid Title", "123", strings.Repeat("x", 100)}
		for _, title := range valid {
			if !models.IsValidTitle(title) {
				t.Errorf("expected title '%s' to be valid", title)
			}
		}
	})

	t.Run("invalid titles", func(t *testing.T) {
		invalid := []string{"", "ab", strings.Repeat("x", 101)}
		for _, title := range invalid {
			if models.IsValidTitle(title) {
				t.Errorf("expected title '%s' to be invalid", title)
			}
		}
	})
}
