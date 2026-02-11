package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

type API struct {
	store   store.Store
	service *models.TaskService
}

func NewAPI(s store.Store) *API {
	return &API{store: s, service: models.NewTaskService()}
}

func (a *API) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if models.HandleError(w, json.NewDecoder(r.Body).Decode(&t), http.StatusBadRequest) {
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
			if models.HandleError(w, err, http.StatusBadRequest) {
				return
			}
			t.DueDate = &parsed
		}

	}

	if models.HandleError(w, a.service.ValidateCreate(t), http.StatusBadRequest) {
		return
	}

	created := a.store.Create(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)

}

func (a *API) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := a.store.List()
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
	if models.HandleError(w, err, http.StatusNotFound) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	task, err := a.store.Get(id)
	if models.HandleError(w, err, http.StatusNotFound) {
		return
	}

	var patch map[string]interface{}
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if models.HandleError(w, json.NewDecoder(r.Body).Decode(&patch), http.StatusBadRequest) {
			return
		}
	} else {
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
			if models.HandleError(w, err, http.StatusBadRequest) {
				return
			}
			patch["due_date"] = parsed
		}

	}
	if models.HandleError(w, a.service.ValidateUpdate(task, patch), http.StatusBadRequest) {
		return
	}

	t, err := a.store.Update(id, patch)
	if models.HandleError(w, err, http.StatusNotFound) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (a *API) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if models.HandleError(w, a.store.Delete(id), http.StatusNotFound) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
