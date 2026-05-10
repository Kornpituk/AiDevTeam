package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type AgentTeamRepository interface {
	Create(team *model.AgentTeam) error
	GetAll() ([]model.AgentTeam, error)
	GetByID(id string) (*model.AgentTeam, error)
}

type AgentTeamMemberRepository interface {
	Create(member *model.AgentTeamMember) error
	GetByTeamID(teamID string) ([]model.AgentTeamMember, error)
}

type AgentTeamHandler struct {
	teamRepo       AgentTeamRepository
	memberRepo     AgentTeamMemberRepository
}

func NewAgentTeamHandler(teamRepo AgentTeamRepository, memberRepo AgentTeamMemberRepository) *AgentTeamHandler {
	return &AgentTeamHandler{teamRepo: teamRepo, memberRepo: memberRepo}
}

func (h *AgentTeamHandler) CreateAgentTeam(w http.ResponseWriter, r *http.Request) {
	var team model.AgentTeam
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if team.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if err := h.teamRepo.Create(&team); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create agent team")
		return
	}

	respondJSON(w, http.StatusCreated, team)
}

func (h *AgentTeamHandler) GetAgentTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teamRepo.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get agent teams")
		return
	}

	respondJSON(w, http.StatusOK, teams)
}

func (h *AgentTeamHandler) GetAgentTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid agent team ID")
		return
	}

	team, err := h.teamRepo.GetByID(id)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Agent team not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get agent team")
		}
		return
	}

	respondJSON(w, http.StatusOK, team)
}

func (h *AgentTeamHandler) CreateTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	if !isValidUUID(teamID) {
		respondError(w, http.StatusBadRequest, "Invalid agent team ID")
		return
	}

	var member model.AgentTeamMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if member.ProfileID == "" {
		respondError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	if !isValidUUID(member.ProfileID) {
		respondError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	if member.MemberRole == "" {
		respondError(w, http.StatusBadRequest, "Member role is required")
		return
	}

	member.TeamID = teamID

	if err := h.memberRepo.Create(&member); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add team member")
		return
	}

	respondJSON(w, http.StatusCreated, member)
}

func (h *AgentTeamHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["id"]

	if !isValidUUID(teamID) {
		respondError(w, http.StatusBadRequest, "Invalid agent team ID")
		return
	}

	members, err := h.memberRepo.GetByTeamID(teamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get team members")
		return
	}

	respondJSON(w, http.StatusOK, members)
}
