package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/tasksapi/handlers"
	"example.com/tasksapi/models"
	"example.com/tasksapi/router"
	"example.com/tasksapi/store"
	"github.com/gorilla/mux"
)

func TestIntegrationUpdateCompletedTask(t *testing.T) {
	r := router.NewWithLogger(&models.NoOpLogger{})

	// Create a task
	createJSON := `{"title":"Task to Complete","status":"pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(createJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create task: %d", w.Code)
	}

	var created models.Task
	json.NewDecoder(w.Body).Decode(&created)

	// Update to completed
	updateToCompleted := `{"status":"completed"}`
	req = httptest.NewRequest("PUT", "/tasks/"+created.ID, bytes.NewBufferString(updateToCompleted))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to update task to completed: %d", w.Code)
	}

	// Try to update completed task - should fail with 409
	updateCompleted := `{"title":"Try to Update"}`
	req = httptest.NewRequest("PUT", "/tasks/"+created.ID, bytes.NewBufferString(updateCompleted))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for business rule violation, got %d", w.Code)
	}

	var errResp models.APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Code != 409 {
		t.Errorf("expected error code 409, got %d", errResp.Code)
	}
	if errResp.Message != "completed tasks cannot be edited" {
		t.Errorf("unexpected error message: %s", errResp.Message)
	}
}

func TestIntegrationCreateTaskInvalidStatus(t *testing.T) {
	r := router.NewWithLogger(&models.NoOpLogger{})

	createJSON := `{"title":"Test","status":"invalid_status"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(createJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d", w.Code)
	}
}

func TestIntegrationFullCRUD(t *testing.T) {
	r := router.NewWithLogger(&models.NoOpLogger{})

	// Create
	createJSON := `{"title":"Full CRUD Test","status":"pending","priority":"high"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(createJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", w.Code)
	}

	var task models.Task
	json.NewDecoder(w.Body).Decode(&task)

	// Read
	req = httptest.NewRequest("GET", "/tasks/"+task.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get failed: %d", w.Code)
	}

	// Update
	updateJSON := `{"title":"Updated Title"}`
	req = httptest.NewRequest("PUT", "/tasks/"+task.ID, bytes.NewBufferString(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update failed: %d", w.Code)
	}

	var updated models.Task
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/tasks/"+task.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete failed: %d", w.Code)
	}

	// Verify deleted
	req = httptest.NewRequest("GET", "/tasks/"+task.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

func TestIntegrationTaskWithDueDate(t *testing.T) {
	// Create isolated store for this test
	s := store.New()
	logger := &models.NoOpLogger{}
	api := handlers.NewAPI(s, logger)

	r := mux.NewRouter()
	r.HandleFunc("/tasks", api.CreateTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", api.GetTask).Methods("GET")
	r.HandleFunc("/tasks/{id}", api.UpdateTask).Methods("PUT")

	// Create task with due_date
	createJSON := `{"title":"Task with Due Date","status":"pending","due_date":"2026-12-31"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(createJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", w.Code)
	}

	var task models.Task
	json.NewDecoder(w.Body).Decode(&task)

	// Verify due_date is set correctly
	if task.DueDate == nil {
		t.Fatal("expected due_date to be set")
	}
	if task.DueDate.String() != "2026-12-31" {
		t.Errorf("expected due_date '2026-12-31', got '%s'", task.DueDate.String())
	}

	// Get task and verify date is returned
	req = httptest.NewRequest("GET", "/tasks/"+task.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var retrieved models.Task
	json.NewDecoder(w.Body).Decode(&retrieved)
	if retrieved.DueDate == nil {
		t.Fatal("expected due_date in retrieved task")
	}
	if retrieved.DueDate.String() != "2026-12-31" {
		t.Errorf("expected retrieved due_date '2026-12-31', got '%s'", retrieved.DueDate.String())
	}

	// Update due_date
	updateJSON := `{"due_date":"2027-01-15"}`
	req = httptest.NewRequest("PUT", "/tasks/"+task.ID, bytes.NewBufferString(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update failed: %d", w.Code)
	}

	var updated models.Task
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.DueDate == nil {
		t.Fatal("expected due_date in updated task")
	}
	if updated.DueDate.String() != "2027-01-15" {
		t.Errorf("expected updated due_date '2027-01-15', got '%s'", updated.DueDate.String())
	}
}

func TestIntegrationFilterByDueDateAndPriority(t *testing.T) {
	// Create isolated store for this test
	s := store.New()
	logger := &models.NoOpLogger{}
	api := handlers.NewAPI(s, logger)

	r := mux.NewRouter()
	r.HandleFunc("/tasks", api.CreateTask).Methods("POST")
	r.HandleFunc("/tasks", api.ListTasks).Methods("GET")

	// Create tasks with different combinations
	tasks := []string{
		`{"title":"Filter Task 1","status":"pending","priority":"high","due_date":"2026-12-31"}`,
		`{"title":"Filter Task 2","status":"pending","priority":"low","due_date":"2026-12-31"}`,
		`{"title":"Filter Task 3","status":"pending","priority":"high"}`,
		`{"title":"Filter Task 4","status":"pending"}`,
	}

	for _, taskJSON := range tasks {
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create task: %d", w.Code)
		}
	}

	// Filter by due_date
	req := httptest.NewRequest("GET", "/tasks?due_date=2026-12-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response models.TaskListResponse
	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 2 {
		t.Errorf("expected 2 tasks with due_date, got %d", len(response.Tasks))
	}
	if response.TotalItems != 2 {
		t.Errorf("expected total_items=2, got %d", response.TotalItems)
	}

	// Filter by null due_date
	req = httptest.NewRequest("GET", "/tasks?due_date=null", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 2 {
		t.Errorf("expected 2 tasks without due_date, got %d", len(response.Tasks))
	}

	// Filter by priority and due_date
	req = httptest.NewRequest("GET", "/tasks?priority=high&due_date=2026-12-31", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&response)
	if len(response.Tasks) != 1 {
		t.Errorf("expected 1 task with high priority and due_date, got %d", len(response.Tasks))
	}
	if len(response.Tasks) > 0 && response.Tasks[0].Title != "Filter Task 1" {
		t.Errorf("expected 'Filter Task 1', got '%s'", response.Tasks[0].Title)
	}
	if response.TotalItems != 1 {
		t.Errorf("expected total_items=1, got %d", response.TotalItems)
	}
}
