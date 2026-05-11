package repository

import (
	"database/sql"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	query := `INSERT INTO ai_tasks (title, description, status) VALUES ($1, $2, $3) RETURNING id, plan, review_notes, created_at, updated_at`
	var plan sql.NullString
	var reviewNotes sql.NullString
	err := r.db.QueryRow(query, task.Title, task.Description, task.Status).Scan(&task.ID, &plan, &reviewNotes, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return err
	}
	task.Plan = plan.String
	task.ReviewNotes = reviewNotes.String
	return nil
}

func (r *TaskRepository) GetAll() ([]model.Task, error) {
	query := `SELECT id, title, description, status, plan, review_notes, created_at, updated_at FROM ai_tasks ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var plan sql.NullString
		var reviewNotes sql.NullString
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &plan, &reviewNotes, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		task.Plan = plan.String
		task.ReviewNotes = reviewNotes.String
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) GetByID(id string) (*model.Task, error) {
	query := `SELECT id, title, description, status, plan, review_notes, created_at, updated_at FROM ai_tasks WHERE id = $1`
	var task model.Task
	var plan sql.NullString
	var reviewNotes sql.NullString
	err := r.db.QueryRow(query, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &plan, &reviewNotes, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	task.Plan = plan.String
	task.ReviewNotes = reviewNotes.String
	return &task, nil
}

func (r *TaskRepository) UpdateStatus(id string, status string) (*model.Task, error) {
	query := `UPDATE ai_tasks SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, title, description, status, plan, review_notes, created_at, updated_at`
	var task model.Task
	var plan sql.NullString
	var reviewNotes sql.NullString
	err := r.db.QueryRow(query, status, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &plan, &reviewNotes, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	task.Plan = plan.String
	task.ReviewNotes = reviewNotes.String
	return &task, nil
}

func (r *TaskRepository) UpdatePlan(id string, plan string) (*model.Task, error) {
	query := `UPDATE ai_tasks SET plan = $1, updated_at = NOW() WHERE id = $2 RETURNING id, title, description, status, plan, review_notes, created_at, updated_at`
	var task model.Task
	var planResult sql.NullString
	var reviewNotes sql.NullString
	err := r.db.QueryRow(query, plan, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &planResult, &reviewNotes, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	task.Plan = planResult.String
	task.ReviewNotes = reviewNotes.String
	return &task, nil
}

func (r *TaskRepository) UpdateReviewNotes(id string, reviewNotes string) (*model.Task, error) {
	query := `UPDATE ai_tasks SET review_notes = $1, updated_at = NOW() WHERE id = $2 RETURNING id, title, description, status, plan, review_notes, created_at, updated_at`
	var task model.Task
	var plan sql.NullString
	var reviewNotesResult sql.NullString
	err := r.db.QueryRow(query, reviewNotes, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &plan, &reviewNotesResult, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	task.Plan = plan.String
	task.ReviewNotes = reviewNotesResult.String
	return &task, nil
}

func (r *TaskRepository) DB() *sql.DB {
	return r.db
}

func (r *TaskRepository) SetConnMaxLifetime(d time.Duration) {
	r.db.SetConnMaxLifetime(d)
}

func (r *TaskRepository) SetMaxOpenConns(n int) {
	r.db.SetMaxOpenConns(n)
}

func (r *TaskRepository) SetMaxIdleConns(n int) {
	r.db.SetMaxIdleConns(n)
}
