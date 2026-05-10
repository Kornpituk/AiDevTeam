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

const testUUID = "550e8400-e29b-41d4-a716-446655440000"
const testUUID2 = "550e8400-e29b-41d4-a716-446655440001"

type mockTaskRepo struct {
	createErr       error
	getAllErr       error
	getByIDErr      error
	updateStatusErr error
	updatePlanErr   error
	updateReviewErr error
	tasks           []model.Task
	createdTask     *model.Task
	updatedID       string
	updatedStatus   string
	updatedPlan     string
	updatedReview   string
}

func (m *mockTaskRepo) Create(task *model.Task) error {
	if m.createErr != nil {
		return m.createErr
	}
	task.ID = testUUID
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	m.createdTask = task
	return nil
}

func (m *mockTaskRepo) GetAll() ([]model.Task, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.tasks, nil
}

func (m *mockTaskRepo) GetByID(id string) (*model.Task, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, t := range m.tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockTaskRepo) UpdateStatus(id string, status string) (*model.Task, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	m.updatedID = id
	m.updatedStatus = status
	for i, t := range m.tasks {
		if t.ID == id {
			updated := t
			updated.Status = status
			m.tasks[i] = updated
			return &updated, nil
		}
	}
	return &model.Task{ID: id, Status: status}, nil
}

func (m *mockTaskRepo) UpdatePlan(id string, plan string) (*model.Task, error) {
	if m.updatePlanErr != nil {
		return nil, m.updatePlanErr
	}
	m.updatedID = id
	m.updatedPlan = plan
	for i, t := range m.tasks {
		if t.ID == id {
			updated := t
			updated.Plan = plan
			m.tasks[i] = updated
			return &updated, nil
		}
	}
	return &model.Task{ID: id, Plan: plan}, nil
}

func (m *mockTaskRepo) UpdateReviewNotes(id string, reviewNotes string) (*model.Task, error) {
	if m.updateReviewErr != nil {
		return nil, m.updateReviewErr
	}
	m.updatedID = id
	m.updatedReview = reviewNotes
	for i, t := range m.tasks {
		if t.ID == id {
			updated := t
			updated.ReviewNotes = reviewNotes
			m.tasks[i] = updated
			return &updated, nil
		}
	}
	return &model.Task{ID: id, ReviewNotes: reviewNotes}, nil
}

func createMockTasks() []model.Task {
	return []model.Task{
		{
			ID:          testUUID,
			Title:       "Test Task 1",
			Description: "Test Description 1",
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          testUUID2,
			Title:       "Test Task 2",
			Description: "Test Description 2",
			Status:      "in_progress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           map[string]string{"title": "Test Task", "description": "Test desc"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing title",
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
			body:           map[string]string{"title": "Test Task"},
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

			req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{createErr: tt.mockErr}
			handler := NewTaskHandler(mock)
			handler.CreateTask(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetTasks(t *testing.T) {
	tests := []struct {
		name           string
		tasks          []model.Task
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with tasks",
			tasks:          createMockTasks(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			tasks:          []model.Task{},
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
			req := httptest.NewRequest("GET", "/tasks", nil)
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{tasks: tt.tasks, getAllErr: tt.mockErr}
			handler := NewTaskHandler(mock)
			handler.GetTasks(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		tasks          []model.Task
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			tasks:          createMockTasks(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id format (too short)",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid id format (wrong hyphens)",
			id:             "550e8400e29b41d4a716446655440000",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             "550e8400-e29b-41d4-a716-446655449999",
			tasks:          createMockTasks(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tasks/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{tasks: tt.tasks}
			handler := NewTaskHandler(mock)
			handler.GetTask(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with valid status",
			id:             testUUID,
			body:           map[string]string{"status": "in_progress"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with completed status",
			id:             testUUID,
			body:           map[string]string{"status": "completed"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           map[string]string{"status": "in_progress"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			id:             testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			id:             testUUID,
			body:           map[string]string{"status": "invalid_status"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			id:             testUUID,
			body:           map[string]string{"status": "completed"},
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

			req := httptest.NewRequest("PATCH", "/tasks/"+tt.id+"/status", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{updateStatusErr: tt.mockErr, tasks: createMockTasks()}
			handler := NewTaskHandler(mock)
			handler.UpdateTaskStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateTaskPlan(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			body:           map[string]string{"plan": "Phase 1: Do this\nPhase 2: Do that"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with empty plan",
			id:             testUUID,
			body:           map[string]string{"plan": ""},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           map[string]string{"plan": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			id:             testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			id:             testUUID,
			body:           map[string]string{"plan": "test"},
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

			req := httptest.NewRequest("PATCH", "/tasks/"+tt.id+"/plan", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{updatePlanErr: tt.mockErr}
			handler := NewTaskHandler(mock)
			handler.UpdateTaskPlan(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateTaskReviewNotes(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			body:           map[string]string{"review_notes": "Looks good! Some minor changes needed."},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with empty notes",
			id:             testUUID,
			body:           map[string]string{"review_notes": ""},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           map[string]string{"review_notes": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			id:             testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			id:             testUUID,
			body:           map[string]string{"review_notes": "test"},
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

			req := httptest.NewRequest("PATCH", "/tasks/"+tt.id+"/review-notes", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockTaskRepo{updateReviewErr: tt.mockErr}
			handler := NewTaskHandler(mock)
			handler.UpdateTaskReviewNotes(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestStatusValidation(t *testing.T) {
	validTestStatuses := []string{"pending", "planning", "approved", "in_progress", "reviewing", "completed", "failed"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "DONE", "IN_PROGRESS", "unknown"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid UUID", testUUID, true},
		{"valid UUID 2", testUUID2, true},
		{"too short", "abc", false},
		{"too long", "550e8400-e29b-41d4-a716-4466554400000", false},
		{"wrong hyphen positions", "550e8400e29b-41d4-a716-446655440000", false},
		{"no hyphens", "550e8400e29b41d4a716446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUUID(tt.id)
			if result != tt.expected {
				t.Errorf("isValidUUID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}
