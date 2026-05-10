package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/gorilla/mux"
)

type mockAgentTeamRepo struct {
	createErr  error
	getAllErr  error
	getByIDErr error
	teams      []model.AgentTeam
}

type mockAgentTeamMemberRepo struct {
	createErr      error
	getByTeamIDErr error
	members        []model.AgentTeamMember
}

func (m *mockAgentTeamRepo) Create(team *model.AgentTeam) error {
	if m.createErr != nil {
		return m.createErr
	}
	team.ID = testUUID
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	return nil
}

func (m *mockAgentTeamRepo) GetAll() ([]model.AgentTeam, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.teams, nil
}

func (m *mockAgentTeamRepo) GetByID(id string) (*model.AgentTeam, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, t := range m.teams {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockAgentTeamMemberRepo) Create(member *model.AgentTeamMember) error {
	if m.createErr != nil {
		return m.createErr
	}
	member.ID = testUUID
	member.CreatedAt = time.Now()
	return nil
}

func (m *mockAgentTeamMemberRepo) GetByTeamID(teamID string) ([]model.AgentTeamMember, error) {
	if m.getByTeamIDErr != nil {
		return nil, m.getByTeamIDErr
	}
	var result []model.AgentTeamMember
	for _, m := range m.members {
		if m.TeamID == teamID {
			result = append(result, m)
		}
	}
	return result, nil
}

func createMockTeams() []model.AgentTeam {
	return []model.AgentTeam{
		{
			ID:          testUUID,
			Name:        "Development Team",
			Description: "A team for development tasks",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          testUUID2,
			Name:        "Review Team",
			Description: "A team for code reviews",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

func createMockTeamMembers() []model.AgentTeamMember {
	return []model.AgentTeamMember{
		{
			ID:         testUUID,
			TeamID:     testUUID,
			ProfileID:  testUUID2,
			MemberRole: "lead_developer",
			Position:   0,
			CreatedAt:  time.Now(),
		},
	}
}

func TestCreateAgentTeam(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           map[string]string{"name": "Test Team", "description": "Test description"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing name",
			body:           map[string]string{"description": "Test desc"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			body:           map[string]string{"name": "Test Team"},
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/agent-teams", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mockTeam := &mockAgentTeamRepo{createErr: tt.mockErr}
			mockMember := &mockAgentTeamMemberRepo{}
			handler := NewAgentTeamHandler(mockTeam, mockMember)
			handler.CreateAgentTeam(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentTeams(t *testing.T) {
	tests := []struct {
		name           string
		teams          []model.AgentTeam
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with teams",
			teams:          createMockTeams(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			teams:          []model.AgentTeam{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "repo error",
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-teams", nil)
			w := httptest.NewRecorder()

			mockTeam := &mockAgentTeamRepo{teams: tt.teams, getAllErr: tt.mockErr}
			mockMember := &mockAgentTeamMemberRepo{}
			handler := NewAgentTeamHandler(mockTeam, mockMember)
			handler.GetAgentTeams(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentTeam(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		teams          []model.AgentTeam
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			teams:          createMockTeams(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id format",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             "550e8400-e29b-41d4-a716-446655449999",
			teams:          createMockTeams(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-teams/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mockTeam := &mockAgentTeamRepo{teams: tt.teams}
			mockMember := &mockAgentTeamMemberRepo{}
			handler := NewAgentTeamHandler(mockTeam, mockMember)
			handler.GetAgentTeam(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestCreateTeamMember(t *testing.T) {
	tests := []struct {
		name           string
		teamID         string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			teamID:         testUUID,
			body:           map[string]string{"profile_id": testUUID2, "member_role": "developer"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing profile_id",
			teamID:         testUUID,
			body:           map[string]string{"member_role": "developer"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing member_role",
			teamID:         testUUID,
			body:           map[string]string{"profile_id": testUUID2},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid profile_id",
			teamID:         testUUID,
			body:           map[string]string{"profile_id": "invalid", "member_role": "developer"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid team_id",
			teamID:         "abc",
			body:           map[string]string{"profile_id": testUUID2, "member_role": "developer"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			teamID:         testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			teamID:         testUUID,
			body:           map[string]string{"profile_id": testUUID2, "member_role": "developer"},
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/agent-teams/"+tt.teamID+"/members", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.teamID})
			w := httptest.NewRecorder()

			mockTeam := &mockAgentTeamRepo{}
			mockMember := &mockAgentTeamMemberRepo{createErr: tt.mockErr}
			handler := NewAgentTeamHandler(mockTeam, mockMember)
			handler.CreateTeamMember(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetTeamMembers(t *testing.T) {
	tests := []struct {
		name           string
		teamID         string
		members        []model.AgentTeamMember
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with members",
			teamID:         testUUID,
			members:        createMockTeamMembers(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			teamID:         testUUID,
			members:        []model.AgentTeamMember{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid team_id",
			teamID:         "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			teamID:         testUUID,
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-teams/"+tt.teamID+"/members", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.teamID})
			w := httptest.NewRecorder()

			mockTeam := &mockAgentTeamRepo{}
			mockMember := &mockAgentTeamMemberRepo{members: tt.members, getByTeamIDErr: tt.mockErr}
			handler := NewAgentTeamHandler(mockTeam, mockMember)
			handler.GetTeamMembers(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
