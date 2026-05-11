package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
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
	getRunErr            error
	getTeamMembersErr    error
	getProfileErr        error
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
			orch := NewOrchestrator(repo, fakeLLM)

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
			orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, blockingLLM)

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
			orch := NewOrchestrator(repo, fakeLLM)

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
			orch := NewOrchestrator(repo, fakeLLM)

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

func TestNewOrchestrator_InitializesRunningMap(t *testing.T) {
	repo := newMockOrchestratorRepo()
	fakeLLM := llm.NewFakeProvider()
	orch := NewOrchestrator(repo, fakeLLM)

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
	
	orch1 := NewOrchestrator(repo1, fakeLLM)
	orch2 := NewOrchestrator(repo2, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, fakeLLM)

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
	orch := NewOrchestrator(repo, blockingLLM)

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
		orch := NewOrchestrator(repo, fakeLLM)

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
		orch := NewOrchestrator(repo, fakeLLM)

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
