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
	query := `INSERT INTO ai_task_artifacts (task_id, name, artifact_type, content, content_type, metadata) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	var metadata []byte
	if len(artifact.Metadata) > 0 {
		metadata = artifact.Metadata
	} else {
		metadata = []byte("{}")
	}
	return r.db.QueryRow(query, artifact.TaskID, artifact.Name, artifact.ArtifactType, artifact.Content, artifact.ContentType, metadata).Scan(&artifact.ID, &artifact.CreatedAt)
}

func (r *ArtifactRepository) GetByTaskID(taskID string) ([]model.TaskArtifact, error) {
	query := `SELECT id, task_id, name, artifact_type, content, content_type, metadata, created_at FROM ai_task_artifacts WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []model.TaskArtifact
	for rows.Next() {
		var artifact model.TaskArtifact
		var content sql.NullString
		var contentType sql.NullString
		var metadata []byte
		if err := rows.Scan(&artifact.ID, &artifact.TaskID, &artifact.Name, &artifact.ArtifactType, &content, &contentType, &metadata, &artifact.CreatedAt); err != nil {
			return nil, err
		}
		artifact.Content = content.String
		artifact.ContentType = contentType.String
		artifact.Metadata = metadata
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}
