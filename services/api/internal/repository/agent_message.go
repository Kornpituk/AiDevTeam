package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentMessageRepository struct {
	db *sql.DB
}

func NewAgentMessageRepository(db *sql.DB) *AgentMessageRepository {
	return &AgentMessageRepository{db: db}
}

func (r *AgentMessageRepository) Create(message *model.AgentMessage) error {
	query := `INSERT INTO agent_messages (run_id, step_id, profile_id, role, content, metadata) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	var stepID sql.NullString
	var profileID sql.NullString
	var metadata []byte
	if message.StepID == "" {
		stepID = sql.NullString{Valid: false}
	} else {
		stepID = sql.NullString{String: message.StepID, Valid: true}
	}
	if message.ProfileID == "" {
		profileID = sql.NullString{Valid: false}
	} else {
		profileID = sql.NullString{String: message.ProfileID, Valid: true}
	}
	if len(message.Metadata) > 0 {
		metadata = message.Metadata
	} else {
		metadata = []byte("{}")
	}
	return r.db.QueryRow(query, message.RunID, stepID, profileID, message.Role, message.Content, metadata).Scan(&message.ID, &message.CreatedAt)
}

func (r *AgentMessageRepository) GetByRunID(runID string) ([]model.AgentMessage, error) {
	query := `SELECT id, run_id, step_id, profile_id, role, content, metadata, created_at FROM agent_messages WHERE run_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.AgentMessage, 0)
	for rows.Next() {
		var message model.AgentMessage
		var stepID sql.NullString
		var profileID sql.NullString
		var metadata []byte
		if err := rows.Scan(&message.ID, &message.RunID, &stepID, &profileID, &message.Role, &message.Content, &metadata, &message.CreatedAt); err != nil {
			return nil, err
		}
		message.StepID = stepID.String
		message.ProfileID = profileID.String
		message.Metadata = metadata
		messages = append(messages, message)
	}
	return messages, nil
}
