package store

import (
	"errors"
	"sync"
	"time"

	"example.com/tasksapi/models"
	"github.com/google/uuid"
)

const queryTimeout = 10 * time.Second

var ErrNotFound = errors.New("task not found")

// Store interface define o contrato do armazenamento das Tasks
type Store interface {
	Create(t models.Task) models.Task
	List() []models.Task
	Get(id string) (models.Task, error)
	Update(id string, patch map[string]interface{}) (models.Task, error)
	Delete(id string) error
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]models.Task
}

func New() Store {
	return &InMemoryStore{items: make(map[string]models.Task)}
}

func (s *InMemoryStore) Create(t models.Task) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New().String()
	t.ID = id

	// Leave DueDate as nil if not provided
	if t.DueDate != nil && t.DueDate.IsZero() {
		t.DueDate = nil
	}
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = nil
	s.items[id] = t
	return t
}

func (s *InMemoryStore) List() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Task, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	return out
}

func (s *InMemoryStore) Get(id string) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func (s *InMemoryStore) Update(id string, patch map[string]interface{}) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	if v, ok := patch["title"]; ok {
		if s, ok := v.(string); ok {
			t.Title = s
		}
	}
	if v, ok := patch["description"]; ok {
		if s, ok := v.(string); ok {
			t.Description = s
		}
	}
	if v, ok := patch["status"]; ok {
		if s2, ok := v.(string); ok {
			t.Status = s2
		}
	}
	if v, ok := patch["priority"]; ok {
		if s2, ok := v.(string); ok {
			t.Priority = s2
		}
	}
	if v, ok := patch["completed"]; ok {
		if b, ok := v.(bool); ok {
			t.Completed = b
		}
	}
	if v, ok := patch["due_date"]; ok {
		switch vv := v.(type) {
		case time.Time:
			t.DueDate = &vv
		case *time.Time:
			if vv != nil {
				t.DueDate = vv
			}
		case string:
			if parsed, err := models.ParseDate(vv); err == nil {
				t.DueDate = &parsed
			}
		}
	}
	now := time.Now().UTC()
	t.UpdatedAt = &now
	s.items[id] = t
	return t, nil
}

func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
