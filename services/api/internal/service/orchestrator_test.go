package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/tool"
)

const (
	testRunID     = "550e8400-e29b-41d4-a716-446655440001"
	testRunID2    = "550e8400-e29b-41d4-a716-446655440002"
	testTeamID    = "550e8400-e29b-41d4-a716-446655440010"
	testProfileID = "550e8400-e29b-41d4-a716-446655440020"
	testStepID    = "550e8400-e29b-41d4-a716-446655440030"
	testStepID2   = "550e8400-e29b-41d4-a716-446655440031"
)

type mockOrchestratorRepo struct {
	run                  *model.AgentRun
	steps                []model.AgentRunStep
	members              []model.AgentTeamMember
	profiles             map[string]*model.AgentProfile
	task                 *model.Task
	getRunErr            error
	getTeamMembersErr    error
	getProfileErr        error
	getTaskErr           error
	createStepErr        error
	getStepsErr          error
	updateStatusErr      error
	updateSummaryErr     error
	markStepStartedErr   error
	markStepCompletedErr error
	markStepFailedErr    error
	createMessageErr     error

	createdSteps  []model.AgentRunStep
	statusHistory []string
}

func newMockOrchestratorRepo() *mockOrchestratorRepo {
	return &mockOrchestratorRepo{
		profiles: make(map[string]*model.AgentProfile),
	}
}

