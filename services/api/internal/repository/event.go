package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *model.TaskEvent) error {
	query := `INSERT INTO ai_task_events (task_id, event_type, message, metadata) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	var metadata []byte
	if len(event.Metadata) > 0 {
		metadata = event.Metadata
	} else {
		metadata = []byte("{}")
	}
	return r.db.QueryRow(query, event.TaskID, event.EventType, event.Message, metadata).Scan(&event.ID, &event.CreatedAt)
}

func (r *EventRepository) GetByTaskID(taskID string) ([]model.TaskEvent, error) {
	query := `SELECT id, task_id, event_type, message, metadata, created_at FROM ai_task_events WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]model.TaskEvent, 0)
	for rows.Next() {
		var event model.TaskEvent
		var message sql.NullString
		var metadata []byte
		if err := rows.Scan(&event.ID, &event.TaskID, &event.EventType, &message, &metadata, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.Message = message.String
		event.Metadata = metadata
		events = append(events, event)
	}
	return events, nil
}
