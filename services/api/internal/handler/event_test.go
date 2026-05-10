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
	createErr    error
	getByIDErr   error
	events       []model.TaskEvent
	createdEvent *model.TaskEvent
}

func (m *mockEventRepo) Create(event *model.TaskEvent) error {
	if m.createErr != nil {
		return m.createErr
	}
	event.ID = testUUID
	m.createdEvent = event
	return nil
}

func (m *mockEventRepo) GetByTaskID(taskID string) ([]model.TaskEvent, error) {
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
			id:             testUUID,
			body:           map[string]string{"event_type": "note", "message": "Test note"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid id format",
			id:             "abc",
			body:           map[string]string{"event_type": "note"},
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
			id:             testUUID,
			events: []model.TaskEvent{
				{ID: "e1", TaskID: testUUID, EventType: "note", Message: "Test 1"},
				{ID: "e2", TaskID: testUUID, EventType: "log", Message: "Test 2"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			id:             testUUID,
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
			id:             testUUID,
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
