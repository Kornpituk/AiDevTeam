package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentRunStepRepository struct {
	db *sql.DB
}

func NewAgentRunStepRepository(db *sql.DB) *AgentRunStepRepository {
	return &AgentRunStepRepository{db: db}
}

func (r *AgentRunStepRepository) Create(step *model.AgentRunStep) error {
	query := `INSERT INTO agent_run_steps (run_id, profile_id, step_type, status, title, instructions, output, position) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	var profileID sql.NullString
	var instructions sql.NullString
	var output sql.NullString
	if step.ProfileID == "" {
		profileID = sql.NullString{Valid: false}
	} else {
		profileID = sql.NullString{String: step.ProfileID, Valid: true}
	}
	if step.Instructions == "" {
		instructions = sql.NullString{Valid: false}
	} else {
		instructions = sql.NullString{String: step.Instructions, Valid: true}
	}
	if step.Output == "" {
		output = sql.NullString{Valid: false}
	} else {
		output = sql.NullString{String: step.Output, Valid: true}
	}
	return r.db.QueryRow(query, step.RunID, profileID, step.StepType, step.Status, step.Title, instructions, output, step.Position).Scan(&step.ID, &step.CreatedAt, &step.UpdatedAt)
}

func (r *AgentRunStepRepository) GetByRunID(runID string) ([]model.AgentRunStep, error) {
	query := `SELECT id, run_id, profile_id, step_type, status, title, instructions, output, position, started_at, completed_at, created_at, updated_at FROM agent_run_steps WHERE run_id = $1 ORDER BY position ASC, created_at ASC`
	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.AgentRunStep
	for rows.Next() {
		var step model.AgentRunStep
		var profileID sql.NullString
		var instructions sql.NullString
		var output sql.NullString
		var startedAt sql.NullTime
		var completedAt sql.NullTime
		if err := rows.Scan(&step.ID, &step.RunID, &profileID, &step.StepType, &step.Status, &step.Title, &instructions, &output, &step.Position, &startedAt, &completedAt, &step.CreatedAt, &step.UpdatedAt); err != nil {
			return nil, err
		}
		step.ProfileID = profileID.String
		step.Instructions = instructions.String
		step.Output = output.String
		if startedAt.Valid {
			t := startedAt.Time
			step.StartedAt = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			step.CompletedAt = &t
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (r *AgentRunStepRepository) UpdateStatus(id string, status string) (*model.AgentRunStep, error) {
	query := `UPDATE agent_run_steps SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, run_id, profile_id, step_type, status, title, instructions, output, position, started_at, completed_at, created_at, updated_at`
	var step model.AgentRunStep
	var profileID sql.NullString
	var instructions sql.NullString
	var output sql.NullString
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, status, id).Scan(&step.ID, &step.RunID, &profileID, &step.StepType, &step.Status, &step.Title, &instructions, &output, &step.Position, &startedAt, &completedAt, &step.CreatedAt, &step.UpdatedAt)
	if err != nil {
		return nil, err
	}
	step.ProfileID = profileID.String
	step.Instructions = instructions.String
	step.Output = output.String
	if startedAt.Valid {
		t := startedAt.Time
		step.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		step.CompletedAt = &t
	}
	return &step, nil
}

func (r *AgentRunStepRepository) GetByID(id string) (*model.AgentRunStep, error) {
	query := `SELECT id, run_id, profile_id, step_type, status, title, instructions, output, position, started_at, completed_at, created_at, updated_at FROM agent_run_steps WHERE id = $1`
	var step model.AgentRunStep
	var profileID sql.NullString
	var instructions sql.NullString
	var output sql.NullString
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(&step.ID, &step.RunID, &profileID, &step.StepType, &step.Status, &step.Title, &instructions, &output, &step.Position, &startedAt, &completedAt, &step.CreatedAt, &step.UpdatedAt)
	if err != nil {
		return nil, err
	}
	step.ProfileID = profileID.String
	step.Instructions = instructions.String
	step.Output = output.String
	if startedAt.Valid {
		t := startedAt.Time
		step.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		step.CompletedAt = &t
	}
	return &step, nil
}