func (m *mockOrchestratorRepo) GetRunByID(id string) (*model.AgentRun, error) {
	if m.getRunErr != nil {
		return nil, m.getRunErr
	}
	if m.run == nil {
		return nil, errors.New("not found")
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) UpdateRunStatus(id string, status string) (*model.AgentRun, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	m.statusHistory = append(m.statusHistory, status)
	if m.run != nil {
		m.run.Status = status
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) GetTeamMembers(teamID string) ([]model.AgentTeamMember, error) {
	if m.getTeamMembersErr != nil {
		return nil, m.getTeamMembersErr
	}
	return m.members, nil
}

func (m *mockOrchestratorRepo) GetProfileByID(id string) (*model.AgentProfile, error) {
	if m.getProfileErr != nil {
		return nil, m.getProfileErr
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
	if m.createStepErr != nil {
		return m.createStepErr
	}
	step.ID = testStepID + "-" + string(rune(len(m.createdSteps)+48))
	m.createdSteps = append(m.createdSteps, *step)
	return nil
}

func (m *mockOrchestratorRepo) GetStepsByRunID(runID string) ([]model.AgentRunStep, error) {
	if m.getStepsErr != nil {
		return nil, m.getStepsErr
	}
	return m.steps, nil
}

func (m *mockOrchestratorRepo) MarkStepStarted(id string) (*model.AgentRunStep, error) {
	if m.markStepStartedErr != nil {
		return nil, m.markStepStartedErr
	}
	for i := range m.steps {
		if m.steps[i].ID == id {
			now := time.Now()
			m.steps[i].Status = "in_progress"
			m.steps[i].StartedAt = &now
			return &m.steps[i], nil
		}
	}
	return &model.AgentRunStep{ID: id, Status: "in_progress"}, nil
}

func (m *mockOrchestratorRepo) MarkStepCompleted(id string, output string) (*model.AgentRunStep, error) {
	if m.markStepCompletedErr != nil {
		return nil, m.markStepCompletedErr
	}
	for i := range m.steps {
		if m.steps[i].ID == id {
			now := time.Now()
			m.steps[i].Status = "completed"
			m.steps[i].Output = output
			m.steps[i].CompletedAt = &now
			return &m.steps[i], nil
		}
	}
	return &model.AgentRunStep{ID: id, Status: "completed", Output: output}, nil
}

func (m *mockOrchestratorRepo) MarkStepFailed(id string, output string) (*model.AgentRunStep, error) {
	if m.markStepFailedErr != nil {
		return nil, m.markStepFailedErr
	}
	for i := range m.steps {
		if m.steps[i].ID == id {
			m.steps[i].Status = "failed"
			m.steps[i].Output = output
			return &m.steps[i], nil
		}
	}
	return &model.AgentRunStep{ID: id, Status: "failed", Output: output}, nil
}

func (m *mockOrchestratorRepo) UpdateRunStatusIfIn(id string, newStatus string, allowedStatuses []string) (*model.AgentRun, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
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

func (m *mockOrchestratorRepo) CreateMessage(message *model.AgentMessage) error {
	return m.createMessageErr
}

func (m *mockOrchestratorRepo) UpdateRunSummary(id string, summary string) (*model.AgentRun, error) {
	if m.updateSummaryErr != nil {
		return nil, m.updateSummaryErr
	}
	if m.run != nil {
		m.run.Summary = summary
	}
	return m.run, nil
}

func (m *mockOrchestratorRepo) GetTaskByID(id string) (*model.Task, error) {
	if m.getTaskErr != nil {
		return nil, m.getTaskErr
	}
	return m.task, nil
}

func (m *mockOrchestratorRepo) CreateToolCall(toolCall *model.AgentToolCall) error {
	return nil
}

func (m *mockOrchestratorRepo) UpdateToolCallStatus(id string, status string) (*model.AgentToolCall, error) {
	return &model.AgentToolCall{ID: id, Status: status}, nil
}

func (m *mockOrchestratorRepo) CreateApproval(approval *model.HumanApproval) error {
	return nil
}

func (m *mockOrchestratorRepo) GetApprovedToolApprovals(runID, stepID string) ([]model.HumanApproval, error) {
	return nil, nil
}

func (m *mockOrchestratorRepo) UpdateToolCallOutput(id string, output []byte, status string) (*model.AgentToolCall, error) {
	return &model.AgentToolCall{ID: id, Status: status, Output: output}, nil
}

type blockingFakeLLM struct {
	llm.LLMProvider
	blockChan chan struct{}
}

func newBlockingFakeLLM() *blockingFakeLLM {
	return &blockingFakeLLM{
		blockChan: make(chan struct{}),
	}
}

func (b *blockingFakeLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-b.blockChan:
		return "blocked response", nil
	}
}

func (b *blockingFakeLLM) ChatCompletionWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition, toolChoice string) (*llm.ChatCompletionResponse, error) {
	content, err := b.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, err
	}
	return &llm.ChatCompletionResponse{Content: content}, nil
}

func (b *blockingFakeLLM) unblock() {
	close(b.blockChan)
}

func TestMapMemberRoleToStepType(t *testing.T) {
	tests := []struct {
		role         string
		expectedType string
	}{
		{"planner", "plan"},
		{"PLANNER", "plan"},
		{"Planner", "plan"},
		{"plan", "plan"},
		{"implementer", "implement"},
		{"backend", "implement"},
		{"frontend", "implement"},
		{"database", "implement"},
		{"reviewer", "review"},
		{"review", "review"},
		{"qa", "test"},
		{"test", "test"},
		{"unknown", "implement"},
		{"", "implement"},
	}

	for _, tt := range tests {
		t.Run("role:"+tt.role, func(t *testing.T) {
			got := mapMemberRoleToStepType(tt.role)
			if got != tt.expectedType {
				t.Errorf("mapMemberRoleToStepType(%q) = %q, want %q", tt.role, got, tt.expectedType)
			}
		})
	}
}

func TestStartRun_ValidatesStartableStatuses(t *testing.T) {
	startableStatuses := []string{"draft", "planned", "waiting_approval", "approved"}
	nonStartableStatuses := []string{"running", "completed", "failed", "cancelled", "paused"}

	for _, status := range startableStatuses {
		t.Run("startable:"+status, func(t *testing.T) {
			repo := newMockOrchestratorRepo()
			repo.run = &model.AgentRun{
				ID:     testRunID,
				Status: status,
			}

			fakeLLM := llm.NewFakeProvider()
			orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

			err := orch.StartRun(context.Background(), testRunID)
			if err != nil {
				t.Errorf("StartRun with status %q should succeed, got: %v", status, err)
			}
		})
	}

	for _, status := range nonStartableStatuses {
		t.Run("non-startable:"+status, func(t *testing.T) {
			repo := newMockOrchestratorRepo()
			repo.run = &model.AgentRun{
				ID:     testRunID,
				Status: status,
			}

			fakeLLM := llm.NewFakeProvider()
			orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

			err := orch.StartRun(context.Background(), testRunID)
			if err == nil {
				t.Errorf("StartRun with status %q should return error", status)
			}
		})
	}
}

func TestStartRun_CreatesStepsFromTeam(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TeamID: testTeamID,
		Status: "draft",
		Goal:   "Test goal",
	}
	repo.steps = []model.AgentRunStep{}
	repo.members = []model.AgentTeamMember{
		{
			ID:        "member-1",
			TeamID:    testTeamID,
			ProfileID: testProfileID,
			MemberRole: "planner",
			Position:  1,
		},
		{
			ID:        "member-2",
			TeamID:    testTeamID,
			ProfileID: testProfileID + "2",
			MemberRole: "implementer",
			Position:  2,
		},
	}
	repo.profiles[testProfileID] = &model.AgentProfile{
		ID:           testProfileID,
		Name:         "Planner Agent",
		Role:         "planner",
		SystemPrompt: "You plan things.",
	}
	repo.profiles[testProfileID+"2"] = &model.AgentProfile{
		ID:           testProfileID + "2",
		Name:         "Implementer Agent",
		Role:         "implementer",
		SystemPrompt: "You implement things.",
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err := orch.StartRun(context.Background(), testRunID)
	if err != nil {
		t.Fatalf("StartRun failed: %v", err)
	}

	if len(repo.createdSteps) != 2 {
		t.Errorf("Expected 2 steps to be created, got %d", len(repo.createdSteps))
	}

	if len(repo.createdSteps) >= 1 {
		if repo.createdSteps[0].StepType != "plan" {
			t.Errorf("First step type should be 'plan', got %q", repo.createdSteps[0].StepType)
		}
		if repo.createdSteps[0].Position != 1 {
			t.Errorf("First step position should be 1, got %d", repo.createdSteps[0].Position)
		}
	}

	if len(repo.createdSteps) >= 2 {
		if repo.createdSteps[1].StepType != "implement" {
			t.Errorf("Second step type should be 'implement', got %q", repo.createdSteps[1].StepType)
		}
		if repo.createdSteps[1].Position != 2 {
			t.Errorf("Second step position should be 2, got %d", repo.createdSteps[1].Position)
		}
	}
}

func TestStartRun_DoesNotCreateDuplicateSteps(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TeamID: testTeamID,
		Status: "draft",
		Goal:   "Test goal",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err := orch.StartRun(context.Background(), testRunID)
	if err != nil {
		t.Fatalf("StartRun failed: %v", err)
	}

	if len(repo.createdSteps) != 0 {
		t.Errorf("Expected 0 steps to be created (already have steps), got %d", len(repo.createdSteps))
	}
}

func TestExecuteRun_NoStepsCompletesImmediately(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	if len(repo.statusHistory) == 0 {
		t.Error("Expected at least one status update")
	} else {
		lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
		if lastStatus != "completed" {
			t.Errorf("Expected last status to be 'completed', got %q", lastStatus)
		}
	}
}

func TestExecuteRun_CancelledBeforeExecution(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	orch.executeRun(ctx, testRunID)

	if len(repo.statusHistory) == 0 {
		t.Error("Expected at least one status update")
	} else {
		lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
		if lastStatus != "cancelled" {
			t.Errorf("Expected last status to be 'cancelled', got %q. History: %v", lastStatus, repo.statusHistory)
		}
	}
}

func TestExecuteRun_CancelledDuringStep(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	blockingLLM := newBlockingFakeLLM()
	orch := NewOrchestrator(repo, blockingLLM, tool.ToolOptions{})

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		orch.executeRun(ctx, testRunID)
	}()

	cancel()
	wg.Wait()

	if len(repo.statusHistory) == 0 {
		t.Error("Expected at least one status update")
	} else {
		lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
		if lastStatus != "cancelled" {
			t.Errorf("Expected last status to be 'cancelled', got %q. History: %v", lastStatus, repo.statusHistory)
		}
	}
}

