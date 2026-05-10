package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type ArtifactRepository interface {
	Create(artifact *model.TaskArtifact) error
	GetByTaskID(taskID int) ([]model.TaskArtifact, error)
}

type ArtifactHandler struct {
	artifactRepo ArtifactRepository
}

func NewArtifactHandler(artifactRepo ArtifactRepository) *ArtifactHandler {
	return &ArtifactHandler{artifactRepo: artifactRepo}
}

func (h *ArtifactHandler) CreateArtifact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var artifact model.TaskArtifact
	if err := json.NewDecoder(r.Body).Decode(&artifact); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	artifact.TaskID = taskID

	if err := h.artifactRepo.Create(&artifact); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create artifact")
		return
	}

	respondJSON(w, http.StatusCreated, artifact)
}

func (h *ArtifactHandler) GetArtifacts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	artifacts, err := h.artifactRepo.GetByTaskID(taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get artifacts")
		return
	}

	respondJSON(w, http.StatusOK, artifacts)
}