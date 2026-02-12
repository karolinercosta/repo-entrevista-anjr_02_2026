package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/tasksapi/handlers"
	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

func TestListTasksReturns200(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{} // Use NoOpLogger for tests
	h := handlers.NewAPI(s, logger)
	r := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()

	h.ListTasks(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateTaskWithValidData(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	taskJSON := `{"title":"Test Task","status":"pending","priority":"high"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskJSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateTaskMissingTitle(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	taskJSON := `{"status":"pending"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskJSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var errResp models.APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Code != 400 {
		t.Errorf("expected error code 400, got %d", errResp.Code)
	}
}