func TestCancelRun_TerminalStatesRejected(t *testing.T) {
	terminalStates := []string{"completed", "failed", "cancelled"}

	for _, status := range terminalStates {
		t.Run("terminal:"+status, func(t *testing.T) {
			repo := newMockOrchestratorRepo()
			repo.run = &model.AgentRun{
				ID:     testRunID,
				Status: status,
			}

			fakeLLM := llm.NewFakeProvider()
			orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

			err := orch.CancelRun(testRunID)
			if err == nil {
				t.Errorf("CancelRun with status %q should return error", status)
			}
		})
	}
}

func TestCancelRun_NonTerminalCallsCancelAndUpdatesStatus(t *testing.T) {
	nonTerminalStates := []string{"draft", "planned", "waiting_approval", "approved", "running", "paused"}

	for _, status := range nonTerminalStates {
		t.Run("non-terminal:"+status, func(t *testing.T) {
			repo := newMockOrchestratorRepo()
			repo.run = &model.AgentRun{
				ID:     testRunID,
				Status: status,
			}

			fakeLLM := llm.NewFakeProvider()
			orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

			ctx, cancel := context.WithCancel(context.Background())
			orch.mu.Lock()
			orch.running[testRunID] = cancel
			orch.mu.Unlock()

			select {
			case <-ctx.Done():
				t.Error("context should not be cancelled yet")
			default:
			}

			err := orch.CancelRun(testRunID)
			if err != nil {
				t.Errorf("CancelRun with status %q should succeed, got: %v", status, err)
			}

			select {
			case <-ctx.Done():
			default:
				t.Error("context should be cancelled after CancelRun")
			}

			if repo.run.Status != "cancelled" {
				t.Errorf("Expected status to be 'cancelled', got %q", repo.run.Status)
			}

			orch.mu.Lock()
			_, stillRunning := orch.running[testRunID]
			orch.mu.Unlock()

			if stillRunning {
				t.Error("run should be removed from running map after cancel")
			}
		})
	}
}

func TestStartRun_NoStepsCreatedOnFailedTransition(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TeamID: testTeamID,
		Status: "running",
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err := orch.StartRun(context.Background(), testRunID)
	if err == nil {
		t.Error("StartRun should fail when run is already running")
	}

	if len(repo.createdSteps) != 0 {
		t.Errorf("Expected 0 steps created on failed transition, got %d", len(repo.createdSteps))
	}

	orch.mu.Lock()
	_, exists := orch.running[testRunID]
	orch.mu.Unlock()
	if exists {
		t.Error("Running map should NOT be populated when transition fails")
	}
}

func TestStartRun_CleansUpOnFailedStepCreation(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TeamID: testTeamID,
		Status: "draft",
		Goal:   "Test goal",
	}
	repo.steps = []model.AgentRunStep{}
	repo.members = []model.AgentTeamMember{
		{
			ID:         "member-1",
			TeamID:     testTeamID,
			ProfileID:  testProfileID,
			MemberRole: "planner",
			Position:   1,
		},
	}
	repo.profiles[testProfileID] = &model.AgentProfile{
		ID:           testProfileID,
		Name:         "Planner Agent",
		Role:         "planner",
		SystemPrompt: "You plan things.",
	}
	repo.createStepErr = errors.New("db error")

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err := orch.StartRun(context.Background(), testRunID)
	if err == nil {
		t.Fatal("StartRun should fail when createStepsFromTeam fails")
	}

	if repo.run.Status != "failed" {
		t.Errorf("Expected run status to be 'failed', got %q", repo.run.Status)
	}

	if !strings.Contains(repo.run.Summary, "failed to create steps") {
		t.Errorf("Summary should mention step creation failure, got: %q", repo.run.Summary)
	}

	orch.mu.Lock()
	_, exists := orch.running[testRunID]
	orch.mu.Unlock()
	if exists {
		t.Error("Running map should be empty after cleanup")
	}

	foundRunning := false
	foundFailed := false
	for _, s := range repo.statusHistory {
		if s == "running" {
			foundRunning = true
		}
		if s == "failed" {
			foundFailed = true
		}
	}
	if !foundRunning {
		t.Error("Status history should contain 'running'")
	}
	if !foundFailed {
		t.Error("Status history should contain 'failed'")
	}
}

func TestStartRun_SimulatedConcurrentStart(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TeamID: testTeamID,
		Status: "draft",
		Goal:   "Test goal",
	}
	repo.steps = []model.AgentRunStep{}
	repo.members = []model.AgentTeamMember{
		{
			ID:         "member-1",
			TeamID:     testTeamID,
			ProfileID:  testProfileID,
			MemberRole: "planner",
			Position:   1,
		},
	}
	repo.profiles[testProfileID] = &model.AgentProfile{
		ID:           testProfileID,
		Name:         "Planner Agent",
		Role:         "planner",
		SystemPrompt: "You plan things.",
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err1 := orch.StartRun(context.Background(), testRunID)
	if err1 != nil {
		t.Fatalf("First StartRun should succeed: %v", err1)
	}

	err2 := orch.StartRun(context.Background(), testRunID)
	if err2 == nil {
		t.Error("Second StartRun should fail when run is already running")
	}

	if len(repo.createdSteps) != 1 {
		t.Errorf("Expected exactly 1 step created, got %d", len(repo.createdSteps))
	}

	runCount := 0
	for _, s := range repo.statusHistory {
		if s == "running" {
			runCount++
		}
	}
	if runCount != 1 {
		t.Errorf("Expected exactly 1 'running' transition, got %d. History: %v", runCount, repo.statusHistory)
	}
}

