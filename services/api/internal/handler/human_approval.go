package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

var validApprovalStatuses = map[string]bool{
	"pending":   true,
	"approved":  true,
	"rejected":  true,
	"cancelled": true,
}

type HumanApprovalRepository interface {
	Create(approval *model.HumanApproval) error
	GetByRunID(runID string) ([]model.HumanApproval, error)
	UpdateStatus(id string, status string) (*model.HumanApproval, error)
	GetByID(id string) (*model.HumanApproval, error)
}

type HumanApprovalHandler struct {
	approvalRepo HumanApprovalRepository
}

func NewHumanApprovalHandler(approvalRepo HumanApprovalRepository) *HumanApprovalHandler {
	return &HumanApprovalHandler{approvalRepo: approvalRepo}
}

func (h *HumanApprovalHandler) CreateApproval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	var approval model.HumanApproval
	if err := json.NewDecoder(r.Body).Decode(&approval); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	approval.RunID = runID

	if approval.ApprovalType == "" {
		respondError(w, http.StatusBadRequest, "Approval type is required")
		return
	}

	if approval.Status == "" {
		approval.Status = "pending"
	}

	if !validApprovalStatuses[approval.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if approval.TaskID != "" && !isValidUUID(approval.TaskID) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if approval.StepID != "" && !isValidUUID(approval.StepID) {
		respondError(w, http.StatusBadRequest, "Invalid step ID")
		return
	}

	if err := h.approvalRepo.Create(&approval); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create approval")
		return
	}

	respondJSON(w, http.StatusCreated, approval)
}

func (h *HumanApprovalHandler) GetApprovals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]

	if !isValidUUID(runID) {
		respondError(w, http.StatusBadRequest, "Invalid agent run ID")
		return
	}

	approvals, err := h.approvalRepo.GetByRunID(runID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get approvals")
		return
	}

	respondJSON(w, http.StatusOK, approvals)
}

func (h *HumanApprovalHandler) UpdateApprovalStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid approval ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validApprovalStatuses[req.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	approval, err := h.approvalRepo.UpdateStatus(id, req.Status)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Approval not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update approval status")
		}
		return
	}

	respondJSON(w, http.StatusOK, approval)
}
