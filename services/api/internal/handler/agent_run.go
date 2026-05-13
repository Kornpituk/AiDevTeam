package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/service"
)

var validRunStatuses = map[string]bool{
	"draft":           true,
	"planned":         true,
	"waiting_approval": true,
	"approved":        true,
	"running":         true,
	"paused":          true,
	"completed":       true,
	"failed":          true,
	"cancelled":       true,
}

type AgentRunRepository interface {
	Create(run *model.AgentRun) error
	GetByTaskID(taskID string) ([]model.AgentRun, error)
	GetByID(id string) (*model.AgentRun, error)
}

type AgentRunHandler struct {
	runRepo      AgentRunRepository
	orchestrator *service.Orchestrator
}

func NewAgentRunHandler(runRepo AgentRunRepository, orchestrator *service.Orchestrator) *AgentRunHandler {
	return &AgentRunHandler{
		runRepo:      runRepo,
		orchestrator: orchestrator,
	}
}

func (h *AgentRunHandler) CreateAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if !isValidUUID(taskID) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var run model.AgentRun
	if err := json.NewDecoder(r.Body).Decode(&run); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	run.TaskID = taskID

	if run.TeamID != "" && !isValidUUID(run.TeamID) {
		respondError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	if run.Status == "" {
		run.Status = "draft"
	}

	if !validRunStatuses[run.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.runRepo.Create(&run); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create agent run")
		return
	}

	respondJSON(w, http.StatusCreated, run)
}

func (h *AgentRunHandler) GetAgentRunsByTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if !isValidUUID(taskID) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	runs, err := h.runRepo.GetByTaskID(taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get agent runs")
		return
	}

	respondJSON(w, http.StatusOK, runs)
}

func (h *AgentRunHandler) GetAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	run, err := h.runRepo.GetByID(id)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Agent run not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get agent run")
		}
		return
	}

	respondJSON(w, http.StatusOK, run)
}

func (h *AgentRunHandler) StartAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	if h.orchestrator == nil {
		respondError(w, http.StatusInternalServerError, "Orchestrator not initialized")
		return
	}

	err := h.orchestrator.StartRun(nil, id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	run, err := h.runRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated run")
		return
	}

	respondJSON(w, http.StatusOK, run)
}

func (h *AgentRunHandler) ResumeAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	if h.orchestrator == nil {
		respondError(w, http.StatusInternalServerError, "Orchestrator not initialized")
		return
	}

	err := h.orchestrator.ResumeRun(nil, id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	run, err := h.runRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated run")
		return
	}

	respondJSON(w, http.StatusOK, run)
}

func (h *AgentRunHandler) CancelAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	if h.orchestrator == nil {
		respondError(w, http.StatusInternalServerError, "Orchestrator not initialized")
		return
	}

	err := h.orchestrator.CancelRun(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	run, err := h.runRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated run")
		return
	}

	respondJSON(w, http.StatusOK, run)
}
