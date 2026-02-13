package store

import (
	"time"

	"example.com/tasksapi/models"
)

// LoggingStore é um decorator que intercepta e loga todas as operações do Store
type LoggingStore struct {
	store  Store
	logger models.Logger
}

// NewLoggingStore cria um Store com logging automático de todas as operações
func NewLoggingStore(s Store, logger models.Logger) Store {
	if logger == nil {
		logger = models.NewDefaultLogger()
	}
	return &LoggingStore{
		store:  s,
		logger: logger,
	}
}

func (l *LoggingStore) Create(t models.Task) models.Task {
	start := time.Now()
	l.logger.Info("[STORE] Creating task: title=%s, status=%s", t.Title, t.Status)

	result := l.store.Create(t)

	duration := time.Since(start)
	l.logger.Info("[STORE] Created task: id=%s, duration=%v", result.ID, duration)

	return result
}

func (l *LoggingStore) Get(id string) (models.Task, error) {
	start := time.Now()
	l.logger.Info("[STORE] Getting task: id=%s", id)

	result, err := l.store.Get(id)

	duration := time.Since(start)
	if err != nil {
		l.logger.Warn("[STORE] Failed to get task: id=%s, error=%v, duration=%v", id, err, duration)
	} else {
		l.logger.Info("[STORE] Retrieved task: id=%s, title=%s, duration=%v", id, result.Title, duration)
	}

	return result, err
}

func (l *LoggingStore) List() []models.Task {
	start := time.Now()
	l.logger.Info("[STORE] Listing all tasks")

	result := l.store.List()

	duration := time.Since(start)
	l.logger.Info("[STORE] Listed %d tasks, duration=%v", len(result), duration)

	return result
}

func (l *LoggingStore) Update(id string, patch map[string]interface{}) (models.Task, error) {
	start := time.Now()
	l.logger.Info("[STORE] Updating task: id=%s, fields=%v", id, getFieldNames(patch))

	result, err := l.store.Update(id, patch)

	duration := time.Since(start)
	if err != nil {
		l.logger.Warn("[STORE] Failed to update task: id=%s, error=%v, duration=%v", id, err, duration)
	} else {
		l.logger.Info("[STORE] Updated task: id=%s, duration=%v", id, duration)
	}

	return result, err
}

func (l *LoggingStore) Delete(id string) error {
	start := time.Now()
	l.logger.Info("[STORE] Deleting task: id=%s", id)

	err := l.store.Delete(id)

	duration := time.Since(start)
	if err != nil {
		l.logger.Warn("[STORE] Failed to delete task: id=%s, error=%v, duration=%v", id, err, duration)
	} else {
		l.logger.Info("[STORE] Deleted task: id=%s, duration=%v", id, duration)
	}

	return err
}

// getFieldNames extrai os nomes dos campos do patch para logging
func getFieldNames(patch map[string]interface{}) []string {
	fields := make([]string, 0, len(patch))
	for k := range patch {
		fields = append(fields, k)
	}
	return fields
}
