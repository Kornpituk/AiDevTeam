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

type mockHumanApprovalRepo struct {
	createErr      error
	getByRunIDErr  error
	updateStatusErr error
	getByIDErr     error
	approvals      []model.HumanApproval
}

func (m *mockHumanApprovalRepo) Create(approval *model.HumanApproval) error {
	if m.createErr != nil {
		return m.createErr
	}
	approval.ID = testUUID
	approval.CreatedAt = time.Now()
	return nil
}

func (m *mockHumanApprovalRepo) GetByRunID(runID string) ([]model.HumanApproval, error) {
	if m.getByRunIDErr != nil {
		return nil, m.getByRunIDErr
	}
	var result []model.HumanApproval
	for _, a := range m.approvals {
		if a.RunID == runID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockHumanApprovalRepo) UpdateStatus(id string, status string) (*model.HumanApproval, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	for i, a := range m.approvals {
		if a.ID == id {
			updated := a
			updated.Status = status
			m.approvals[i] = updated
			return &updated, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockHumanApprovalRepo) GetByID(id string) (*model.HumanApproval, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, a := range m.approvals {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockApprovals() []model.HumanApproval {
	return []model.HumanApproval{
		{
			ID:            testUUID,
			TaskID:        "550e8400-e29b-41d4-a716-446655440002",
			RunID:         testUUID2,
			StepID:        "550e8400-e29b-41d4-a716-446655440003",
			ApprovalType:  "step_execution",
			Status:        "pending",
			RequestedBy:   "ai_developer",
			RequestNotes:  "Need approval to execute this step",
			CreatedAt:     time.Now(),
		},
		{
			ID:            testUUID2,
			TaskID:        "550e8400-e29b-41d4-a716-446655440002",
			RunID:         testUUID2,
			ApprovalType:  "run_completion",
			Status:        "pending",
			RequestNotes:  "Need approval to mark run as complete",
			CreatedAt:     time.Now(),
		},
	}
}

func TestCreateApproval(t *testing.T) {
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
			body:           map[string]string{"approval_type": "step_execution", "request_notes": "Need approval"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "success defaults status to pending",
			runID:          testUUID,
			body:           map[string]string{"approval_type": "step_execution"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid run ID",
			runID:          "abc",
			body:           map[string]string{"approval_type": "step_execution"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			runID:          testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing approval_type",
			runID:          testUUID,
			body:           map[string]string{"request_notes": "Need approval"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			runID:          testUUID,
			body:           map[string]string{"approval_type": "step_execution", "status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid task_id",
			runID:          testUUID,
			body:           map[string]string{"approval_type": "step_execution", "task_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid step_id",
			runID:          testUUID,
			body:           map[string]string{"approval_type": "step_execution", "step_id": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			runID:          testUUID,
			body:           map[string]string{"approval_type": "step_execution"},
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

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.runID+"/approvals", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockHumanApprovalRepo{createErr: tt.mockErr}
			handler := NewHumanApprovalHandler(mock)
			handler.CreateApproval(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetApprovals(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		approvals      []model.HumanApproval
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with approvals",
			runID:          testUUID2,
			approvals:      createMockApprovals(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			runID:          testUUID,
			approvals:      []model.HumanApproval{},
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
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.runID+"/approvals", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.runID})
			w := httptest.NewRecorder()

			mock := &mockHumanApprovalRepo{approvals: tt.approvals, getByRunIDErr: tt.mockErr}
			handler := NewHumanApprovalHandler(mock)
			handler.GetApprovals(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateApprovalStatus(t *testing.T) {
	tests := []struct {
		name           string
		approvalID     string
		body           interface{}
		approvals      []model.HumanApproval
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			approvalID:     testUUID,
			body:           map[string]string{"status": "approved"},
			approvals:      createMockApprovals(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with rejected",
			approvalID:     testUUID,
			body:           map[string]string{"status": "rejected"},
			approvals:      createMockApprovals(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid approval ID",
			approvalID:     "abc",
			body:           map[string]string{"status": "approved"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			approvalID:     testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			approvalID:     testUUID,
			body:           map[string]string{"status": "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			approvalID:     "550e8400-e29b-41d4-a716-446655449999",
			body:           map[string]string{"status": "approved"},
			approvals:      createMockApprovals(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repo error",
			approvalID:     testUUID,
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

			req := httptest.NewRequest("PATCH", "/human-approvals/"+tt.approvalID+"/status", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.approvalID})
			w := httptest.NewRecorder()

			mock := &mockHumanApprovalRepo{approvals: tt.approvals, updateStatusErr: tt.mockErr}
			handler := NewHumanApprovalHandler(mock)
			handler.UpdateApprovalStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestApprovalStatusValidation(t *testing.T) {
	validTestStatuses := []string{"pending", "approved", "rejected", "cancelled"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validApprovalStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "PENDING", "done", "running"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validApprovalStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}
