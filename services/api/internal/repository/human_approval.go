package repository

import (
	"database/sql"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type HumanApprovalRepository struct {
	db *sql.DB
}

func NewHumanApprovalRepository(db *sql.DB) *HumanApprovalRepository {
	return &HumanApprovalRepository{db: db}
}

func (r *HumanApprovalRepository) Create(approval *model.HumanApproval) error {
	query := `INSERT INTO human_approvals (task_id, run_id, step_id, approval_type, status, requested_by, decided_by, request_notes, decision_notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`
	var taskID sql.NullString
	var runID sql.NullString
	var stepID sql.NullString
	var requestedBy sql.NullString
	var decidedBy sql.NullString
	var requestNotes sql.NullString
	var decisionNotes sql.NullString
	if approval.TaskID == "" {
		taskID = sql.NullString{Valid: false}
	} else {
		taskID = sql.NullString{String: approval.TaskID, Valid: true}
	}
	if approval.RunID == "" {
		runID = sql.NullString{Valid: false}
	} else {
		runID = sql.NullString{String: approval.RunID, Valid: true}
	}
	if approval.StepID == "" {
		stepID = sql.NullString{Valid: false}
	} else {
		stepID = sql.NullString{String: approval.StepID, Valid: true}
	}
	if approval.RequestedBy == "" {
		requestedBy = sql.NullString{Valid: false}
	} else {
		requestedBy = sql.NullString{String: approval.RequestedBy, Valid: true}
	}
	if approval.DecidedBy == "" {
		decidedBy = sql.NullString{Valid: false}
	} else {
		decidedBy = sql.NullString{String: approval.DecidedBy, Valid: true}
	}
	if approval.RequestNotes == "" {
		requestNotes = sql.NullString{Valid: false}
	} else {
		requestNotes = sql.NullString{String: approval.RequestNotes, Valid: true}
	}
	if approval.DecisionNotes == "" {
		decisionNotes = sql.NullString{Valid: false}
	} else {
		decisionNotes = sql.NullString{String: approval.DecisionNotes, Valid: true}
	}
	return r.db.QueryRow(query, taskID, runID, stepID, approval.ApprovalType, approval.Status, requestedBy, decidedBy, requestNotes, decisionNotes).Scan(&approval.ID, &approval.CreatedAt)
}

func (r *HumanApprovalRepository) GetByRunID(runID string) ([]model.HumanApproval, error) {
	query := `SELECT id, task_id, run_id, step_id, approval_type, status, requested_by, decided_by, request_notes, decision_notes, created_at, decided_at FROM human_approvals WHERE run_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	approvals := make([]model.HumanApproval, 0)
	for rows.Next() {
		var approval model.HumanApproval
		var taskID sql.NullString
		var runID_ sql.NullString
		var stepID sql.NullString
		var requestedBy sql.NullString
		var decidedBy sql.NullString
		var requestNotes sql.NullString
		var decisionNotes sql.NullString
		var decidedAt sql.NullTime
		if err := rows.Scan(&approval.ID, &taskID, &runID_, &stepID, &approval.ApprovalType, &approval.Status, &requestedBy, &decidedBy, &requestNotes, &decisionNotes, &approval.CreatedAt, &decidedAt); err != nil {
			return nil, err
		}
		approval.TaskID = taskID.String
		approval.RunID = runID_.String
		approval.StepID = stepID.String
		approval.RequestedBy = requestedBy.String
		approval.DecidedBy = decidedBy.String
		approval.RequestNotes = requestNotes.String
		approval.DecisionNotes = decisionNotes.String
		if decidedAt.Valid {
			t := decidedAt.Time
			approval.DecidedAt = &t
		}
		approvals = append(approvals, approval)
	}
	return approvals, nil
}

func (r *HumanApprovalRepository) UpdateStatus(id string, status string) (*model.HumanApproval, error) {
	query := `UPDATE human_approvals SET status = $1, decided_at = NOW() WHERE id = $2 RETURNING id, task_id, run_id, step_id, approval_type, status, requested_by, decided_by, request_notes, decision_notes, created_at, decided_at`
	var approval model.HumanApproval
	var taskID sql.NullString
	var runID sql.NullString
	var stepID sql.NullString
	var requestedBy sql.NullString
	var decidedBy sql.NullString
	var requestNotes sql.NullString
	var decisionNotes sql.NullString
	var decidedAt sql.NullTime
	err := r.db.QueryRow(query, status, id).Scan(&approval.ID, &taskID, &runID, &stepID, &approval.ApprovalType, &approval.Status, &requestedBy, &decidedBy, &requestNotes, &decisionNotes, &approval.CreatedAt, &decidedAt)
	if err != nil {
		return nil, err
	}
	approval.TaskID = taskID.String
	approval.RunID = runID.String
	approval.StepID = stepID.String
	approval.RequestedBy = requestedBy.String
	approval.DecidedBy = decidedBy.String
	approval.RequestNotes = requestNotes.String
	approval.DecisionNotes = decisionNotes.String
	if decidedAt.Valid {
		t := decidedAt.Time
		approval.DecidedAt = &t
	}
	return &approval, nil
}

func (r *HumanApprovalRepository) GetByID(id string) (*model.HumanApproval, error) {
	query := `SELECT id, task_id, run_id, step_id, approval_type, status, requested_by, decided_by, request_notes, decision_notes, created_at, decided_at FROM human_approvals WHERE id = $1`
	var approval model.HumanApproval
	var taskID sql.NullString
	var runID sql.NullString
	var stepID sql.NullString
	var requestedBy sql.NullString
	var decidedBy sql.NullString
	var requestNotes sql.NullString
	var decisionNotes sql.NullString
	var decidedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(&approval.ID, &taskID, &runID, &stepID, &approval.ApprovalType, &approval.Status, &requestedBy, &decidedBy, &requestNotes, &decisionNotes, &approval.CreatedAt, &decidedAt)
	if err != nil {
		return nil, err
	}
	approval.TaskID = taskID.String
	approval.RunID = runID.String
	approval.StepID = stepID.String
	approval.RequestedBy = requestedBy.String
	approval.DecidedBy = decidedBy.String
	approval.RequestNotes = requestNotes.String
	approval.DecisionNotes = decisionNotes.String
	if decidedAt.Valid {
		t := decidedAt.Time
		approval.DecidedAt = &t
	}
	return &approval, nil
}
