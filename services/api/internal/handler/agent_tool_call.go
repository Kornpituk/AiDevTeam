package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

var validToolCallStatuses = map[string]bool{
	"recorded":  true,
	"approved":  true,
	"rejected":  true,
	"completed": true,
	"failed":    true,
}

type AgentToolCallRepository interface {
	Create(toolCall *model.AgentToolCall) error
	GetByRunID(runID string) ([]model.AgentToolCall, error)
	UpdateStatus(id string, status string) (*model.AgentToolCall, error)
	GetByID(id string) (*model.AgentToolCall, error)
}

type AgentToolCallHandler struct {
	toolCallRepo AgentToolCallRepository
}

func NewAgentToolCallHandler(toolCallRepo AgentToolCallRepository) *AgentToolCallHandler {
	return &AgentToolCallHandler{toolCallRepo: toolCallRepo}
}

func (h *AgentToolCallHandler) CreateToolCall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	var toolCall model.AgentToolCall
	if err := json.NewDecoder(r.Body).Decode(&toolCall); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	toolCall.RunID = runID

	if toolCall.ToolName == "" {
		respondError(w, http.StatusBadRequest, "Tool name is required")
		return
	}

	if toolCall.Status == "" {
		toolCall.Status = "recorded"
	}

	if !validToolCallStatuses[toolCall.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if toolCall.StepID != "" && !isValidUUID(toolCall.StepID) {
		respondError(w, http.StatusBadRequest, "Invalid step ID")
		return
	}

	if err := h.toolCallRepo.Create(&toolCall); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create tool call")
		return
	}

	respondJSON(w, http.StatusCreated, toolCall)
}

func (h *AgentToolCallHandler) GetToolCalls(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	toolCalls, err := h.toolCallRepo.GetByRunID(runID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get tool calls")
		return
	}

	respondJSON(w, http.StatusOK, toolCalls)
}

func (h *AgentToolCallHandler) UpdateToolCallStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid tool call ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validToolCallStatuses[req.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	toolCall, err := h.toolCallRepo.UpdateStatus(id, req.Status)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Tool call not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update tool call status")
		}
		return
	}

	respondJSON(w, http.StatusOK, toolCall)
}
