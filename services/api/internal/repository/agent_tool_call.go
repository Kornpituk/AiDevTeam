package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentToolCallRepository struct {
	db *sql.DB
}

func NewAgentToolCallRepository(db *sql.DB) *AgentToolCallRepository {
	return &AgentToolCallRepository{db: db}
}

func (r *AgentToolCallRepository) Create(toolCall *model.AgentToolCall) error {
	query := `INSERT INTO agent_tool_calls (run_id, step_id, tool_name, input, output, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	var stepID sql.NullString
	var input []byte
	var output []byte

	if toolCall.StepID == "" {
		stepID = sql.NullString{Valid: false}
	} else {
		stepID = sql.NullString{String: toolCall.StepID, Valid: true}
	}
	if len(toolCall.Input) > 0 {
		input = toolCall.Input
	} else {
		input = []byte("{}")
	}
	if len(toolCall.Output) > 0 {
		output = toolCall.Output
	} else {
		output = []byte("{}")
	}
	return r.db.QueryRow(query, toolCall.RunID, stepID, toolCall.ToolName, input, output, toolCall.Status).Scan(&toolCall.ID, &toolCall.CreatedAt)
}

func (r *AgentToolCallRepository) GetByRunID(runID string) ([]model.AgentToolCall, error) {
	query := `SELECT id, run_id, step_id, tool_name, input, output, status, created_at, completed_at FROM agent_tool_calls WHERE run_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	toolCalls := make([]model.AgentToolCall, 0)
	for rows.Next() {
		var toolCall model.AgentToolCall
		var stepID sql.NullString
		var input []byte
		var output []byte
		var completedAt sql.NullTime
		if err := rows.Scan(&toolCall.ID, &toolCall.RunID, &stepID, &toolCall.ToolName, &input, &output, &toolCall.Status, &toolCall.CreatedAt, &completedAt); err != nil {
			return nil, err
		}
		toolCall.StepID = stepID.String
		toolCall.Input = input
		toolCall.Output = output
		if completedAt.Valid {
			t := completedAt.Time
			toolCall.CompletedAt = &t
		}
		toolCalls = append(toolCalls, toolCall)
	}
	return toolCalls, nil
}

func (r *AgentToolCallRepository) UpdateStatus(id string, status string) (*model.AgentToolCall, error) {
	query := `UPDATE agent_tool_calls SET status = $1, completed_at = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE NULL END WHERE id = $3 RETURNING id, run_id, step_id, tool_name, input, output, status, created_at, completed_at`
	var toolCall model.AgentToolCall
	var stepID sql.NullString
	var input []byte
	var output []byte
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, status, status, id).Scan(&toolCall.ID, &toolCall.RunID, &stepID, &toolCall.ToolName, &input, &output, &toolCall.Status, &toolCall.CreatedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	toolCall.StepID = stepID.String
	toolCall.Input = input
	toolCall.Output = output
	if completedAt.Valid {
		t := completedAt.Time
		toolCall.CompletedAt = &t
	}
	return &toolCall, nil
}

func (r *AgentToolCallRepository) UpdateOutput(id string, output []byte, status string) (*model.AgentToolCall, error) {
	query := `UPDATE agent_tool_calls SET output = $1, status = $2, completed_at = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE NULL END WHERE id = $3 RETURNING id, run_id, step_id, tool_name, input, output, status, created_at, completed_at`
	var toolCall model.AgentToolCall
	var stepID sql.NullString
	var input []byte
	var outputScan []byte
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, output, status, id).Scan(&toolCall.ID, &toolCall.RunID, &stepID, &toolCall.ToolName, &input, &outputScan, &toolCall.Status, &toolCall.CreatedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	toolCall.StepID = stepID.String
	toolCall.Input = input
	toolCall.Output = outputScan
	if completedAt.Valid {
		t := completedAt.Time
		toolCall.CompletedAt = &t
	}
	return &toolCall, nil
}

func (r *AgentToolCallRepository) GetByID(id string) (*model.AgentToolCall, error) {
	query := `SELECT id, run_id, step_id, tool_name, input, output, status, created_at, completed_at FROM agent_tool_calls WHERE id = $1`
	var toolCall model.AgentToolCall
	var stepID sql.NullString
	var input []byte
	var output []byte
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(&toolCall.ID, &toolCall.RunID, &stepID, &toolCall.ToolName, &input, &output, &toolCall.Status, &toolCall.CreatedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	toolCall.StepID = stepID.String
	toolCall.Input = input
	toolCall.Output = output
	if completedAt.Valid {
		t := completedAt.Time
		toolCall.CompletedAt = &t
	}
	return &toolCall, nil
}
