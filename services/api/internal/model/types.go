package model

import (
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskEvent struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	EventType string    `json:"event_type"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskArtifact struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}