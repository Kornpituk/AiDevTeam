package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type EventRepository interface {
	Create(event *model.TaskEvent) error
	GetByTaskID(taskID string) ([]model.TaskEvent, error)
}

type EventHandler struct {
	eventRepo EventRepository
}

func NewEventHandler(eventRepo EventRepository) *EventHandler {
	return &EventHandler{eventRepo: eventRepo}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if !isValidUUID(taskID) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var event model.TaskEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	event.TaskID = taskID

	if err := h.eventRepo.Create(&event); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	respondJSON(w, http.StatusCreated, event)
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if !isValidUUID(taskID) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	events, err := h.eventRepo.GetByTaskID(taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	respondJSON(w, http.StatusOK, events)
}
