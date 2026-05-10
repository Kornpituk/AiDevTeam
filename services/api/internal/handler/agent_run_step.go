package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

var validStepStatuses = map[string]bool{
	"pending":          true,
	"waiting_approval": true,
	"running":          true,
	"completed":        true,
	"failed":           true,
	"skipped":          true,
	"cancelled":        true,
}

type AgentRunStepRepository interface {
	Create(step *model.AgentRunStep) error
	GetByRunID(runID string) ([]model.AgentRunStep, error)
	UpdateStatus(id string, status string) (*model.AgentRunStep, error)
	GetByID(id string) (*model.AgentRunStep, error)
}

type AgentRunStepHandler struct {
	stepRepo AgentRunStepRepository
}

func NewAgentRunStepHandler(stepRepo AgentRunStepRepository) *AgentRunStepHandler {
	return &AgentRunStepHandler{stepRepo: stepRepo}
}

func (h *AgentRunStepHandler) CreateRunStep(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	var step model.AgentRunStep
	if err := json.NewDecoder(r.Body).Decode(&step); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	step.RunID = runID

	if step.ProfileID != "" && !isValidUUID(step.ProfileID) {
		respondError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	if step.StepType == "" {
		respondError(w, http.StatusBadRequest, "Step type is required")
		return
	}

	if step.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if step.Status == "" {
		step.Status = "pending"
	}

	if !validStepStatuses[step.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.stepRepo.Create(&step); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create run step")
		return
	}

	respondJSON(w, http.StatusCreated, step)
}

func (h *AgentRunStepHandler) GetRunSteps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	steps, err := h.stepRepo.GetByRunID(runID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get run steps")
		return
	}

	respondJSON(w, http.StatusOK, steps)
}

func (h *AgentRunStepHandler) UpdateStepStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid step ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validStepStatuses[req.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	step, err := h.stepRepo.UpdateStatus(id, req.Status)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Step not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update step status")
		}
		return
	}

	respondJSON(w, http.StatusOK, step)
}
