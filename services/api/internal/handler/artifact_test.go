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

type mockArtifactRepo struct {
	createErr   error
	getByIDErr  error
	artifacts   []model.TaskArtifact
	createdArtifact *model.TaskArtifact
}

func (m *mockArtifactRepo) Create(artifact *model.TaskArtifact) error {
	if m.createErr != nil {
		return m.createErr
	}
	artifact.ID = 1
	m.createdArtifact = artifact
	return nil
}

func (m *mockArtifactRepo) GetByTaskID(taskID int) ([]model.TaskArtifact, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	var result []model.TaskArtifact
	for _, a := range m.artifacts {
		if a.TaskID == taskID {
			result = append(result, a)
		}
	}
	return result, nil
}

func TestCreateArtifact(t *testing.T) {
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
			body:           map[string]string{"name": "test.go", "type": "code", "content": "package main"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           map[string]string{"name": "test.go"},
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
			body:           map[string]string{"name": "test.go"},
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

			req := httptest.NewRequest("POST", "/tasks/"+tt.id+"/artifacts", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockArtifactRepo{createErr: tt.mockErr}
			handler := NewArtifactHandler(mock)
			handler.CreateArtifact(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetArtifacts(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		artifacts      []model.TaskArtifact
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with artifacts",
			id:             "1",
			artifacts:      []model.TaskArtifact{{ID: 1, TaskID: 1, Name: "test.go"}, {ID: 2, TaskID: 1, Name: "test2.go"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			id:             "1",
			artifacts:      []model.TaskArtifact{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			id:             "1",
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tasks/"+tt.id+"/artifacts", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockArtifactRepo{artifacts: tt.artifacts, getByIDErr: tt.mockErr}
			handler := NewArtifactHandler(mock)
			handler.GetArtifacts(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