func TestStartRun_HappyPathNoTeam(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "draft",
	}
	repo.steps = []model.AgentRunStep{}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	err := orch.StartRun(context.Background(), testRunID)
	if err != nil {
		t.Fatalf("StartRun should succeed: %v", err)
	}

	orch.mu.Lock()
	_, exists := orch.running[testRunID]
	orch.mu.Unlock()
	if !exists {
		t.Error("Running map should contain runID")
	}

	if len(repo.createdSteps) != 0 {
		t.Errorf("Expected 0 steps created, got %d", len(repo.createdSteps))
	}

	if repo.run.Status != "running" {
		t.Errorf("Expected status 'running', got %q", repo.run.Status)
	}
}

func TestNewOrchestrator_InitializesRunningMap(t *testing.T) {
	repo := newMockOrchestratorRepo()
	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	if orch.running == nil {
		t.Error("running map should be initialized")
	}

	if len(orch.running) != 0 {
		t.Error("running map should be empty initially")
	}
}

func TestMultipleRuns_CanRunConcurrently(t *testing.T) {
	repo1 := newMockOrchestratorRepo()
	repo1.run = &model.AgentRun{
		ID:     testRunID,
		Status: "draft",
	}
	repo1.steps = []model.AgentRunStep{}

	repo2 := newMockOrchestratorRepo()
	repo2.run = &model.AgentRun{
		ID:     testRunID2,
		Status: "draft",
	}
	repo2.steps = []model.AgentRunStep{}

	fakeLLM := llm.NewFakeProvider()
	
	orch1 := NewOrchestrator(repo1, fakeLLM, tool.ToolOptions{})
	orch2 := NewOrchestrator(repo2, fakeLLM, tool.ToolOptions{})

	err1 := orch1.StartRun(context.Background(), testRunID)
	err2 := orch2.StartRun(context.Background(), testRunID2)

	if err1 != nil {
		t.Errorf("First run start failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second run start failed: %v", err2)
	}
}

func TestExecuteRun_CompletedStepsNotRerun(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	now := time.Now()
	repo.steps = []model.AgentRunStep{
		{
			ID:          testStepID,
			RunID:       testRunID,
			ProfileID:   testProfileID,
			StepType:    "plan",
			Status:      "completed",
			Position:    1,
			Output:      "Existing completed output",
			CompletedAt: &now,
		},
		{
			ID:          testStepID2,
			RunID:       testRunID,
			ProfileID:   testProfileID,
			StepType:    "implement",
			Status:      "pending",
			Position:    2,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("Expected final status 'completed', got %q", lastStatus)
	}

	if repo.steps[0].Status != "completed" {
		t.Errorf("Completed step should remain 'completed', got %q", repo.steps[0].Status)
	}

	if repo.steps[1].Status != "completed" {
		t.Errorf("Pending step should now be 'completed', got %q", repo.steps[1].Status)
	}
}

func TestExecuteRun_SkippedStepsNotRerun(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "skipped",
			Position:  1,
		},
		{
			ID:        testStepID2,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "implement",
			Status:    "pending",
			Position:  2,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("Expected final status 'completed', got %q", lastStatus)
	}

	if repo.steps[0].Status != "skipped" {
		t.Errorf("Skipped step should remain 'skipped', got %q", repo.steps[0].Status)
	}

	if repo.steps[1].Status != "completed" {
		t.Errorf("Pending step should now be 'completed', got %q", repo.steps[1].Status)
	}
}

func TestExecuteRun_ExistingFailedStepFailsRun(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	now := time.Now()
	repo.steps = []model.AgentRunStep{
		{
			ID:          testStepID,
			RunID:       testRunID,
			ProfileID:   testProfileID,
			StepType:    "plan",
			Status:      "failed",
			Position:    1,
			Output:      "Previous failure",
			CompletedAt: &now,
		},
		{
			ID:        testStepID2,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "implement",
			Status:    "pending",
			Position:  2,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "failed" {
		t.Errorf("Expected final status 'failed', got %q", lastStatus)
	}

	if repo.steps[1].Status != "pending" {
		t.Errorf("Pending step should NOT have been executed, got %q", repo.steps[1].Status)
	}

	if !strings.Contains(repo.run.Summary, "existing failed step") {
		t.Errorf("Summary should mention existing failed step, got: %q", repo.run.Summary)
	}
}

func TestExecuteRun_NoPendingStepsCompletes(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	now := time.Now()
	repo.steps = []model.AgentRunStep{
		{
			ID:          testStepID,
			RunID:       testRunID,
			ProfileID:   testProfileID,
			StepType:    "plan",
			Status:      "completed",
			Position:    1,
			Output:      "Step 1 output",
			CompletedAt: &now,
		},
		{
			ID:          testStepID2,
			RunID:       testRunID,
			ProfileID:   testProfileID,
			StepType:    "review",
			Status:      "skipped",
			Position:    2,
			CompletedAt: &now,
		},
	}

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("Expected final status 'completed', got %q", lastStatus)
	}

	if !strings.Contains(repo.run.Summary, "Completed: 1") {
		t.Errorf("Summary should mention Completed count, got: %q", repo.run.Summary)
	}
	if !strings.Contains(repo.run.Summary, "Skipped: 1") {
		t.Errorf("Summary should mention Skipped count, got: %q", repo.run.Summary)
	}
}

func TestExecuteRun_CancelledDuringLLM_MarksCancelled(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	blockingLLM := newBlockingFakeLLM()
	orch := NewOrchestrator(repo, blockingLLM, tool.ToolOptions{})

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		orch.executeRun(ctx, testRunID)
	}()

	<-time.After(50 * time.Millisecond)
	cancel()
	wg.Wait()

	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "cancelled" {
		t.Errorf("Expected final status 'cancelled', got %q. History: %v", lastStatus, repo.statusHistory)
	}

	if !strings.Contains(repo.run.Summary, "Run cancelled") {
		t.Errorf("Summary should say 'Run cancelled.', got: %q", repo.run.Summary)
	}
}

