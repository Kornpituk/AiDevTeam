package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentTeamRepository struct {
	db *sql.DB
}

func NewAgentTeamRepository(db *sql.DB) *AgentTeamRepository {
	return &AgentTeamRepository{db: db}
}

func (r *AgentTeamRepository) Create(team *model.AgentTeam) error {
	query := `INSERT INTO agent_teams (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	var desc sql.NullString
	if team.Description == "" {
		desc = sql.NullString{Valid: false}
	} else {
		desc = sql.NullString{String: team.Description, Valid: true}
	}
	return r.db.QueryRow(query, team.Name, desc).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *AgentTeamRepository) GetAll() ([]model.AgentTeam, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM agent_teams ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.AgentTeam
	for rows.Next() {
		var team model.AgentTeam
		var desc sql.NullString
		if err := rows.Scan(&team.ID, &team.Name, &desc, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, err
		}
		team.Description = desc.String
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *AgentTeamRepository) GetByID(id string) (*model.AgentTeam, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM agent_teams WHERE id = $1`
	var team model.AgentTeam
	var desc sql.NullString
	err := r.db.QueryRow(query, id).Scan(&team.ID, &team.Name, &desc, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		return nil, err
	}
	team.Description = desc.String
	return &team, nil
}
