package model

import (
	"encoding/json"
	"time"
)

type AgentProfile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Description  string    `json:"description"`
	SystemPrompt string    `json:"system_prompt"`
	DefaultModel string    `json:"default_model"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AgentTeam struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AgentTeamMember struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	ProfileID string    `json:"profile_id"`
	MemberRole string   `json:"member_role"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type AgentRun struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	TeamID    string    `json:"team_id"`
	Status    string    `json:"status"`
	Goal      string    `json:"goal"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgentRunStep struct {
	ID           string     `json:"id"`
	RunID        string     `json:"run_id"`
	ProfileID    string     `json:"profile_id"`
	StepType     string     `json:"step_type"`
	Status       string     `json:"status"`
	Title        string     `json:"title"`
	Instructions string     `json:"instructions"`
	Output       string     `json:"output"`
	Position     int        `json:"position"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AgentMessage struct {
	ID        string          `json:"id"`
	RunID     string          `json:"run_id"`
	StepID    string          `json:"step_id"`
	ProfileID string          `json:"profile_id"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

type AgentToolCall struct {
	ID          string          `json:"id"`
	RunID       string          `json:"run_id"`
	StepID      string          `json:"step_id"`
	ToolName    string          `json:"tool_name"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at"`
}

type HumanApproval struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	RunID         string     `json:"run_id"`
	StepID        string     `json:"step_id"`
	ApprovalType  string     `json:"approval_type"`
	Status        string     `json:"status"`
	RequestedBy   string     `json:"requested_by"`
	DecidedBy     string     `json:"decided_by"`
	RequestNotes  string     `json:"request_notes"`
	DecisionNotes string     `json:"decision_notes"`
	ToolCallID    *string    `json:"tool_call_id"`
	CreatedAt     time.Time  `json:"created_at"`
	DecidedAt     *time.Time `json:"decided_at"`
}
