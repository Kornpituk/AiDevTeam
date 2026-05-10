package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

var validMessageRoles = map[string]bool{
	"system":    true,
	"user":      true,
	"assistant": true,
	"tool":      true,
	"reviewer":  true,
}

type AgentMessageRepository interface {
	Create(message *model.AgentMessage) error
	GetByRunID(runID string) ([]model.AgentMessage, error)
}

type AgentMessageHandler struct {
	messageRepo AgentMessageRepository
}

func NewAgentMessageHandler(messageRepo AgentMessageRepository) *AgentMessageHandler {
	return &AgentMessageHandler{messageRepo: messageRepo}
}

func (h *AgentMessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	var message model.AgentMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	message.RunID = runID

	if message.Role == "" {
		respondError(w, http.StatusBadRequest, "Role is required")
		return
	}

	if !validMessageRoles[message.Role] {
		respondError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	if message.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	if message.StepID != "" && !isValidUUID(message.StepID) {
		respondError(w, http.StatusBadRequest, "Invalid step ID")
		return
	}

	if message.ProfileID != "" && !isValidUUID(message.ProfileID) {
		respondError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	if err := h.messageRepo.Create(&message); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	respondJSON(w, http.StatusCreated, message)
}

func (h *AgentMessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	messages, err := h.messageRepo.GetByRunID(runID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
}
