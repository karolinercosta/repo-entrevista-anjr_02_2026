package models

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

// Logger interface for dependency injection
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// DefaultLogger implements Logger using standard log package
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *DefaultLogger) Fatal(msg string, args ...interface{}) {
	log.Fatalf("[FATAL] "+msg, args...)
}

// NewDefaultLogger creates a logger using standard log package
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

type NoOpLogger struct{}

func (l *NoOpLogger) Info(msg string, args ...interface{})  {}
func (l *NoOpLogger) Warn(msg string, args ...interface{})  {}
func (l *NoOpLogger) Error(msg string, args ...interface{}) {}
func (l *NoOpLogger) Fatal(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}

type TaskService struct {
	logger Logger
}

func NewTaskService(logger Logger) *TaskService {
	if logger == nil {
		logger = NewDefaultLogger()
	}
	return &TaskService{logger: logger}
}

type BusinessRule func(task Task, patch map[string]interface{}) error

var updateBusinessRules = []BusinessRule{
	PreventCompletedTaskEdits,
}

func PreventCompletedTaskEdits(task Task, patch map[string]interface{}) error {
	if IsCompletedTask(task.Status) {
		return NewBusinessRuleError("completed tasks cannot be edited")
	}
	return nil
}

func AddUpdateRule(rule BusinessRule) {
	updateBusinessRules = append(updateBusinessRules, rule)
}

func (s *TaskService) ValidateCreate(t Task) error {
	// Required fields
	if t.Title == "" {
		return NewValidationError("title is required")
	}
	if t.Status == "" {
		return NewValidationError("status is required")
	}

	// fieldValidators na validação
	patch := make(map[string]interface{})

	if err := fieldValidators["title"](t.Title, patch, "title"); err != nil {
		return err
	}
	if err := fieldValidators["status"](t.Status, patch, "status"); err != nil {
		return err
	}
	if t.Priority != "" {
		if err := fieldValidators["priority"](t.Priority, patch, "priority"); err != nil {
			return err
		}
	}
	if t.DueDate != nil {
		if err := fieldValidators["due_date"](*t.DueDate, patch, "due_date"); err != nil {
			return err
		}
	}

	return nil
}

// FieldValidator valida o valor do campo e transforma
type FieldValidator func(value interface{}, patch map[string]interface{}, fieldName string) error

// Field validators for each updatable field
var fieldValidators = map[string]FieldValidator{
	"status":      ValidateStatusField,
	"priority":    ValidatePriorityField,
	"due_date":    ValidateDueDateField,
	"title":       ValidateTitleField,
	"description": ValidateStringField,
}

func ValidateStatusField(value interface{}, patch map[string]interface{}, fieldName string) error {
	s, ok := value.(string)
	if !ok || !IsValidStatus(s) {
		return NewValidationError("invalid status, allowed: pending, in_progress, completed, cancelled")
	}
	patch[fieldName] = s
	return nil
}

func ValidatePriorityField(value interface{}, patch map[string]interface{}, fieldName string) error {
	s, ok := value.(string)
	if !ok || !IsValidPriority(s) {
		return NewValidationError("invalid priority, allowed: low, medium, high")
	}
	patch[fieldName] = s
	return nil
}

func ValidateDueDateField(value interface{}, patch map[string]interface{}, fieldName string) error {
	switch vv := value.(type) {
	case string:
		parsed, err := ParseDateOnly(vv)
		if err != nil {
			return NewValidationError("invalid date format, expected YYYY-MM-DD")
		}
		if !IsValidDate(parsed) {
			return NewValidationError("date should be in the future")
		}
		patch[fieldName] = parsed
	case Date:
		if !IsValidDate(vv) {
			return NewValidationError("date should be in the future")
		}
		patch[fieldName] = vv
	default:
		return NewValidationError("due_date must be a YYYY-MM-DD string or date")
	}
	return nil
}

func ValidateStringField(value interface{}, patch map[string]interface{}, fieldName string) error {
	if _, ok := value.(string); !ok {
		return NewValidationError(fieldName + " must be a string")
	}
	return nil
}
func ValidateTitleField(value interface{}, patch map[string]interface{}, fieldName string) error {
	s, ok := value.(string)
	if !ok {
		return NewValidationError(fieldName + " must be a string")
	}
	if !IsValidTitle(s) {
		return NewValidationError("invalid title length, it should be between 3 and 100")
	}
	patch[fieldName] = s
	return nil
}

func (s *TaskService) ValidateUpdate(task Task, patch map[string]interface{}) error {
	// Apply business rules
	for _, rule := range updateBusinessRules {
		if err := rule(task, patch); err != nil {
			return err
		}
	}

	if len(patch) == 0 {
		return NewValidationError("no fields to update")
	}

	for fieldName, value := range patch {
		if _, ok := allowedUpdateFields[fieldName]; !ok {
			return NewValidationError("unknown field: " + fieldName)
		}

		// Use field validator if available
		if validator, ok := fieldValidators[fieldName]; ok {
			if err := validator(value, patch, fieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

func getUpdateableFields() map[string]struct{} {
	fields := make(map[string]struct{})
	t := reflect.TypeOf(Task{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		// Extract field name from json tag (before comma)
		fieldName := strings.Split(jsonTag, ",")[0]
		// Skip system/read-only fields
		if fieldName != "id" && fieldName != "created_at" && fieldName != "updated_at" {
			fields[fieldName] = struct{}{}
		}
	}
	return fields
}

var allowedUpdateFields = getUpdateableFields()

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

func NewValidationError(msg string) error   { return &APIError{Code: 400, Message: msg} }
func NewBusinessRuleError(msg string) error { return &APIError{Code: 409, Message: msg} }

func WriteError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = &APIError{Code: statusCode, Message: err.Error()}
	}
	w.WriteHeader(apiErr.Code)
	_ = json.NewEncoder(w).Encode(apiErr)
}

func HandleError(w http.ResponseWriter, err error, statusCode int) bool {
	if err != nil {
		WriteError(w, err, statusCode)
		return true
	}
	return false
}
