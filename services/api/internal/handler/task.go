package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

func isNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

var validStatuses = map[string]bool{
	"pending":     true,
	"planning":    true,
	"approved":    true,
	"in_progress": true,
	"reviewing":   true,
	"completed":   true,
	"failed":      true,
}

type TaskRepository interface {
	Create(task *model.Task) error
	GetAll() ([]model.Task, error)
	GetByID(id string) (*model.Task, error)
	UpdateStatus(id string, status string) (*model.Task, error)
	UpdatePlan(id string, plan string) (*model.Task, error)
	UpdateReviewNotes(id string, reviewNotes string) (*model.Task, error)
}

type TaskHandler struct {
	taskRepo TaskRepository
}

func NewTaskHandler(taskRepo TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
}

func isValidUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}
	for i, c := range id {
		switch i {
		case 8, 13, 18, 23:
			continue
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if task.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	task.Status = "pending"

	if err := h.taskRepo.Create(&task); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskRepo.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get task")
		}
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validStatuses[req.Status] {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	task, err := h.taskRepo.UpdateStatus(id, req.Status)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update task status")
		}
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	task, err := h.taskRepo.UpdatePlan(id, req.Plan)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update task plan")
		}
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskReviewNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req struct {
		ReviewNotes string `json:"review_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	task, err := h.taskRepo.UpdateReviewNotes(id, req.ReviewNotes)
	if err != nil {
		if isNotFoundError(err) {
			respondError(w, http.StatusNotFound, "Task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update task review notes")
		}
		return
	}

	respondJSON(w, http.StatusOK, task)
}
