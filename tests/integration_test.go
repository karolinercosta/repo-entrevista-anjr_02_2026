package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/tasksapi/models"
	"example.com/tasksapi/router"
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
