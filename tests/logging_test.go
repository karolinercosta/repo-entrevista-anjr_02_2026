package tests

import (
	"strings"
	"testing"

	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

// MockLogger captura logs para testes
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.logs = append(m.logs, formatLog("INFO", msg, args...))
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.logs = append(m.logs, formatLog("WARN", msg, args...))
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.logs = append(m.logs, formatLog("ERROR", msg, args...))
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	panic(formatLog("FATAL", msg, args...))
}

func formatLog(level string, msg string, args ...interface{}) string {
	// Simples formatação para testes
	return level + ": " + msg
}

func (m *MockLogger) Contains(substr string) bool {
	for _, log := range m.logs {
		if strings.Contains(log, substr) {
			return true
		}
	}
	return false
}

func TestLoggingStore(t *testing.T) {
	mockLogger := &MockLogger{logs: make([]string, 0)}
	baseStore := store.New()
	loggingStore := store.NewLoggingStore(baseStore, mockLogger)

	t.Run("logs create operation", func(t *testing.T) {
		mockLogger.logs = nil // Reset logs
		task := models.Task{
			Title:  "Test Task",
			Status: "pending",
		}

		loggingStore.Create(task)

		if !mockLogger.Contains("[STORE] Creating task") {
			t.Error("expected create start log")
		}
		if !mockLogger.Contains("[STORE] Created task") {
			t.Error("expected create completion log")
		}
	})

	t.Run("logs list operation", func(t *testing.T) {
		mockLogger.logs = nil

		loggingStore.List()

		if !mockLogger.Contains("[STORE] Listing all tasks") {
			t.Error("expected list start log")
		}
		if !mockLogger.Contains("[STORE] Listed") {
			t.Error("expected list completion log with count")
		}
	})

	t.Run("logs get operation", func(t *testing.T) {
		mockLogger.logs = nil
		task := loggingStore.Create(models.Task{Title: "Get Test", Status: "pending"})
		mockLogger.logs = nil // Reset after create

		loggingStore.Get(task.ID)

		if !mockLogger.Contains("[STORE] Getting task") {
			t.Error("expected get start log")
		}
		if !mockLogger.Contains("[STORE] Retrieved task") {
			t.Error("expected get completion log")
		}
	})

	t.Run("logs update operation", func(t *testing.T) {
		mockLogger.logs = nil
		task := loggingStore.Create(models.Task{Title: "Update Test", Status: "pending"})
		mockLogger.logs = nil

		patch := map[string]interface{}{"title": "Updated"}
		loggingStore.Update(task.ID, patch)

		if !mockLogger.Contains("[STORE] Updating task") {
			t.Error("expected update start log")
		}
		if !mockLogger.Contains("[STORE] Updated task") {
			t.Error("expected update completion log")
		}
	})

	t.Run("logs delete operation", func(t *testing.T) {
		mockLogger.logs = nil
		task := loggingStore.Create(models.Task{Title: "Delete Test", Status: "pending"})
		mockLogger.logs = nil

		loggingStore.Delete(task.ID)

		if !mockLogger.Contains("[STORE] Deleting task") {
			t.Error("expected delete start log")
		}
		if !mockLogger.Contains("[STORE] Deleted task") {
			t.Error("expected delete completion log")
		}
	})

	t.Run("logs errors", func(t *testing.T) {
		mockLogger.logs = nil

		_, err := loggingStore.Get("nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent task")
		}
		if !mockLogger.Contains("Failed to get task") {
			t.Error("expected error log for failed get")
		}
	})
}
