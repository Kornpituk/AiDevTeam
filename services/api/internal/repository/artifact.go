package repository

import (
	"database/sql"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type ArtifactRepository struct {
	db *sql.DB
}

func NewArtifactRepository(db *sql.DB) *ArtifactRepository {
	return &ArtifactRepository{db: db}
}

func (r *ArtifactRepository) Create(artifact *model.TaskArtifact) error {
	query := `INSERT INTO ai_task_artifacts (task_id, name, type, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, artifact.TaskID, artifact.Name, artifact.Type, artifact.Content).Scan(&artifact.ID, &artifact.CreatedAt)
}

func (r *ArtifactRepository) GetByTaskID(taskID int) ([]model.TaskArtifact, error) {
	query := `SELECT id, task_id, name, type, content, created_at FROM ai_task_artifacts WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []model.TaskArtifact
	for rows.Next() {
		var artifact model.TaskArtifact
		if err := rows.Scan(&artifact.ID, &artifact.TaskID, &artifact.Name, &artifact.Type, &artifact.Content, &artifact.CreatedAt); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}