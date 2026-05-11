package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentRunRepository struct {
	db *sql.DB
}

func NewAgentRunRepository(db *sql.DB) *AgentRunRepository {
	return &AgentRunRepository{db: db}
}

func (r *AgentRunRepository) Create(run *model.AgentRun) error {
	query := `INSERT INTO agent_runs (task_id, team_id, status, goal, summary) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	var teamID sql.NullString
	var goal sql.NullString
	var summary sql.NullString
	if run.TeamID == "" {
		teamID = sql.NullString{Valid: false}
	} else {
		teamID = sql.NullString{String: run.TeamID, Valid: true}
	}
	if run.Goal == "" {
		goal = sql.NullString{Valid: false}
	} else {
		goal = sql.NullString{String: run.Goal, Valid: true}
	}
	if run.Summary == "" {
		summary = sql.NullString{Valid: false}
	} else {
		summary = sql.NullString{String: run.Summary, Valid: true}
	}
	return r.db.QueryRow(query, run.TaskID, teamID, run.Status, goal, summary).Scan(&run.ID, &run.CreatedAt, &run.UpdatedAt)
}

func (r *AgentRunRepository) GetByTaskID(taskID string) ([]model.AgentRun, error) {
	query := `SELECT id, task_id, team_id, status, goal, summary, created_at, updated_at FROM agent_runs WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []model.AgentRun
	for rows.Next() {
		var run model.AgentRun
		var teamID sql.NullString
		var goal sql.NullString
		var summary sql.NullString
		if err := rows.Scan(&run.ID, &run.TaskID, &teamID, &run.Status, &goal, &summary, &run.CreatedAt, &run.UpdatedAt); err != nil {
			return nil, err
		}
		run.TeamID = teamID.String
		run.Goal = goal.String
		run.Summary = summary.String
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *AgentRunRepository) GetByID(id string) (*model.AgentRun, error) {
	query := `SELECT id, task_id, team_id, status, goal, summary, created_at, updated_at FROM agent_runs WHERE id = $1`
	var run model.AgentRun
	var teamID sql.NullString
	var goal sql.NullString
	var summary sql.NullString
	err := r.db.QueryRow(query, id).Scan(&run.ID, &run.TaskID, &teamID, &run.Status, &goal, &summary, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		return nil, err
	}
	run.TeamID = teamID.String
	run.Goal = goal.String
	run.Summary = summary.String
	return &run, nil
}

func (r *AgentRunRepository) UpdateStatus(id string, status string) (*model.AgentRun, error) {
	query := `UPDATE agent_runs SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, task_id, team_id, status, goal, summary, created_at, updated_at`
	var run model.AgentRun
	var teamID sql.NullString
	var goal sql.NullString
	var summary sql.NullString
	err := r.db.QueryRow(query, status, id).Scan(&run.ID, &run.TaskID, &teamID, &run.Status, &goal, &summary, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		return nil, err
	}
	run.TeamID = teamID.String
	run.Goal = goal.String
	run.Summary = summary.String
	return &run, nil
}

func (r *AgentRunRepository) UpdateSummary(id string, summary string) (*model.AgentRun, error) {
	query := `UPDATE agent_runs SET summary = $1, updated_at = NOW() WHERE id = $2 RETURNING id, task_id, team_id, status, goal, summary, created_at, updated_at`
	var run model.AgentRun
	var teamID sql.NullString
	var goal sql.NullString
	var summaryResult sql.NullString
	err := r.db.QueryRow(query, summary, id).Scan(&run.ID, &run.TaskID, &teamID, &run.Status, &goal, &summaryResult, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		return nil, err
	}
	run.TeamID = teamID.String
	run.Goal = goal.String
	run.Summary = summaryResult.String
	return &run, nil
}
