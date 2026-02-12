package tests

import (
	"testing"
	"time"

	"example.com/tasksapi/models"
)

func TestTaskServiceValidateCreate(t *testing.T) {
	logger := &models.NoOpLogger{}
	service := models.NewTaskService(logger)

	t.Run("valid task", func(t *testing.T) {
		task := models.Task{
			Title:  "Valid Task",
			Status: "pending",
		}
		err := service.ValidateCreate(task)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing title", func(t *testing.T) {
		task := models.Task{
			Status: "pending",
		}
		err := service.ValidateCreate(task)
		if err == nil {
			t.Error("expected error for missing title")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		task := models.Task{
			Title:  "Task",
			Status: "invalid_status",
		}
		err := service.ValidateCreate(task)
		if err == nil {
			t.Error("expected error for invalid status")
		}
	})

	t.Run("invalid priority", func(t *testing.T) {
		task := models.Task{
			Title:    "Task",
			Status:   "pending",
			Priority: "urgent",
		}
		err := service.ValidateCreate(task)
		if err == nil {
			t.Error("expected error for invalid priority")
		}
	})
}

func TestTaskServiceValidateUpdate(t *testing.T) {
	logger := &models.NoOpLogger{}
	service := models.NewTaskService(logger)

	t.Run("valid update", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{
			"title": "Updated Title",
		}
		err := service.ValidateUpdate(task, patch)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("prevent completed task edit", func(t *testing.T) {
		task := models.Task{Status: "completed"}
		patch := map[string]interface{}{
			"title": "Try to Update",
		}
		err := service.ValidateUpdate(task, patch)
		if err == nil {
			t.Error("expected error for editing completed task")
		}
		apiErr, ok := err.(*models.APIError)
		if !ok || apiErr.Code != 409 {
			t.Errorf("expected 409 business rule error, got %v", err)
		}
	})

	t.Run("empty patch", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{}
		err := service.ValidateUpdate(task, patch)
		if err == nil {
			t.Error("expected error for empty patch")
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{
			"unknown_field": "value",
		}
		err := service.ValidateUpdate(task, patch)
		if err == nil {
			t.Error("expected error for unknown field")
		}
	})

	t.Run("invalid status value", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{
			"status": "invalid",
		}
		err := service.ValidateUpdate(task, patch)
		if err == nil {
			t.Error("expected error for invalid status")
		}
	})

	t.Run("invalid priority value", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{
			"priority": "urgent",
		}
		err := service.ValidateUpdate(task, patch)
		if err == nil {
			t.Error("expected error for invalid priority")
		}
	})
}

func TestBusinessRules(t *testing.T) {
	t.Run("prevent completed task edits", func(t *testing.T) {
		task := models.Task{Status: "completed"}
		patch := map[string]interface{}{"title": "New"}

		err := models.PreventCompletedTaskEdits(task, patch)
		if err == nil {
			t.Error("expected business rule to prevent completed task edit")
		}
	})

	t.Run("allow pending task edits", func(t *testing.T) {
		task := models.Task{Status: "pending"}
		patch := map[string]interface{}{"title": "New"}

		err := models.PreventCompletedTaskEdits(task, patch)
		if err != nil {
			t.Errorf("expected no error for pending task, got %v", err)
		}
	})
}

func TestFieldValidators(t *testing.T) {
	t.Run("validate status field", func(t *testing.T) {
		patch := map[string]interface{}{}
		err := models.ValidateStatusField("pending", patch, "status")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if patch["status"] != "pending" {
			t.Error("expected patch to contain validated status")
		}
	})

	t.Run("validate priority field", func(t *testing.T) {
		patch := map[string]interface{}{}
		err := models.ValidatePriorityField("high", patch, "priority")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if patch["priority"] != "high" {
			t.Error("expected patch to contain validated priority")
		}
	})

	t.Run("validate string field", func(t *testing.T) {
		patch := map[string]interface{}{}
		err := models.ValidateStringField("test value", patch, "title")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("validate date field - future date string", func(t *testing.T) {
		patch := map[string]interface{}{}
		tomorrowTime := time.Now().UTC().AddDate(0, 0, 1)
		tomorrowStr := tomorrowTime.Format("2006-01-02")
		err := models.ValidateDueDateField(tomorrowStr, patch, "due_date")
		if err != nil {
			t.Errorf("expected no error for future date, got %v", err)
		}
	})

	t.Run("validate date field - past date fails", func(t *testing.T) {
		patch := map[string]interface{}{}
		yesterdayTime := time.Now().UTC().AddDate(0, 0, -1)
		yesterdayStr := yesterdayTime.Format("2006-01-02")
		err := models.ValidateDueDateField(yesterdayStr, patch, "due_date")
		if err == nil {
			t.Error("expected error for past date")
		}
	})

	t.Run("validate date field - Date type", func(t *testing.T) {
		patch := map[string]interface{}{}
		tomorrowTime := time.Now().UTC().AddDate(0, 0, 1)
		tomorrow := models.NewDate(tomorrowTime.Year(), tomorrowTime.Month(), tomorrowTime.Day())
		err := models.ValidateDueDateField(tomorrow, patch, "due_date")
		if err != nil {
			t.Errorf("expected no error for future Date, got %v", err)
		}
	})

	t.Run("validate date field - invalid format", func(t *testing.T) {
		patch := map[string]interface{}{}
		err := models.ValidateDueDateField("invalid-date", patch, "due_date")
		if err == nil {
			t.Error("expected error for invalid date format")
		}
	})
}

func TestLogger(t *testing.T) {
	t.Run("NoOpLogger doesn't panic", func(t *testing.T) {
		logger := &models.NoOpLogger{}
		logger.Info("test")
		logger.Warn("test")
		logger.Error("test")
		// Fatal would panic, so we test it separately
	})

	t.Run("DefaultLogger can be created", func(t *testing.T) {
		logger := models.NewDefaultLogger()
		if logger == nil {
			t.Error("expected logger to be created")
		}
	})

	t.Run("TaskService uses default logger if nil", func(t *testing.T) {
		service := models.NewTaskService(nil)
		if service == nil {
			t.Error("expected service to be created with default logger")
		}
	})
}