func TestStartRun_AtomicStatusTransition(t *testing.T) {
	t.Run("fails when status not allowed", func(t *testing.T) {
		repo := newMockOrchestratorRepo()
		repo.run = &model.AgentRun{
			ID:     testRunID,
			Status: "running",
		}

		fakeLLM := llm.NewFakeProvider()
		orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

		err := orch.StartRun(context.Background(), testRunID)
		if err == nil {
			t.Error("StartRun should fail when run is already running")
		}

		orch.mu.Lock()
		_, exists := orch.running[testRunID]
		orch.mu.Unlock()
		if exists {
			t.Error("Running map should NOT be populated when transition fails")
		}
	})

	t.Run("succeeds when status is allowed", func(t *testing.T) {
		repo := newMockOrchestratorRepo()
		repo.run = &model.AgentRun{
			ID:     testRunID,
			Status: "draft",
		}
		repo.steps = []model.AgentRunStep{}

		fakeLLM := llm.NewFakeProvider()
		orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

		err := orch.StartRun(context.Background(), testRunID)
		if err != nil {
			t.Fatalf("StartRun should succeed for draft status: %v", err)
		}

		orch.mu.Lock()
		cancel, exists := orch.running[testRunID]
		orch.mu.Unlock()
		if !exists {
			t.Error("Running map SHOULD be populated when transition succeeds")
		}

		cancel()
		<-time.After(50 * time.Millisecond)

		orch.mu.Lock()
		delete(orch.running, testRunID)
		orch.mu.Unlock()
	})
}

type capturingLLM struct {
	llm.LLMProvider
	messages []llm.Message
}

func (c *capturingLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	c.messages = messages
	return "captured response", nil
}

func (c *capturingLLM) ChatCompletionWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition, toolChoice string) (*llm.ChatCompletionResponse, error) {
	c.messages = messages
	return &llm.ChatCompletionResponse{Content: "captured response"}, nil
}

func TestExecuteRun_IncludesTaskContextInMessages(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TaskID: "task-123",
		Status: "running",
		Goal:   "Build a web app",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:           testStepID,
			RunID:        testRunID,
			ProfileID:    testProfileID,
			StepType:     "plan",
			Status:       "pending",
			Position:     1,
			Instructions: "Plan the architecture",
		},
	}
	repo.task = &model.Task{
		ID:          "task-123",
		Title:       "Test Task",
		Description: "A test task description",
		Plan:        "Step 1, Step 2",
		ReviewNotes: "Needs review",
	}

	capLLM := &capturingLLM{}
	orch := NewOrchestrator(repo, capLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	if len(capLLM.messages) < 2 {
		t.Fatal("Expected at least 2 messages (system + user)")
	}

	userMsg := capLLM.messages[len(capLLM.messages)-1]
	if !strings.Contains(userMsg.Content, "=== TASK CONTEXT ===") {
		t.Error("Expected task context section in user message")
	}
	if !strings.Contains(userMsg.Content, "Test Task") {
		t.Error("Expected task title in user message")
	}
	if !strings.Contains(userMsg.Content, "A test task description") {
		t.Error("Expected task description in user message")
	}
	if !strings.Contains(userMsg.Content, "Step 1, Step 2") {
		t.Error("Expected task plan in user message")
	}
	if !strings.Contains(userMsg.Content, "Needs review") {
		t.Error("Expected review notes in user message")
	}
	if !strings.Contains(userMsg.Content, "Build a web app") {
		t.Error("Expected run goal in user message")
	}
	if !strings.Contains(userMsg.Content, "Plan the architecture") {
		t.Error("Expected step instructions in user message")
	}
}

func TestExecuteRun_TaskLoadFailureFailsRun(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TaskID: "task-123",
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}
	repo.getTaskErr = errors.New("db error")

	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	if len(repo.statusHistory) == 0 {
		t.Fatal("Expected at least one status update")
	}
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "failed" {
		t.Errorf("Expected final status 'failed', got %q", lastStatus)
	}
	if !strings.Contains(repo.run.Summary, "failed to load task context") {
		t.Errorf("Summary should mention task load failure, got: %q", repo.run.Summary)
	}
}

func TestExecuteRun_EmptyTaskIDSkipsTaskContext(t *testing.T) {
	repo := newMockOrchestratorRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		TaskID: "",
		Status: "running",
		Goal:   "Build something",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:           testStepID,
			RunID:        testRunID,
			ProfileID:    testProfileID,
			StepType:     "plan",
			Status:       "pending",
			Position:     1,
			Instructions: "Do the thing",
		},
	}

	capLLM := &capturingLLM{}
	orch := NewOrchestrator(repo, capLLM, tool.ToolOptions{})

	orch.executeRun(context.Background(), testRunID)

	if len(capLLM.messages) < 2 {
		t.Fatal("Expected at least 2 messages (system + user)")
	}

	userMsg := capLLM.messages[len(capLLM.messages)-1]
	if strings.Contains(userMsg.Content, "=== TASK CONTEXT ===") {
		t.Error("Should NOT include task context when TaskID is empty")
	}
	if !strings.Contains(userMsg.Content, "Instructions:") {
		t.Error("Should use old prompt format when TaskID is empty")
	}

	if len(repo.statusHistory) == 0 {
		t.Fatal("Expected at least one status update")
	}
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("Expected run to complete successfully, got status: %q", lastStatus)
	}
}

