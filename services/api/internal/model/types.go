package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Plan        string    `json:"plan"`
	ReviewNotes string    `json:"review_notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskEvent struct {
	ID        string          `json:"id"`
	TaskID    string          `json:"task_id"`
	EventType string          `json:"event_type"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

type TaskArtifact struct {
	ID          string          `json:"id"`
	TaskID      string          `json:"task_id"`
	Name        string          `json:"name"`
	ArtifactType string         `json:"artifact_type"`
	Content     string          `json:"content"`
	ContentType string          `json:"content_type"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &j)
}
