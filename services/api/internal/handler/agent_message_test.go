package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/gorilla/mux"
)

type mockAgentMessageRepo struct {
	createErr      error
	getByRunIDErr  error
	messages       []model.AgentMessage
}

func (m *mockAgentMessageRepo) Create(message *model.AgentMessage) error {
	if m.createErr != nil {
		return m.createErr
	}
	message.ID = testUUID
	message.CreatedAt = time.Now()
	return nil
}

func (m *mockAgentMessageRepo) GetByRunID(runID string) ([]model.AgentMessage, error) {
	if m.getByRunIDErr != nil {
		return nil, m.getByRunIDErr
	}
	var result []model.AgentMessage
	for _, msg := range m.messages {
		if msg.RunID == runID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func createMockMessages() []model.AgentMessage {
	return []model.AgentMessage{
		{
			ID:        testUUID,
			RunID:     testUUID2,
			StepID:    "550e8400-e29b-41d4-a716-446655440002",
			ProfileID: "550e8400-e29b-41d4-a716-446655440003",
			Role:      "user",
			Content:   "Please implement this feature",
			Metadata:  []byte("{}"),
			CreatedAt: time.Now(),
		},
		{
			ID:        testUUID2,
			RunID:     testUUID2,
			ProfileID: "550e8400-e29b-41d4-a716-446655440003",
			Role:      "assistant",
			Content:   "I'll implement this feature now",
			Metadata:  []byte("{}"),
			CreatedAt: time.Now(),
		},
	}
}

func TestCreateMessage(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			runID:          testUUID,
			body:           map[string]string{"role": "user", "content": "Hello world"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid run ID",
			runID:          "abc",
			body:           map[string]string{"role": "user", "content": "Hello"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			runID:          testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing role",
			runID:          testUUID,
			body:           map[string]string{"content": "Hello"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid role",
			runID:          testUUID,
			body:           map[string]string{"role": "invalid", "content": "Hello"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing content",
			runID:          testUUID,
			body:           map[string]string{"role": "user"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid step_id",
			runID:          testUUID,
			body:           map[string]string{"role": "user", "content": "Hello", "step_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid profile_id",
			runID:          testUUID,
			body:           map[string]string{"role": "user", "content": "Hello", "profile_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			runID:          testUUID,
			body:           map[string]string{"role": "user", "content": "Hello"},
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

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.runID+"/messages", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentMessageRepo{createErr: tt.mockErr}
			handler := NewAgentMessageHandler(mock)
			handler.CreateMessage(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		messages       []model.AgentMessage
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with messages",
			runID:          testUUID2,
			messages:       createMockMessages(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			runID:          testUUID,
			messages:       []model.AgentMessage{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid run ID",
			runID:          "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			runID:          testUUID,
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.runID+"/messages", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentMessageRepo{messages: tt.messages, getByRunIDErr: tt.mockErr}
			handler := NewAgentMessageHandler(mock)
			handler.GetMessages(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMessageRoleValidation(t *testing.T) {
	validTestRoles := []string{"system", "user", "assistant", "tool", "reviewer"}

	for _, role := range validTestRoles {
		t.Run("valid role: "+role, func(t *testing.T) {
			if !validMessageRoles[role] {
				t.Errorf("expected %s to be valid", role)
			}
		})
	}

	invalidTestRoles := []string{"", "invalid", "SYSTEM", "admin", "moderator"}

	for _, role := range invalidTestRoles {
		t.Run("invalid role: "+role, func(t *testing.T) {
			if validMessageRoles[role] {
				t.Errorf("expected %s to be invalid", role)
			}
		})
	}
}
