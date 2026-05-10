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

type mockAgentRunRepo struct {
	createErr    error
	getByTaskIDErr error
	getByIDErr   error
	runs         []model.AgentRun
}

func (m *mockAgentRunRepo) Create(run *model.AgentRun) error {
	if m.createErr != nil {
		return m.createErr
	}
	run.ID = testUUID
	run.CreatedAt = time.Now()
	run.UpdatedAt = time.Now()
	return nil
}

func (m *mockAgentRunRepo) GetByTaskID(taskID string) ([]model.AgentRun, error) {
	if m.getByTaskIDErr != nil {
		return nil, m.getByTaskIDErr
	}
	var result []model.AgentRun
	for _, r := range m.runs {
		if r.TaskID == taskID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockAgentRunRepo) GetByID(id string) (*model.AgentRun, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, r := range m.runs {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockRuns() []model.AgentRun {
	return []model.AgentRun{
		{
			ID:        testUUID,
			TaskID:    testUUID2,
			TeamID:    "550e8400-e29b-41d4-a716-446655440002",
			Status:    "draft",
			Goal:      "Implement feature X",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func TestCreateAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			taskID:         testUUID,
			body:           map[string]string{"status": "draft", "goal": "Test goal"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "success with default status",
			taskID:         testUUID,
			body:           map[string]string{"goal": "Test goal"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid task_id",
			taskID:         "abc",
			body:           map[string]string{"goal": "Test goal"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			taskID:         testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			taskID:         testUUID,
			body:           map[string]string{"status": "invalid", "goal": "Test goal"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			taskID:         testUUID,
			body:           map[string]string{"goal": "Test goal"},
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

			req := httptest.NewRequest("POST", "/tasks/"+tt.taskID+"/agent-runs", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.taskID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{createErr: tt.mockErr}
			handler := NewAgentRunHandler(mock)
			handler.CreateAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentRunsByTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		runs           []model.AgentRun
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with runs",
			taskID:         testUUID2,
			runs:           createMockRuns(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			taskID:         testUUID,
			runs:           []model.AgentRun{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid task_id",
			taskID:         "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			taskID:         testUUID,
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID+"/agent-runs", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.taskID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{runs: tt.runs, getByTaskIDErr: tt.mockErr}
			handler := NewAgentRunHandler(mock)
			handler.GetAgentRunsByTask(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		runs           []model.AgentRun
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			runs:           createMockRuns(),
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
			runs:           createMockRuns(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{runs: tt.runs}
			handler := NewAgentRunHandler(mock)
			handler.GetAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRunStatusValidation(t *testing.T) {
	validTestStatuses := []string{"draft", "planned", "waiting_approval", "approved", "running", "paused", "completed", "failed", "cancelled"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validRunStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "DRAFT", "unknown"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validRunStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}
