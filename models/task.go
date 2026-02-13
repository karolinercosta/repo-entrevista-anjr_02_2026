package models

import (
	"time"
)

type Task struct {
	ID          string     `json:"id" bson:"id,omitempty"`
	Title       string     `json:"title" bson:"title"`
	Description string     `json:"description,omitempty" bson:"description,omitempty"`
	Status      string     `json:"status" bson:"status"`
	Priority    string     `json:"priority" bson:"priority,omitempty"`
	DueDate     *Date      `json:"due_date" bson:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}

type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	TotalItems int    `json:"total_items"`
}
