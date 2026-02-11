package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/tasksapi/handlers"
	"example.com/tasksapi/store"
)

func TestListTasksReturns200(t *testing.T) {
	s := store.New()
	h := handlers.NewAPI(s)
	r := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()

	h.ListTasks(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
