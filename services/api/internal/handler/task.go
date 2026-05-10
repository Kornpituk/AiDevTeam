package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type TaskRepository interface {
	Create(task *model.Task) error
	GetAll() ([]model.Task, error)
	GetByID(id int) (*model.Task, error)
	UpdateStatus(id int, status string) error
}

type TaskHandler struct {
	taskRepo TaskRepository
}

func NewTaskHandler(taskRepo TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
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

	if err := h.taskRepo.UpdateStatus(id, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update task status")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
}