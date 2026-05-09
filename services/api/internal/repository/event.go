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
	query := `INSERT INTO ai_task_events (task_id, event_type, data) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(query, event.TaskID, event.EventType, event.Data).Scan(&event.ID, &event.CreatedAt)
}

func (r *EventRepository) GetByTaskID(taskID int) ([]model.TaskEvent, error) {
	query := `SELECT id, task_id, event_type, data, created_at FROM ai_task_events WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.TaskEvent
	for rows.Next() {
		var event model.TaskEvent
		if err := rows.Scan(&event.ID, &event.TaskID, &event.EventType, &event.Data, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}