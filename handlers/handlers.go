package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

type API struct {
	store store.Store
}

func NewAPI(s store.Store) *API {
	return &API{store: s}
}

func (a *API) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request")
			return
		}
	} else {
		// suporte para multipart/form-data e x-www-form-urlencoded
		if strings.HasPrefix(ct, "multipart/form-data") {
			_ = r.ParseMultipartForm(10 << 20)
		} else {
			_ = r.ParseForm()
		}
		t.Title = r.FormValue("title")
		t.Description = r.FormValue("description")
		t.Status = r.FormValue("status")
		t.Priority = r.FormValue("priority")
		if v := r.FormValue("due_date"); v != "" {
			parsed, err := models.ParseDate(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
				return
			}
			t.DueDate = &parsed
		}

	}
	if t.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "title is required")
		return
	}
	if t.Status == "" {
		writeJSONError(w, http.StatusBadRequest, "status is required")
		return
	}
	if !models.IsValidStatus(t.Status) {
		writeJSONError(w, http.StatusBadRequest, "invalid status, allowed: pending, in_progress, completed, cancelled")
		return
	}
	if t.Priority != "" && !models.IsValidPriority(t.Priority) {
		writeJSONError(w, http.StatusBadRequest, "invalid priority, allowed: low, medium, high")
		return
	}
	if t.DueDate != nil {
		if !models.IsValidDate(*t.DueDate) {
			writeJSONError(w, http.StatusBadRequest, "date should be in the future")
			return
		}
	}
	created := a.store.Create(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)

}

func (a *API) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := a.store.List()

	// Apply query parameter filters
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	filtered := make([]models.Task, 0)
	for _, task := range tasks {
		// Filtros
		if status != "" && task.Status != status {
			continue
		}
		if priority != "" && task.Priority != priority {
			continue
		}

		filtered = append(filtered, task)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filtered)
}

func (a *API) GetTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	t, err := a.store.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	task, err := a.store.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if models.IsCompletedTask(task.Status) {
		writeJSONError(w, http.StatusConflict, "completed tasks cannot be edited")
		return
	}
	var patch map[string]interface{}
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request")
			return
		}
	} else {
		// handle form-data / urlencoded
		if strings.HasPrefix(ct, "multipart/form-data") {
			_ = r.ParseMultipartForm(10 << 20)
		} else {
			_ = r.ParseForm()
		}
		patch = map[string]interface{}{}
		if v := r.FormValue("title"); v != "" {
			patch["title"] = v
		}
		if v := r.FormValue("description"); v != "" {
			patch["description"] = v
		}
		if v := r.FormValue("status"); v != "" {
			patch["status"] = v
		}
		if v := r.FormValue("priority"); v != "" {
			patch["priority"] = v
		}
		if v := r.FormValue("due_date"); v != "" {
			parsed, err := models.ParseDate(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
				return
			}
			patch["due_date"] = parsed
		}

	}
	log.Println("UpdateTask: id=%s incoming patch=%#v", id, patch)
	if len(patch) == 0 {
		writeJSONError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	allowed := map[string]struct{}{"title": {}, "description": {}, "status": {}, "priority": {}, "due_date": {}}
	for k, v := range patch {
		if _, ok := allowed[k]; !ok {
			writeJSONError(w, http.StatusBadRequest, "unknown field: "+k)
			return
		}
		switch k {
		case "status":
			s, ok := v.(string)
			if !ok || !models.IsValidStatus(s) {
				writeJSONError(w, http.StatusBadRequest, "invalid status, allowed: pending, in_progress, completed, cancelled")
				return
			}
			patch[k] = s

		case "priority":
			s, ok := v.(string)
			if !ok || !models.IsValidPriority(s) {
				writeJSONError(w, http.StatusBadRequest, "invalid priority, allowed: low, medium, high")
				return
			}
			patch[k] = s
		case "due_date":
			switch vv := v.(type) {
			case string:
				parsed, err := models.ParseDate(vv)
				if err != nil {
					writeJSONError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
					return
				}
				if !models.IsValidDate(parsed) {
					writeJSONError(w, http.StatusBadRequest, "date should be in the future")
					return
				}
				patch[k] = parsed
			case time.Time:
				if !models.IsValidDate(vv) {
					writeJSONError(w, http.StatusBadRequest, "date should be in the future")
					return
				}
				patch[k] = vv
			default:
				writeJSONError(w, http.StatusBadRequest, "due_date must be a YYYY-MM-DD string or date")
				return
			}
		case "title", "description":
			if _, ok := v.(string); !ok {
				writeJSONError(w, http.StatusBadRequest, k+" must be a string")
				return
			}
		}
	}
	t, err := a.store.Update(id, patch)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := a.store.Delete(id); err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": msg,
		"code":  status,
	})
}
