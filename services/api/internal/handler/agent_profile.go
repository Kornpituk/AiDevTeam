package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentProfileRepository interface {
	Create(profile *model.AgentProfile) error
	GetAll() ([]model.AgentProfile, error)
	GetByID(id string) (*model.AgentProfile, error)
}

type AgentProfileHandler struct {
	profileRepo AgentProfileRepository
}

func NewAgentProfileHandler(profileRepo AgentProfileRepository) *AgentProfileHandler {
	return &AgentProfileHandler{profileRepo: profileRepo}
}

func (h *AgentProfileHandler) CreateAgentProfile(w http.ResponseWriter, r *http.Request) {
	var profile model.AgentProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if profile.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if profile.Role == "" {
		respondError(w, http.StatusBadRequest, "Role is required")
		return
	}

	if err := h.profileRepo.Create(&profile); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create agent profile")
		return
	}

	respondJSON(w, http.StatusCreated, profile)
}

func (h *AgentProfileHandler) GetAgentProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.profileRepo.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get agent profiles")
		return
	}

	respondJSON(w, http.StatusOK, profiles)
}

func (h *AgentProfileHandler) GetAgentProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent profile ID")
		return
	}

	profile, err := h.profileRepo.GetByID(id)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Agent profile not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get agent profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, profile)
}