// captureMockRepo records created tool calls, messages, and approvals for verification
type captureMockRepo struct {
	*mockOrchestratorRepo
	createdToolCalls     []*model.AgentToolCall
	createdMessages      []*model.AgentMessage
	createdApprovals     []*model.HumanApproval
	toolCallCounter      int
}

func newCaptureMockRepo() *captureMockRepo {
	return &captureMockRepo{
		mockOrchestratorRepo: newMockOrchestratorRepo(),
	}
}

func (m *captureMockRepo) CreateToolCall(toolCall *model.AgentToolCall) error {
	m.toolCallCounter++
	toolCall.ID = fmt.Sprintf("tc-%d", m.toolCallCounter)
	m.createdToolCalls = append(m.createdToolCalls, toolCall)
	return nil
}

func (m *captureMockRepo) CreateMessage(message *model.AgentMessage) error {
	m.createdMessages = append(m.createdMessages, message)
	return nil
}

func (m *captureMockRepo) CreateApproval(approval *model.HumanApproval) error {
	approval.ID = fmt.Sprintf("approval-%d", len(m.createdApprovals)+1)
	m.createdApprovals = append(m.createdApprovals, approval)
	return nil
}

// toolResponseLLM returns responses in sequence across multiple calls.
// First call returns the first response, second call returns the second, etc.
type toolResponseLLM struct {
	llm.LLMProvider
	responses []string
	callCount int
}

func newToolResponseLLM(responses ...string) *toolResponseLLM {
	return &toolResponseLLM{responses: responses}
}

func (t *toolResponseLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	if t.callCount >= len(t.responses) {
		return "No more responses configured.", nil
	}
	resp := t.responses[t.callCount]
	t.callCount++
	return resp, nil
}

func (t *toolResponseLLM) ChatCompletionWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition, toolChoice string) (*llm.ChatCompletionResponse, error) {
	if t.callCount >= len(t.responses) {
		return &llm.ChatCompletionResponse{Content: "No more responses configured."}, nil
	}
	resp := t.responses[t.callCount]
	t.callCount++
	return &llm.ChatCompletionResponse{Content: resp}, nil
}

func TestExecuteRun_ToolCallsCreateAgentToolCallRecords(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	// LLM returns tool calls first, then a normal response to terminate the loop
	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"list_files","input":{"path":"."}}]}`,
		`Listed files successfully. Here is the directory structure.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{RequireApproval: []string{}})
	orch.executeRun(context.Background(), testRunID)

	// Verify tool call was created
	if len(repo.createdToolCalls) != 1 {
		t.Fatalf("expected 1 tool call created, got %d", len(repo.createdToolCalls))
	}

	tc := repo.createdToolCalls[0]
	if tc.ToolName != "list_files" {
		t.Errorf("expected tool_name 'list_files', got %q", tc.ToolName)
	}
	if tc.RunID != testRunID {
		t.Errorf("expected run_id %q, got %q", testRunID, tc.RunID)
	}
	if tc.Status != "completed" {
		t.Errorf("expected tool call status 'completed', got %q", tc.Status)
	}
}

func TestExecuteRun_ToolResultCreatesToolMessage(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"list_files","input":{"path":"."}}]}`,
		`Directory listed. Analysis complete.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{})
	orch.executeRun(context.Background(), testRunID)

	// Verify tool message was created (should have role "tool")
	toolMessages := 0
	for _, msg := range repo.createdMessages {
		if msg.Role == "tool" {
			toolMessages++
		}
	}
	if toolMessages != 1 {
		t.Errorf("expected 1 tool message, got %d. All messages: %+v", toolMessages, repo.createdMessages)
	}
}

func TestExecuteRun_FailedToolCallDoesNotCrash(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	// Unknown tool should fail gracefully; then return normal text to terminate loop
	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"unknown_tool","input":{"path":"."}}]}`,
		`Completed processing with some tool errors.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{})
	orch.executeRun(context.Background(), testRunID)

	// Verify tool call was created with failed status
	if len(repo.createdToolCalls) != 1 {
		t.Fatalf("expected 1 tool call created, got %d", len(repo.createdToolCalls))
	}
	if repo.createdToolCalls[0].Status != "failed" {
		t.Errorf("expected failed tool call status, got %q", repo.createdToolCalls[0].Status)
	}

	// Run should still complete (tool error != run failure)
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("expected run to complete despite tool error, last status: %q", lastStatus)
	}
}

// messagesRecorderLLM records all messages passed to ChatCompletion for verification.
type messagesRecorderLLM struct {
	responses  []string
	callCount  int
	allCalls   [][]llm.Message
}

func newMessagesRecorderLLM(responses ...string) *messagesRecorderLLM {
	return &messagesRecorderLLM{responses: responses}
}

