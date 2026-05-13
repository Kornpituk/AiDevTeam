package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/service"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/tool"
	"github.com/gorilla/mux"
)

type mockAgentRunRepo struct {
	createErr    error
	getByTaskIDErr error
	getByIDErr   error
	runs         []model.AgentRun
}

func (m *mockAgentRunRepo) Create(run *model.AgentRun) error {
	if m.createErr != nil {
		return m.createErr
	}
	run.ID = testUUID
	run.CreatedAt = time.Now()
	run.UpdatedAt = time.Now()
	return nil
}

func (m *mockAgentRunRepo) GetByTaskID(taskID string) ([]model.AgentRun, error) {
	if m.getByTaskIDErr != nil {
		return nil, m.getByTaskIDErr
	}
	var result []model.AgentRun
	for _, r := range m.runs {
		if r.TaskID == taskID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockAgentRunRepo) GetByID(id string) (*model.AgentRun, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, r := range m.runs {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, sql.ErrNoRows
}

func createMockRuns() []model.AgentRun {
	return []model.AgentRun{
		{
			ID:        testUUID,
			TaskID:    testUUID2,
			TeamID:    "550e8400-e29b-41d4-a716-446655440002",
			Status:    "draft",
			Goal:      "Implement feature X",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func TestCreateAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			taskID:         testUUID,
			body:           map[string]string{"status": "draft", "goal": "Test goal"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "success with default status",
			taskID:         testUUID,
			body:           map[string]string{"goal": "Test goal"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid task_id",
			taskID:         "abc",
			body:           map[string]string{"goal": "Test goal"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			taskID:         testUUID,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			taskID:         testUUID,
			body:           map[string]string{"status": "invalid", "goal": "Test goal"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			taskID:         testUUID,
			body:           map[string]string{"goal": "Test goal"},
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/tasks/"+tt.taskID+"/agent-runs", bytes.NewReader(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": tt.taskID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{createErr: tt.mockErr}
			handler := NewAgentRunHandler(mock, nil)
			handler.CreateAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentRunsByTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		runs           []model.AgentRun
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success with runs",
			taskID:         testUUID2,
			runs:           createMockRuns(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success empty",
			taskID:         testUUID,
			runs:           []model.AgentRun{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid task_id",
			taskID:         "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			taskID:         testUUID,
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID+"/agent-runs", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.taskID})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{runs: tt.runs, getByTaskIDErr: tt.mockErr}
			handler := NewAgentRunHandler(mock, nil)
			handler.GetAgentRunsByTask(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		runs           []model.AgentRun
		expectedStatus int
	}{
		{
			name:           "success",
			id:             testUUID,
			runs:           createMockRuns(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id format",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             "550e8400-e29b-41d4-a716-446655449999",
			runs:           createMockRuns(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/agent-runs/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			mock := &mockAgentRunRepo{runs: tt.runs}
			handler := NewAgentRunHandler(mock, nil)
			handler.GetAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRunStatusValidation(t *testing.T) {
	validTestStatuses := []string{"draft", "planned", "waiting_approval", "approved", "running", "paused", "completed", "failed", "cancelled"}

	for _, status := range validTestStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			if !validRunStatuses[status] {
				t.Errorf("expected %s to be valid", status)
			}
		})
	}

	invalidTestStatuses := []string{"", "invalid", "DRAFT", "unknown"}

	for _, status := range invalidTestStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			if validRunStatuses[status] {
				t.Errorf("expected %s to be invalid", status)
			}
		})
	}
}

type mockOrchestratorRepo struct {
	run       *model.AgentRun
	steps     []model.AgentRunStep
	members   []model.AgentTeamMember
	profiles  map[string]*model.AgentProfile
	returnErr error

	createdSteps  []model.AgentRunStep
	statusHistory []string
}

func newMockOrchestratorRepo() *mockOrchestratorRepo {
	return &mockOrchestratorRepo{
		profiles: make(map[string]*model.AgentProfile),
	}
}

func (m *mockOrchestratorRepo) GetRunByID(id string) (*model.AgentRun, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if m.run == nil {
		return nil, sql.ErrNoRows
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) UpdateRunStatus(id string, status string) (*model.AgentRun, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	m.statusHistory = append(m.statusHistory, status)
	if m.run != nil {
		m.run.Status = status
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) GetTeamMembers(teamID string) ([]model.AgentTeamMember, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.members, nil
}

func (m *mockOrchestratorRepo) GetProfileByID(id string) (*model.AgentProfile, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if p, ok := m.profiles[id]; ok {
		return p, nil
	}
	return &model.AgentProfile{
		ID:           id,
		Name:         "Test Agent",
		Role:         "planner",
		SystemPrompt: "You are a test agent.",
	}, nil
}

func (m *mockOrchestratorRepo) CreateStep(step *model.AgentRunStep) error {
	if m.returnErr != nil {
		return m.returnErr
	}
	m.createdSteps = append(m.createdSteps, *step)
	return nil
}

func (m *mockOrchestratorRepo) GetStepsByRunID(runID string) ([]model.AgentRunStep, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.steps, nil
}

func (m *mockOrchestratorRepo) MarkStepStarted(id string) (*model.AgentRunStep, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return &model.AgentRunStep{ID: id, Status: "in_progress"}, nil
}

func (m *mockOrchestratorRepo) MarkStepCompleted(id string, output string) (*model.AgentRunStep, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return &model.AgentRunStep{ID: id, Status: "completed", Output: output}, nil
}

func (m *mockOrchestratorRepo) MarkStepFailed(id string, output string) (*model.AgentRunStep, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return &model.AgentRunStep{ID: id, Status: "failed", Output: output}, nil
}

func (m *mockOrchestratorRepo) CreateMessage(message *model.AgentMessage) error {
	return m.returnErr
}

func (m *mockOrchestratorRepo) UpdateRunSummary(id string, summary string) (*model.AgentRun, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if m.run != nil {
		m.run.Summary = summary
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) UpdateRunStatusIfIn(id string, newStatus string, allowedStatuses []string) (*model.AgentRun, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if m.run != nil {
		allowed := false
		for _, s := range allowedStatuses {
			if m.run.Status == s {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errors.New("run is not in startable state")
		}
		m.statusHistory = append(m.statusHistory, newStatus)
		m.run.Status = newStatus
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) GetTaskByID(id string) (*model.Task, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return &model.Task{
		ID:    id,
		Title: "Test Task",
	}, nil
}

func (m *mockOrchestratorRepo) CreateToolCall(toolCall *model.AgentToolCall) error {
	return m.returnErr
}

func (m *mockOrchestratorRepo) UpdateToolCallStatus(id string, status string) (*model.AgentToolCall, error) {
	return &model.AgentToolCall{ID: id, Status: status}, nil
}

func (m *mockOrchestratorRepo) CreateApproval(approval *model.HumanApproval) error {
	return m.returnErr
}

func (m *mockOrchestratorRepo) GetApprovedToolApprovals(runID, stepID string) ([]model.HumanApproval, error) {
	return nil, nil
}

func (m *mockOrchestratorRepo) UpdateToolCallOutput(id string, output []byte, status string) (*model.AgentToolCall, error) {
	return &model.AgentToolCall{ID: id, Status: status, Output: output}, nil
}

func (m *mockOrchestratorRepo) MarkStepWaitingApproval(id string) (*model.AgentRunStep, error) {
	for i := range m.steps {
		if m.steps[i].ID == id {
			m.steps[i].Status = "waiting_approval"
			return &m.steps[i], nil
		}
	}
	return &model.AgentRunStep{ID: id, Status: "waiting_approval"}, nil
}

func (m *mockOrchestratorRepo) GetMessagesByStepID(stepID string) ([]model.AgentMessage, error) {
	return nil, nil
}

type mockLLM struct {
	responses   []string
	responseIdx int
	returnErr   error
}

func (m *mockLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if m.returnErr != nil {
		return "", m.returnErr
	}
	if len(m.responses) == 0 {
		return "fake response", nil
	}
	resp := m.responses[m.responseIdx%len(m.responses)]
	m.responseIdx++
	return resp, nil
}

func (m *mockLLM) ChatCompletionWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition, toolChoice string) (*llm.ChatCompletionResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if len(m.responses) == 0 {
		return &llm.ChatCompletionResponse{Content: "fake response"}, nil
	}
	resp := m.responses[m.responseIdx%len(m.responses)]
	m.responseIdx++
	return &llm.ChatCompletionResponse{Content: resp}, nil
}

func TestStartAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupMock      func() (*mockOrchestratorRepo, *mockAgentRunRepo)
		useNilOrch     bool
		expectedStatus int
	}{
		{
			name: "success",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "draft",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{*orchRepo.run},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			id:             "not-a-uuid",
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nil orchestrator",
			id:             testUUID,
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			useNilOrch:     true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "orchestrator error - not startable",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "completed",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{*orchRepo.run},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "run not found",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.returnErr = sql.ErrNoRows
				runRepo := &mockAgentRunRepo{
					getByIDErr: sql.ErrNoRows,
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orchRepo, runRepo := tt.setupMock()

			var orch *service.Orchestrator
			if !tt.useNilOrch {
				fakeLLM := &mockLLM{}
				orch = service.NewOrchestrator(orchRepo, fakeLLM, tool.ToolOptions{})
			}

			handler := NewAgentRunHandler(runRepo, orch)

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.id+"/start", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handler.StartAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestResumeAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupMock      func() (*mockOrchestratorRepo, *mockAgentRunRepo)
		useNilOrch     bool
		expectedStatus int
	}{
		{
			name: "success - paused run",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "paused",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{
						{
							ID:        testUUID,
							TaskID:    testUUID2,
							Status:    "running",
							Goal:      "Test goal",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			id:             "not-a-uuid",
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nil orchestrator",
			id:             testUUID,
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			useNilOrch:     true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "orchestrator error - not paused",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "completed",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{*orchRepo.run},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "run not found",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.returnErr = sql.ErrNoRows
				runRepo := &mockAgentRunRepo{
					getByIDErr: sql.ErrNoRows,
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orchRepo, runRepo := tt.setupMock()

			var orch *service.Orchestrator
			if !tt.useNilOrch {
				fakeLLM := &mockLLM{}
				orch = service.NewOrchestrator(orchRepo, fakeLLM, tool.ToolOptions{})
			}

			handler := NewAgentRunHandler(runRepo, orch)

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.id+"/resume", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handler.ResumeAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestCancelAgentRun(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupMock      func() (*mockOrchestratorRepo, *mockAgentRunRepo)
		useNilOrch     bool
		expectedStatus int
	}{
		{
			name: "success",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "running",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{
						{
							ID:        testUUID,
							TaskID:    testUUID2,
							Status:    "cancelled",
							Goal:      "Test goal",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			id:             "not-a-uuid",
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nil orchestrator",
			id:             testUUID,
			setupMock:      func() (*mockOrchestratorRepo, *mockAgentRunRepo) { return newMockOrchestratorRepo(), &mockAgentRunRepo{} },
			useNilOrch:     true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "orchestrator error - already terminal",
			id:   testUUID,
			setupMock: func() (*mockOrchestratorRepo, *mockAgentRunRepo) {
				orchRepo := newMockOrchestratorRepo()
				orchRepo.run = &model.AgentRun{
					ID:        testUUID,
					TaskID:    testUUID2,
					Status:    "completed",
					Goal:      "Test goal",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				runRepo := &mockAgentRunRepo{
					runs: []model.AgentRun{*orchRepo.run},
				}
				return orchRepo, runRepo
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orchRepo, runRepo := tt.setupMock()

			var orch *service.Orchestrator
			if !tt.useNilOrch {
				fakeLLM := &mockLLM{}
				orch = service.NewOrchestrator(orchRepo, fakeLLM, tool.ToolOptions{})
			}

			handler := NewAgentRunHandler(runRepo, orch)

			req := httptest.NewRequest("POST", "/agent-runs/"+tt.id+"/cancel", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handler.CancelAgentRun(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
