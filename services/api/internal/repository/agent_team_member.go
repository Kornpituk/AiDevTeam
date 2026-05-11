package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentTeamMemberRepository struct {
	db *sql.DB
}

func NewAgentTeamMemberRepository(db *sql.DB) *AgentTeamMemberRepository {
	return &AgentTeamMemberRepository{db: db}
}

func (r *AgentTeamMemberRepository) Create(member *model.AgentTeamMember) error {
	query := `INSERT INTO agent_team_members (team_id, profile_id, member_role, position) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, member.TeamID, member.ProfileID, member.MemberRole, member.Position).Scan(&member.ID, &member.CreatedAt)
}

func (r *AgentTeamMemberRepository) GetByTeamID(teamID string) ([]model.AgentTeamMember, error) {
	query := `SELECT id, team_id, profile_id, member_role, position, created_at FROM agent_team_members WHERE team_id = $1 ORDER BY position ASC`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.AgentTeamMember, 0)
	for rows.Next() {
		var member model.AgentTeamMember
		if err := rows.Scan(&member.ID, &member.TeamID, &member.ProfileID, &member.MemberRole, &member.Position, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}
