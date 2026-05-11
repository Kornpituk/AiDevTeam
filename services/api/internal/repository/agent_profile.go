package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentProfileRepository struct {
	db *sql.DB
}

func NewAgentProfileRepository(db *sql.DB) *AgentProfileRepository {
	return &AgentProfileRepository{db: db}
}

func (r *AgentProfileRepository) Create(profile *model.AgentProfile) error {
	query := `INSERT INTO agent_profiles (name, role, description, system_prompt, default_model) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	var desc sql.NullString
	var sysPrompt sql.NullString
	var defaultModel sql.NullString
	if profile.Description == "" {
		desc = sql.NullString{Valid: false}
	} else {
		desc = sql.NullString{String: profile.Description, Valid: true}
	}
	if profile.SystemPrompt == "" {
		sysPrompt = sql.NullString{Valid: false}
	} else {
		sysPrompt = sql.NullString{String: profile.SystemPrompt, Valid: true}
	}
	if profile.DefaultModel == "" {
		defaultModel = sql.NullString{Valid: false}
	} else {
		defaultModel = sql.NullString{String: profile.DefaultModel, Valid: true}
	}
	return r.db.QueryRow(query, profile.Name, profile.Role, desc, sysPrompt, defaultModel).Scan(&profile.ID, &profile.CreatedAt, &profile.UpdatedAt)
}

func (r *AgentProfileRepository) GetAll() ([]model.AgentProfile, error) {
	query := `SELECT id, name, role, description, system_prompt, default_model, created_at, updated_at FROM agent_profiles ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]model.AgentProfile, 0)
	for rows.Next() {
		var profile model.AgentProfile
		var desc sql.NullString
		var sysPrompt sql.NullString
		var defaultModel sql.NullString
		if err := rows.Scan(&profile.ID, &profile.Name, &profile.Role, &desc, &sysPrompt, &defaultModel, &profile.CreatedAt, &profile.UpdatedAt); err != nil {
			return nil, err
		}
		profile.Description = desc.String
		profile.SystemPrompt = sysPrompt.String
		profile.DefaultModel = defaultModel.String
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (r *AgentProfileRepository) GetByID(id string) (*model.AgentProfile, error) {
	query := `SELECT id, name, role, description, system_prompt, default_model, created_at, updated_at FROM agent_profiles WHERE id = $1`
	var profile model.AgentProfile
	var desc sql.NullString
	var sysPrompt sql.NullString
	var defaultModel sql.NullString
	err := r.db.QueryRow(query, id).Scan(&profile.ID, &profile.Name, &profile.Role, &desc, &sysPrompt, &defaultModel, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return nil, err
	}
	profile.Description = desc.String
	profile.SystemPrompt = sysPrompt.String
	profile.DefaultModel = defaultModel.String
	return &profile, nil
}
