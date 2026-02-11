package tests

import (
	"testing"

	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

func TestInMemoryStoreCreateAndGet(t *testing.T) {
	s := store.New()
	task := models.Task{
		Title:  "Test",
		Status: "pending",
	}
	created := s.Create(task)
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Title != "Test" {
		t.Errorf("expected title Test, got %s", got.Title)
	}
}

func TestInMemoryStoreDelete(t *testing.T) {
	s := store.New()
	task := models.Task{
		Title:  "DeleteMe",
		Status: "pending",
	}
	created := s.Create(task)
	err := s.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = s.Get(created.ID)
	if err == nil {
		t.Errorf("expected error after delete, got nil")
	}
}