func (m *messagesRecorderLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	// Record a copy of the messages
	msgsCopy := make([]llm.Message, len(messages))
	copy(msgsCopy, messages)
	m.allCalls = append(m.allCalls, msgsCopy)

	if m.callCount >= len(m.responses) {
		return "No more responses configured.", nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *messagesRecorderLLM) ChatCompletionWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition, toolChoice string) (*llm.ChatCompletionResponse, error) {
	msgsCopy := make([]llm.Message, len(messages))
	copy(msgsCopy, messages)
	m.allCalls = append(m.allCalls, msgsCopy)

	if m.callCount >= len(m.responses) {
		return &llm.ChatCompletionResponse{Content: "No more responses configured."}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return &llm.ChatCompletionResponse{Content: resp}, nil
}

func TestExecuteRun_MultiTurnFeedsToolResultsToLLM(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	// First call returns a tool call; second call returns a normal response.
	llm := newMessagesRecorderLLM(
		`{"tool_calls":[{"tool_name":"list_files","input":{"path":"."}}]}`,
		`Based on the directory listing, the project has 3 main directories.`,
	)

	orch := NewOrchestrator(repo, llm, tool.ToolOptions{})
	orch.executeRun(context.Background(), testRunID)

	// Verify run completed
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Fatalf("expected run to complete, last status: %q", lastStatus)
	}

	// Verify step output is the final non-tool-call response
	if len(repo.steps) > 0 && repo.steps[0].Status == "completed" {
		if !strings.Contains(repo.steps[0].Output, "3 main directories") {
			t.Errorf("step output should contain final LLM response, got: %q", repo.steps[0].Output)
		}
	}

	// Verify LLM was called twice (tool round + final)
	if len(llm.allCalls) != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", len(llm.allCalls))
	}

	// First call: system + user (initial messages, no tool results yet)
	firstCall := llm.allCalls[0]
	if len(firstCall) < 2 {
		t.Fatalf("first LLM call: expected at least 2 messages (system+user), got %d", len(firstCall))
	}
	if firstCall[0].Role != "system" {
		t.Errorf("first message role should be 'system', got %q", firstCall[0].Role)
	}
	if firstCall[1].Role != "user" {
		t.Errorf("second message role should be 'user', got %q", firstCall[1].Role)
	}

	// Second call: system + user + assistant (tool call JSON) + tool (tool result)
	secondCall := llm.allCalls[1]
	if len(secondCall) < 4 {
		t.Fatalf("second LLM call: expected at least 4 messages (system+user+assistant+tool), got %d", len(secondCall))
	}

	// Verify the assistant message from the first iteration is in context
	if secondCall[2].Role != "assistant" {
		t.Errorf("third message role should be 'assistant', got %q", secondCall[2].Role)
	}
	if !strings.Contains(secondCall[2].Content, "list_files") {
		t.Errorf("assistant message should contain tool call JSON, got: %q", secondCall[2].Content)
	}

	// Verify the tool result from the first iteration is in context
	if secondCall[3].Role != "tool" {
		t.Errorf("fourth message role should be 'tool', got %q", secondCall[3].Role)
	}
	if !strings.Contains(secondCall[3].Content, "list_files") {
		t.Errorf("tool message should contain tool result with 'list_files', got: %q", secondCall[3].Content)
	}
	if !strings.Contains(secondCall[3].Content, "success") {
		t.Errorf("tool message should contain success indicator, got: %q", secondCall[3].Content)
	}
}

func TestExecuteRun_ToolWithApprovalCreatesApprovalRecord(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"read_file","input":{"path":"test.txt"}}]}`,
		`File analysis complete.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{
		RequireApproval: []string{"read_file"},
	})
	orch.executeRun(context.Background(), testRunID)

	// Tool call should be "recorded" not "completed"
	if len(repo.createdToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(repo.createdToolCalls))
	}
	if repo.createdToolCalls[0].Status != "recorded" {
		t.Errorf("expected status 'recorded', got %q", repo.createdToolCalls[0].Status)
	}

	// Tool message should mention approval
	foundApprovalMsg := false
	for _, msg := range repo.createdMessages {
		if msg.Role == "tool" && strings.Contains(msg.Content, "requires human approval") {
			foundApprovalMsg = true
			break
		}
	}
	if !foundApprovalMsg {
		t.Error("expected tool message mentioning 'requires human approval'")
	}

	// Run should complete
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("expected run to complete, got %q", lastStatus)
	}
}

func TestExecuteRun_ToolWithoutApprovalExecutesNormally(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	// Use list_files which succeeds even without explicit workspace root
	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"list_files","input":{"path":"."}}]}`,
		`Directory listing complete.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{
		RequireApproval: []string{},
	})
	orch.executeRun(context.Background(), testRunID)

	// Tool call should be "completed" not "recorded"
	if len(repo.createdToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(repo.createdToolCalls))
	}
	if repo.createdToolCalls[0].Status != "completed" {
		t.Errorf("expected status 'completed', got %q", repo.createdToolCalls[0].Status)
	}

	// Tool message should NOT mention approval
	for _, msg := range repo.createdMessages {
		if msg.Role == "tool" && strings.Contains(msg.Content, "requires human approval") {
			t.Error("tool message should NOT mention 'requires human approval' when no approval gate")
			break
		}
	}

	// Run should complete
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("expected run to complete, got %q", lastStatus)
	}
}

