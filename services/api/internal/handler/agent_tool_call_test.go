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

type mockAgentToolCallRepo struct {
	createErr       error
	getByRunIDErr   error
	updateStatusErr error
	getByIDErr      error
	toolCalls       []model.AgentToolCall
}

func (m *mockAgentToolCallRepo) Create(toolCall *model.AgentToolCall) error {
	if m.createErr != nil {
		return m.createErr
	}
	toolCall.ID = testUUID
	toolCall.CreatedAt = time.Now()
	return nil
}

func (m *mockAgentToolCallRepo) GetByRunID(runID string) ([]model.AgentToolCall, error) {
	if m.getByRunIDErr != nil {
		return nil, m.getByRunIDErr
	}
	var result []model.AgentToolCall
	for _, tc := range m.toolCalls {
		if tc.RunID == runID {
			result = append(result, tc)
		}
	}
	return result, nil
}

func (m *mockAgentToolCallRepo) UpdateStatus(id string, status string) (*model.AgentToolCall, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	for i, tc := range m.toolCalls {
		if tc.ID == id {
			updated := tc
			updated.Status = status
			m.toolCalls[i] = updated
			return &updated, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockAgentToolCallRepo) GetByID(id string) (*model.AgentToolCall, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, tc := range m.toolCalls {
		if tc.ID == id {
			return &tc, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockToolCalls() []model.AgentToolCall {
	return []model.AgentToolCall{
		{
			ID:        testUUID,
			RunID:     testUUID2,
			StepID:    "550e8400-e29b-41d4-a716-446655440002",
			ToolName:  "get_file",
			Input:     []byte(`{"path": "test.go"}`),
			Output:    []byte(`{"content": "package main"}`),
			Status:    "recorded",
			CreatedAt: time.Now(),
		},
		{
			ID:        testUUID2,
			RunID:     testUUID2,
			ToolName:  "edit_file",
			Input:     []byte(`{"path": "test.go"}`),
			Status:    "recorded",
			CreatedAt: time.Now(),
		},
	}
}

func TestCreateToolCall(t *testing.T) {
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
			body:           map[string]string{"tool_name": "get_file", "input": `{"path": "test.go"}`},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "success defaults status to recorded",
			runID:          testUUID,
			body:           map[string]string{"tool_name": "get_file"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid run ID",
			runID:          "abc",
			body:           map[string]string{"tool_name": "get_file"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			runID:          testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing tool_name",
			runID:          testUUID,
			body:           map[string]string{"input": `{"path": "test.go"}`},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			runID:          testUUID,
			body:           map[string]string{"tool_name": "get_file", "status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid step_id",
			runID:          testUUID,
			body:           map[string]string{"tool_name": "get_file", "step_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			runID:          testUUID,
			body:           map[string]string{"tool_name": "get_file"},
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

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.runID+"/tool-calls", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentToolCallRepo{createErr: tt.mockErr}
			handler := NewAgentToolCallHandler(mock)
			handler.CreateToolCall(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetToolCalls(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		toolCalls      []model.AgentToolCall
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with tool calls",
			runID:          testUUID2,
			toolCalls:      createMockToolCalls(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			runID:          testUUID,
			toolCalls:      []model.AgentToolCall{},
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
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.runID+"/tool-calls", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockAgentToolCallRepo{toolCalls: tt.toolCalls, getByRunIDErr: tt.mockErr}
			handler := NewAgentToolCallHandler(mock)
			handler.GetToolCalls(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateToolCallStatus(t *testing.T) {
	tests := []struct {
		name           string
		toolCallID     string
		body           interface{}
		toolCalls      []model.AgentToolCall
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			toolCallID:     testUUID,
			body:           map[string]string{"status": "approved"},
			toolCalls:      createMockToolCalls(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with completed",
			toolCallID:     testUUID,
			body:           map[string]string{"status": "completed"},
			toolCalls:      createMockToolCalls(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with failed",
			toolCallID:     testUUID,
			body:           map[string]string{"status": "failed"},
			toolCalls:      createMockToolCalls(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid tool call ID",
			toolCallID:     "abc",
			body:           map[string]string{"status": "approved"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			toolCallID:     testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			toolCallID:     testUUID,
			body:           map[string]string{"status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			toolCallID:     "550e8400-e29b-41d4-a716-446655449999",
			body:           map[string]string{"status": "approved"},
			toolCalls:      createMockToolCalls(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repo error",
			toolCallID:     testUUID,
			body:           map[string]string{"status": "approved"},
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

			req := httptest.NewRequest("PATCH", "/agent-tool-calls/"+tt.toolCallID+"/status", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.toolCallID})
			w := httptest.NewRecorder()

			mock := &mockAgentToolCallRepo{toolCalls: tt.toolCalls, updateStatusErr: tt.mockErr}
			handler := NewAgentToolCallHandler(mock)
			handler.UpdateToolCallStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestToolCallStatusValidation(t *testing.T) {
	validTestStatuses := []string{"recorded", "approved", "rejected", "completed", "failed"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validToolCallStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "RECORDED", "pending", "running", "done"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validToolCallStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}
