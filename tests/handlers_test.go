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

func TestCreateTaskWithDueDate(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	taskJSON := `{"title":"Task with Date","status":"pending","due_date":"2026-12-31"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskJSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var task models.Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.DueDate == nil {
		t.Error("expected due_date to be set")
	}
	if task.DueDate.String() != "2026-12-31" {
		t.Errorf("expected due_date '2026-12-31', got '%s'", task.DueDate.String())
	}
}

func TestCreateTaskWithPastDueDate(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	taskJSON := `{"title":"Task","status":"pending","due_date":"2020-01-01"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskJSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for past date, got %d", w.Code)
	}
}

func TestListTasksFilterByDueDate(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	// Create task with due_date
	task1JSON := `{"title":"Task 1","status":"pending","due_date":"2026-12-31"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task1JSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateTask(w, r)

	// Create task without due_date
	task2JSON := `{"title":"Task 2","status":"pending"}`
	r = httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task2JSON))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.CreateTask(w, r)

	// Filter by due_date
	r = httptest.NewRequest("GET", "/tasks?due_date=2026-12-31", nil)
	w = httptest.NewRecorder()
	h.ListTasks(w, r)

	var response models.TaskListResponse
	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(response.Tasks))
	}
	if response.Tasks[0].Title != "Task 1" {
		t.Errorf("expected 'Task 1', got '%s'", response.Tasks[0].Title)
	}
	if response.TotalItems != 1 {
		t.Errorf("expected total_items=1, got %d", response.TotalItems)
	}
}

func TestListTasksFilterByNullDueDate(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	// Create task with due_date
	task1JSON := `{"title":"Task 1","status":"pending","due_date":"2026-12-31"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task1JSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateTask(w, r)

	// Create task without due_date
	task2JSON := `{"title":"Task 2","status":"pending"}`
	r = httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task2JSON))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.CreateTask(w, r)

	// Filter by null due_date
	r = httptest.NewRequest("GET", "/tasks?due_date=null", nil)
	w = httptest.NewRecorder()
	h.ListTasks(w, r)

	var response models.TaskListResponse
	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(response.Tasks))
	}
	if response.Tasks[0].Title != "Task 2" {
		t.Errorf("expected 'Task 2', got '%s'", response.Tasks[0].Title)
	}
	if response.TotalItems != 1 {
		t.Errorf("expected total_items=1, got %d", response.TotalItems)
	}
}

func TestListTasksFilterByNullPriority(t *testing.T) {
	s := store.New()
	logger := &models.NoOpLogger{}
	h := handlers.NewAPI(s, logger)

	// Create task with priority
	task1JSON := `{"title":"Task 1","status":"pending","priority":"high"}`
	r := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task1JSON))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateTask(w, r)

	// Create task without priority
	task2JSON := `{"title":"Task 2","status":"pending"}`
	r = httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(task2JSON))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.CreateTask(w, r)

	// Filter by null priority
	r = httptest.NewRequest("GET", "/tasks?priority=null", nil)
	w = httptest.NewRecorder()
	h.ListTasks(w, r)

	var response models.TaskListResponse
	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(response.Tasks))
	}
	if response.Tasks[0].Title != "Task 2" {
		t.Errorf("expected 'Task 2', got '%s'", response.Tasks[0].Title)
	}
	if response.TotalItems != 1 {
		t.Errorf("expected total_items=1, got %d", response.TotalItems)
	}
}
