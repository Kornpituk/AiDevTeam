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

type mockAgentProfileRepo struct {
	createErr  error
	getAllErr  error
	getByIDErr error
	profiles   []model.AgentProfile
}

func (m *mockAgentProfileRepo) Create(profile *model.AgentProfile) error {
	if m.createErr != nil {
		return m.createErr
	}
	profile.ID = testUUID
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	return nil
}

func (m *mockAgentProfileRepo) GetAll() ([]model.AgentProfile, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.profiles, nil
}

func (m *mockAgentProfileRepo) GetByID(id string) (*model.AgentProfile, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, p := range m.profiles {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockProfiles() []model.AgentProfile {
	return []model.AgentProfile{
		{
			ID:           testUUID,
			Name:         "Senior Developer",
			Role:         "developer",
			Description:  "Expert in Go and backend development",
			SystemPrompt: "You are a senior developer...",
			DefaultModel: "gpt-4",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           testUUID2,
			Name:         "Code Reviewer",
			Role:         "reviewer",
			Description:  "Reviews code for quality and best practices",
			SystemPrompt: "You are a code reviewer...",
			DefaultModel: "claude-3",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
}

func TestCreateAgentProfile(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           map[string]string{"name": "Test Agent", "role": "developer", "description": "Test description"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing name",
			body:           map[string]string{"role": "developer"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing role",
			body:           map[string]string{"name": "Test Agent"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			body:           map[string]string{"name": "Test Agent", "role": "developer"},
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

			req := httptest.NewRequest("POST", "/agent-profiles", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mock := &mockAgentProfileRepo{createErr: tt.mockErr}
			handler := NewAgentProfileHandler(mock)
			handler.CreateAgentProfile(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentProfiles(t *testing.T) {
	tests := []struct {
		name           string
		profiles       []model.AgentProfile
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with profiles",
			profiles:       createMockProfiles(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			profiles:       []model.AgentProfile{},
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
			req := httptest.NewRequest("GET", "/agent-profiles", nil)
			w := httptest.NewRecorder()

			mock := &mockAgentProfileRepo{profiles: tt.profiles, getAllErr: tt.mockErr}
			handler := NewAgentProfileHandler(mock)
			handler.GetAgentProfiles(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentProfile(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		profiles       []model.AgentProfile
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			profiles:       createMockProfiles(),
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
			profiles:       createMockProfiles(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-profiles/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockAgentProfileRepo{profiles: tt.profiles}
			handler := NewAgentProfileHandler(mock)
			handler.GetAgentProfile(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