func TestExecuteRun_ToolWithApprovalDoesNotFailRun(t *testing.T) {
	repo := newCaptureMockRepo()
	repo.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	repo.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	toolLLM := newToolResponseLLM(
		`{"tool_calls":[{"tool_name":"read_file","input":{"path":"test.txt"}}]}`,
		`File analysis complete.`,
	)

	orch := NewOrchestrator(repo, toolLLM, tool.ToolOptions{
		RequireApproval: []string{"read_file"},
	})
	orch.executeRun(context.Background(), testRunID)

	// Run should still complete (approval gate != run failure)
	lastStatus := repo.statusHistory[len(repo.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("expected run to complete despite approval gate, got %q", lastStatus)
	}
}

// autoResumeMockRepo extends captureMockRepo to simulate approved approvals appearing
// after the first iteration (as if a human approved the approval externally).
type autoResumeMockRepo struct {
	*captureMockRepo
	getApprovedCallCount       int
	updateToolCallOutputCalled bool
}

func (m *autoResumeMockRepo) GetApprovedToolApprovals(runID, stepID string) ([]model.HumanApproval, error) {
	m.getApprovedCallCount++
	// On the second call (and subsequent), return the stored approval with status "approved"
	// to simulate external approval having been granted between iterations.
	if m.getApprovedCallCount > 1 && len(m.createdApprovals) > 0 {
		last := m.createdApprovals[len(m.createdApprovals)-1]
		appr := *last // shallow copy
		appr.Status = "approved"
		return []model.HumanApproval{appr}, nil
	}
	return nil, nil
}

func (m *autoResumeMockRepo) UpdateToolCallOutput(id string, output []byte, status string) (*model.AgentToolCall, error) {
	m.updateToolCallOutputCalled = true
	return &model.AgentToolCall{ID: id, Status: status, Output: output}, nil
}

func TestExecuteRun_AutoResumeExecutesApprovedToolCall(t *testing.T) {
	capture := newCaptureMockRepo()
	capture.run = &model.AgentRun{
		ID:     testRunID,
		Status: "running",
	}
	capture.steps = []model.AgentRunStep{
		{
			ID:        testStepID,
			RunID:     testRunID,
			ProfileID: testProfileID,
			StepType:  "plan",
			Status:    "pending",
			Position:  1,
		},
	}

	// Create auto-resume mock that simulates approvals being granted externally
	autoResume := &autoResumeMockRepo{
		captureMockRepo: capture,
	}

	// LLM: first call returns a tool call, second returns final response
	llm := newMessagesRecorderLLM(
		`{"tool_calls":[{"tool_name":"read_file","input":{"path":"test.go"}}]}`,
		`After reviewing the file, I can proceed with the implementation.`,
	)

	orch := NewOrchestrator(autoResume, llm, tool.ToolOptions{
		RequireApproval: []string{"read_file"},
	})
	orch.executeRun(context.Background(), testRunID)

	// Run should complete
	lastStatus := capture.statusHistory[len(capture.statusHistory)-1]
	if lastStatus != "completed" {
		t.Errorf("expected run to complete, got %q", lastStatus)
	}

	// Tool call should have been created
	if len(capture.createdToolCalls) != 1 {
		t.Fatalf("expected 1 tool call created, got %d", len(capture.createdToolCalls))
	}

	// UpdateToolCallOutput should have been called (auto-resume executed the tool)
	if !autoResume.updateToolCallOutputCalled {
		t.Error("expected UpdateToolCallOutput to be called by auto-resume")
	}

	// At least one tool message should exist with the result from auto-resume
	toolMessages := 0
	for _, msg := range capture.createdMessages {
		if msg.Role == "tool" {
			toolMessages++
		}
	}
	if toolMessages < 1 {
		t.Errorf("expected at least 1 tool message, got %d", toolMessages)
	}

	// LLM should have been called twice (tool call round + final round)
	if len(llm.allCalls) != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", len(llm.allCalls))
	}

	// The second LLM call should contain the tool result from auto-resume
	secondCall := llm.allCalls[1]
	foundToolResult := false
	for _, msg := range secondCall {
		if msg.Role == "tool" && strings.Contains(msg.Content, "read_file") {
			foundToolResult = true
			break
		}
	}
	if !foundToolResult {
		t.Error("expected LLM context in second call to contain tool result with 'read_file'")
	}
}

func TestToolRegistry_WriteToolsExist(t *testing.T) {
	writeToolNames := []string{"write_file", "edit_file", "bash", "git"}

	for _, name := range writeToolNames {
		tool, err := tool.GetTool(name)
		if err != nil {
			t.Errorf("write tool %q should be registered, got error: %v", name, err)
		}
		if tool == nil {
			t.Errorf("tool %q should not be nil", name)
		}
	}
}

func TestToolRegistry_WriteToolsStillBlocked(t *testing.T) {
	blockedToolNames := []string{"apply_patch", "delete_file", "rename_file", "create_file", "mkdir", "exec_command"}

	for _, name := range blockedToolNames {
		_, err := tool.GetTool(name)
		if err == nil {
			t.Errorf("tool %q should NOT be registered", name)
		}
	}
}

func TestToolRegistry_ReadOnlyToolsExist(t *testing.T) {
	readTools := []string{"read_file", "list_files", "search_code"}

	for _, name := range readTools {
		tool, err := tool.GetTool(name)
		if err != nil {
			t.Errorf("read-only tool %q should be registered, got error: %v", name, err)
		}
		if tool == nil {
			t.Errorf("tool %q should not be nil", name)
		}
	}
}

func TestParseToolCalls_ExtractsFromJSON(t *testing.T) {
	text := `{"tool_calls":[{"tool_name":"read_file","input":{"path":"test.txt"}}]}`
	calls, err := tool.ParseToolCalls(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].ToolName != "read_file" {
		t.Errorf("expected tool_name 'read_file', got %q", calls[0].ToolName)
	}
}

func TestParseToolCalls_ExtractsFromTextBlock(t *testing.T) {
	text := `Here are the files I found:
{"tool_calls":[{"tool_name":"list_files","input":{"path":"src"}}]}
Now let me examine the main file.`
	calls, err := tool.ParseToolCalls(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].ToolName != "list_files" {
		t.Errorf("expected tool_name 'list_files', got %q", calls[0].ToolName)
	}
}

func TestParseToolCalls_ReturnsNilForNoCalls(t *testing.T) {
	text := `This is just a normal response without tool calls.`
	calls, err := tool.ParseToolCalls(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != nil {
		t.Errorf("expected nil tool calls, got %+v", calls)
	}
}

func TestToolCallJSON_IncludesToolResponse(t *testing.T) {
	// Verify that the tool output JSON contains the expected fields
	root := t.TempDir()
	tc := tool.ToolCallRequest{
		ToolName: "list_files",
		Input:    json.RawMessage(`{"path":"."}`),
	}
	opts := tool.ToolOptions{WorkspaceRoot: root}
	resp := tool.ExecuteToolCall(tc, opts)
	output := tool.FormatToolOutput(resp)

	if !strings.Contains(output, `"tool_name":"list_files"`) {
		t.Errorf("output should contain tool_name, got: %s", output)
	}
	if !strings.Contains(output, `"success":true`) {
		t.Errorf("output should indicate success, got: %s", output)
	}
}
