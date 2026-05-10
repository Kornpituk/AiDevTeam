package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/gorilla/mux"
)

type mockTaskRepo struct {
	createErr    error
	getAllErr    error
	getByIDErr   error
	updateErr    error
	tasks        []model.Task
	createdTask  *model.Task
	updatedID    int
	updatedStatus string
}

func (m *mockTaskRepo) Create(task *model.Task) error {
	if m.createErr != nil {
		return m.createErr
	}
	task.ID = 1
	m.createdTask = task
	return nil
}

func (m *mockTaskRepo) GetAll() ([]model.Task, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.tasks, nil
}

func (m *mockTaskRepo) GetByID(id int) (*model.Task, error) {
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

func (m *mockTaskRepo) UpdateStatus(id int, status string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedID = id
	m.updatedStatus = status
	return nil
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
			tasks:          []model.Task{{ID: 1, Title: "Task 1"}, {ID: 2, Title: "Task 2"}},
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
			id:             "1",
			tasks:          []model.Task{{ID: 1, Title: "Task 1"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             "999",
			tasks:          []model.Task{{ID: 1, Title: "Task 1"}},
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
			name:           "success",
			id:             "1",
			body:           map[string]string{"status": "in_progress"},
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
			id:             "1",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			id:             "1",
			body:           map[string]string{"status": "done"},
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

			mock := &mockTaskRepo{updateErr: tt.mockErr}
			handler := NewTaskHandler(mock)
			handler.UpdateTaskStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
