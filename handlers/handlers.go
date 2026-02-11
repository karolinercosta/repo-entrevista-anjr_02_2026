package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
			http.Error(w, "invalid request", http.StatusBadRequest)
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
				http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			t.DueDate = &parsed
		}
		if v := r.FormValue("completed"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				t.Completed = b
			}
		}
	}
	if t.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if t.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}
	if !models.IsValidStatus(t.Status) {
		http.Error(w, "invalid status, allowed: pending, in_progress, completed, cancelled", http.StatusBadRequest)
		return
	}
	if t.Priority != "" && !models.IsValidPriority(t.Priority) {
		http.Error(w, "invalid priority, allowed: low, medium, high", http.StatusBadRequest)
		return
	}

	// Só valida se houver due date
	if t.DueDate != nil {
		if !models.IsValidDate(*t.DueDate) {
			http.Error(w, "date should be in the future", http.StatusBadRequest)
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
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (a *API) GetTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	t, err := a.store.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) UpdateTask(w http.ResponseWriter, r *http.Request) {
	log.Println("aio")
	id := mux.Vars(r)["id"]
	var patch map[string]interface{}
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
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
				http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			patch["due_date"] = parsed
		}
		if v := r.FormValue("completed"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				patch["completed"] = b
			}
		}
	}
	log.Println("UpdateTask: id=%s incoming patch=%#v", id, patch)

	// validate and normalize provided fields
	if len(patch) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	allowed := map[string]struct{}{"title": {}, "description": {}, "status": {}, "priority": {}, "completed": {}, "due_date": {}}
	for k, v := range patch {
		if _, ok := allowed[k]; !ok {
			http.Error(w, "unknown field: "+k, http.StatusBadRequest)
			return
		}
		switch k {
		case "status":
			s, ok := v.(string)
			if !ok || !models.IsValidStatus(s) {
				http.Error(w, "invalid status, allowed: pending, in_progress, completed, cancelled", http.StatusBadRequest)
				return
			}
			patch[k] = s

		case "priority":
			s, ok := v.(string)
			if !ok || !models.IsValidPriority(s) {
				http.Error(w, "invalid priority, allowed: low, medium, high", http.StatusBadRequest)
				return
			}
			patch[k] = s
		case "completed":
			if b, ok := v.(bool); ok {
				patch[k] = b
			} else if s, ok := v.(string); ok {
				if parsed, err := strconv.ParseBool(s); err == nil {
					patch[k] = parsed
				} else {
					http.Error(w, "invalid completed value", http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, "invalid completed value", http.StatusBadRequest)
				return
			}
		case "due_date":
			switch vv := v.(type) {
			case string:
				parsed, err := models.ParseDate(vv)
				if err != nil {
					http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
					return
				}
				if !models.IsValidDate(parsed) {
					http.Error(w, "date should be in the future", http.StatusBadRequest)
					return
				}
				patch[k] = parsed
			case time.Time:
				if !models.IsValidDate(vv) {
					http.Error(w, "date should be in the future", http.StatusBadRequest)
					return
				}
				patch[k] = vv
			default:
				http.Error(w, "due_date must be a YYYY-MM-DD string or date", http.StatusBadRequest)
				return
			}
		case "title", "description":
			if _, ok := v.(string); !ok {
				http.Error(w, k+" must be a string", http.StatusBadRequest)
				return
			}
		}
	}
	t, err := a.store.Update(id, patch)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := a.store.Delete(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
