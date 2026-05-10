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

type mockEventRepo struct {
	createErr   error
	getByIDErr  error
	events      []model.TaskEvent
	createdEvent *model.TaskEvent
}

func (m *mockEventRepo) Create(event *model.TaskEvent) error {
	if m.createErr != nil {
		return m.createErr
	}
	event.ID = 1
	m.createdEvent = event
	return nil
}

func (m *mockEventRepo) GetByTaskID(taskID int) ([]model.TaskEvent, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	var result []model.TaskEvent
	for _, e := range m.events {
		if e.TaskID == taskID {
			result = append(result, e)
		}
	}
	return result, nil
}

func TestCreateEvent(t *testing.T) {
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
			body:           map[string]string{"event_type": "note", "data": "Test note"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           map[string]string{"event_type": "note"},
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
			body:           map[string]string{"event_type": "note"},
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

			req := httptest.NewRequest("POST", "/tasks/"+tt.id+"/events", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockEventRepo{createErr: tt.mockErr}
			handler := NewEventHandler(mock)
			handler.CreateEvent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		events         []model.TaskEvent
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with events",
			id:             "1",
			events:         []model.TaskEvent{{ID: 1, TaskID: 1, EventType: "note"}, {ID: 2, TaskID: 1, EventType: "log"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			id:             "1",
			events:         []model.TaskEvent{},
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
			req := httptest.NewRequest("GET", "/tasks/"+tt.id+"/events", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockEventRepo{events: tt.events, getByIDErr: tt.mockErr}
			handler := NewEventHandler(mock)
			handler.GetEvents(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
