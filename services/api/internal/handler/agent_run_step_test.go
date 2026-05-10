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

type mockAgentRunStepRepo struct {
	createErr      error
	getByRunIDErr  error
	updateStatusErr error
	getByIDErr     error
	steps          []model.AgentRunStep
}

func (m *mockAgentRunStepRepo) Create(step *model.AgentRunStep) error {
	if m.createErr != nil {
		return m.createErr
	}
	step.ID = testUUID
	step.CreatedAt = time.Now()
	step.UpdatedAt = time.Now()
	return nil
}

func (m *mockAgentRunStepRepo) GetByRunID(runID string) ([]model.AgentRunStep, error) {
	if m.getByRunIDErr != nil {
		return nil, m.getByRunIDErr
	}
	var result []model.AgentRunStep
	for _, s := range m.steps {
		if s.RunID == runID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockAgentRunStepRepo) UpdateStatus(id string, status string) (*model.AgentRunStep, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	for i, s := range m.steps {
		if s.ID == id {
			updated := s
			updated.Status = status
			m.steps[i] = updated
			return &updated, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockAgentRunStepRepo) GetByID(id string) (*model.AgentRunStep, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, s := range m.steps {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockSteps() []model.AgentRunStep {
	return []model.AgentRunStep{
		{
			ID:           testUUID,
			RunID:        testUUID2,
			ProfileID:    "550e8400-e29b-41d4-a716-446655440002",
			StepType:     "code_generation",
			Status:       "pending",
			Title:        "Implement feature X",
			Instructions: "Write code for feature X",
			Position:     0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           testUUID2,
			RunID:        testUUID2,
			ProfileID:    "550e8400-e29b-41d4-a716-446655440003",
			StepType:     "code_review",
			Status:       "pending",
			Title:        "Review feature X",
			Instructions: "Review the code for feature X",
			Position:     1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
}

func TestCreateRunStep(t *testing.T) {
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
			body:           map[string]string{"step_type": "code_generation", "title": "Implement feature", "instructions": "Do it"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "success with default status pending",
			runID:          testUUID,
			body:           map[string]string{"step_type": "code_generation", "title": "Implement feature"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid run ID",
			runID:          "abc",
			body:           map[string]string{"step_type": "code_generation", "title": "Implement feature"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			runID:          testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing step_type",
			runID:          testUUID,
			body:           map[string]string{"title": "Implement feature"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing title",
			runID:          testUUID,
			body:           map[string]string{"step_type": "code_generation"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			runID:          testUUID,
			body:           map[string]string{"step_type": "code_generation", "title": "Implement", "status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid profile_id",
			runID:          testUUID,
			body:           map[string]string{"step_type": "code_generation", "title": "Implement", "profile_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			runID:          testUUID,
			body:           map[string]string{"step_type": "code_generation", "title": "Implement feature"},
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

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.runID+"/steps", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunStepRepo{createErr: tt.mockErr}
			handler := NewAgentRunStepHandler(mock)
			handler.CreateRunStep(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetRunSteps(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		steps          []model.AgentRunStep
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with steps",
			runID:          testUUID2,
			steps:          createMockSteps(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			runID:          testUUID,
			steps:          []model.AgentRunStep{},
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
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.runID+"/steps", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunStepRepo{steps: tt.steps, getByRunIDErr: tt.mockErr}
			handler := NewAgentRunStepHandler(mock)
			handler.GetRunSteps(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateStepStatus(t *testing.T) {
	tests := []struct {
		name           string
		stepID         string
		body           interface{}
		steps          []model.AgentRunStep
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			stepID:         testUUID,
			body:           map[string]string{"status": "running"},
			steps:          createMockSteps(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with completed",
			stepID:         testUUID,
			body:           map[string]string{"status": "completed"},
			steps:          createMockSteps(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid step ID",
			stepID:         "abc",
			body:           map[string]string{"status": "running"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			stepID:         testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			stepID:         testUUID,
			body:           map[string]string{"status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			stepID:         "550e8400-e29b-41d4-a716-446655449999",
			body:           map[string]string{"status": "running"},
			steps:          createMockSteps(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repo error",
			stepID:         testUUID,
			body:           map[string]string{"status": "running"},
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

			req := httptest.NewRequest("PATCH", "/agent-run-steps/"+tt.stepID+"/status", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.stepID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunStepRepo{steps: tt.steps, updateStatusErr: tt.mockErr}
			handler := NewAgentRunStepHandler(mock)
			handler.UpdateStepStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestStepStatusValidation(t *testing.T) {
	validTestStatuses := []string{"pending", "waiting_approval", "running", "completed", "failed", "skipped", "cancelled"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validStepStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "PENDING", "unknown", "done"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validStepStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}
