package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/tool"
)

type OrchestratorRepository interface {
	GetRunByID(id string) (*model.AgentRun, error)
	UpdateRunStatus(id string, status string) (*model.AgentRun, error)
	UpdateRunStatusIfIn(id string, newStatus string, allowedStatuses []string) (*model.AgentRun, error)
	GetTeamMembers(teamID string) ([]model.AgentTeamMember, error)
	GetProfileByID(id string) (*model.AgentProfile, error)
	CreateStep(step *model.AgentRunStep) error
	GetStepsByRunID(runID string) ([]model.AgentRunStep, error)
	MarkStepStarted(id string) (*model.AgentRunStep, error)
	MarkStepCompleted(id string, output string) (*model.AgentRunStep, error)
	MarkStepFailed(id string, output string) (*model.AgentRunStep, error)
	CreateMessage(message *model.AgentMessage) error
	UpdateRunSummary(id string, summary string) (*model.AgentRun, error)
	GetTaskByID(taskID string) (*model.Task, error)
	CreateToolCall(toolCall *model.AgentToolCall) error
	UpdateToolCallStatus(id string, status string) (*model.AgentToolCall, error)
	UpdateToolCallOutput(id string, output []byte, status string) (*model.AgentToolCall, error)
	CreateApproval(approval *model.HumanApproval) error
	GetApprovedToolApprovals(runID, stepID string) ([]model.HumanApproval, error)
	MarkStepWaitingApproval(id string) (*model.AgentRunStep, error)
	GetMessagesByStepID(stepID string) ([]model.AgentMessage, error)
}

type Orchestrator struct {
	repo     OrchestratorRepository
	llm      llm.LLMProvider
	running  map[string]context.CancelFunc
	mu       sync.Mutex
	toolOpts tool.ToolOptions
}

func NewOrchestrator(repo OrchestratorRepository, llmProvider llm.LLMProvider, toolOpts tool.ToolOptions) *Orchestrator {
	return &Orchestrator{
		repo:     repo,
		llm:      llmProvider,
		running:  make(map[string]context.CancelFunc),
		toolOpts: toolOpts,
	}
}

func mapMemberRoleToStepType(role string) string {
	roleLower := strings.ToLower(role)
	switch roleLower {
	case "planner", "plan":
		return "plan"
	case "implementer", "backend", "frontend", "database":
		return "implement"
	case "reviewer", "review":
		return "review"
	case "qa", "test":
		return "test"
	default:
		return "implement"
	}
}

func (o *Orchestrator) StartRun(_ context.Context, runID string) error {
	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	allowedStatuses := []string{"draft", "planned", "waiting_approval", "approved"}
	_, err = o.repo.UpdateRunStatusIfIn(runID, "running", allowedStatuses)
	if err != nil {
		return fmt.Errorf("failed to start run: %w", err)
	}

	var executionCtx context.Context
	var cancel context.CancelFunc
	executionCtx, cancel = context.WithTimeout(context.Background(), time.Duration(o.toolOpts.RunTimeout)*time.Second)

	o.mu.Lock()
	o.running[runID] = cancel
	o.mu.Unlock()

	if run.TeamID != "" {
		existingSteps, err := o.repo.GetStepsByRunID(runID)
		if err != nil {
			cancel()
			o.mu.Lock()
			delete(o.running, runID)
			o.mu.Unlock()
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			_, _ = o.repo.UpdateRunSummary(runID, "Run failed: failed to check existing steps.")
			return fmt.Errorf("failed to check existing steps: %w", err)
		}
		if len(existingSteps) == 0 {
			err = o.createStepsFromTeam(runID, run.TeamID, run.Goal)
			if err != nil {
				cancel()
				o.mu.Lock()
				delete(o.running, runID)
				o.mu.Unlock()
				_, _ = o.repo.UpdateRunStatus(runID, "failed")
				_, _ = o.repo.UpdateRunSummary(runID, "Run failed: failed to create steps from team.")
				return fmt.Errorf("failed to create steps from team: %w", err)
			}
		}
	}

	go o.executeRun(executionCtx, runID)

	return nil
}

func (o *Orchestrator) createStepsFromTeam(runID string, teamID string, runGoal string) error {
	members, err := o.repo.GetTeamMembers(teamID)
	if err != nil {
		return err
	}

	for _, member := range members {
		profile, err := o.repo.GetProfileByID(member.ProfileID)
		if err != nil {
			return err
		}

		stepType := mapMemberRoleToStepType(member.MemberRole)

		title := fmt.Sprintf("%s - %s", profile.Name, profile.Role)
		instructions := fmt.Sprintf("Run Goal: %s\n\nProfile System Prompt:\n%s", runGoal, profile.SystemPrompt)

		step := &model.AgentRunStep{
			RunID:        runID,
			ProfileID:    member.ProfileID,
			StepType:     stepType,
			Status:       "pending",
			Title:        title,
			Instructions: instructions,
			Position:     member.Position,
		}

		err = o.repo.CreateStep(step)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) executeRun(ctx context.Context, runID string) {
	defer func() {
		o.mu.Lock()
		if cancel, exists := o.running[runID]; exists {
			cancel()
			delete(o.running, runID)
		}
		o.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
		_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
		return
	default:
	}

	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		if ctx.Err() != nil {
			_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
			_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
			return
		}
		_, _ = o.repo.UpdateRunStatus(runID, "failed")
		_, _ = o.repo.UpdateRunSummary(runID, "Run failed: failed to load run context.")
		return
	}

	var task *model.Task
	if run.TaskID != "" {
		task, err = o.repo.GetTaskByID(run.TaskID)
		if err != nil {
			if ctx.Err() != nil {
				_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
				_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
				return
			}
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			_, _ = o.repo.UpdateRunSummary(runID, "Run failed: failed to load task context.")
			return
		}
	}

	steps, err := o.repo.GetStepsByRunID(runID)
	if err != nil {
		_, _ = o.repo.UpdateRunStatus(runID, "failed")
		return
	}

	if len(steps) == 0 {
		_, _ = o.repo.UpdateRunStatus(runID, "completed")
		_, _ = o.repo.UpdateRunSummary(runID, "Run completed: no steps to execute.")
		return
	}

	var pending, completed, skipped, failed []model.AgentRunStep
	for _, step := range steps {
		switch step.Status {
		case "pending":
			pending = append(pending, step)
		case "waiting_approval":
			pending = append(pending, step)
		case "completed":
			completed = append(completed, step)
		case "skipped":
			skipped = append(skipped, step)
		case "failed":
			failed = append(failed, step)
		}
	}

	if len(failed) > 0 {
		_, _ = o.repo.UpdateRunStatus(runID, "failed")
		_, _ = o.repo.UpdateRunSummary(runID, "Run failed: found existing failed step(s). Cannot continue execution.")
		return
	}

	if len(pending) == 0 {
		_, _ = o.repo.UpdateRunStatus(runID, "completed")
		summary := fmt.Sprintf("Run completed: no pending steps to execute. (Completed: %d, Skipped: %d, Failed: 0)", len(completed), len(skipped))
		_, _ = o.repo.UpdateRunSummary(runID, summary)
		return
	}

	var stepOutputs []string

	for _, step := range completed {
		stepOutputs = append(stepOutputs, step.Output)
	}

	for _, step := range pending {
		select {
		case <-ctx.Done():
			_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
			_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
			return
		default:
		}

		_, err := o.repo.MarkStepStarted(step.ID)
		if err != nil {
			if ctx.Err() != nil {
				_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
				_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
				return
			}
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		profile, err := o.repo.GetProfileByID(step.ProfileID)
		if err != nil {
			if ctx.Err() != nil {
				_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
				_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
				return
			}
			_, _ = o.repo.MarkStepFailed(step.ID, "Failed to get profile")
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		var previousOutputs string
		for i, out := range stepOutputs {
			previousOutputs += fmt.Sprintf("Step %d Output:\n%s\n\n", i+1, out)
		}

		var userContent string
		if task != nil {
			plan := task.Plan
			if plan == "" {
				plan = "Not provided"
			}
			reviewNotes := task.ReviewNotes
			if reviewNotes == "" {
				reviewNotes = "Not provided"
			}
			userContent = fmt.Sprintf(`=== TASK CONTEXT ===
Title: %s
Description: %s
Plan: %s
Review Notes: %s

=== RUN GOAL ===
%s

=== STEP INSTRUCTIONS ===
%s

=== PREVIOUS STEP OUTPUTS ===
%s`, task.Title, task.Description, plan, reviewNotes, run.Goal, step.Instructions, previousOutputs)
		} else {
			userContent = fmt.Sprintf("Instructions: %s\n\nPrevious Step Outputs:\n%s", step.Instructions, previousOutputs)
		}

		var messages []llm.Message

		if step.Status == "waiting_approval" {
			// Resume mode: reload existing messages from DB and prepend fresh system+user context
			existingMessages, loadErr := o.repo.GetMessagesByStepID(step.ID)
			if loadErr != nil {
				if ctx.Err() != nil {
					_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
					_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
					return
				}
				_, _ = o.repo.UpdateRunStatus(runID, "failed")
				return
			}
			// Fresh system+user context + existing assistant/tool messages from DB
			messages = append(messages, llm.Message{Role: "system", Content: profile.SystemPrompt})
			messages = append(messages, llm.Message{Role: "user", Content: userContent})
			for _, msg := range existingMessages {
				messages = append(messages, llm.Message{Role: msg.Role, Content: msg.Content})
			}
		} else {
			messages = []llm.Message{
				{
					Role:    "system",
					Content: profile.SystemPrompt,
				},
				{
					Role:    "user",
					Content: userContent,
				},
			}
		}

		maxIter := o.toolOpts.MaxToolIterations
		if maxIter <= 0 {
			maxIter = 10
		}

		var finalResponse string

		for iter := 0; iter < maxIter; iter++ {
			select {
			case <-ctx.Done():
				_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
				_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
				return
			default:
			}

			// === APPROVAL AUTO-RESUME CHECK ===
			// Before calling LLM, check if any previously gated tool calls have been approved
			approvedApprovals, err := o.repo.GetApprovedToolApprovals(runID, step.ID)
			if err == nil && len(approvedApprovals) > 0 {
				for _, aa := range approvedApprovals {
					if aa.ToolCallID == nil {
						continue
					}
					toolCallID := *aa.ToolCallID

					// Reconstruct the tool call request from approval's request_notes
					tcInput := json.RawMessage(aa.RequestNotes)

					// Determine tool name from approval_type (format: "tool:<name>")
					toolName := strings.TrimPrefix(aa.ApprovalType, "tool:")

					toolReq := tool.ToolCallRequest{
						ToolName: toolName,
						Input:    tcInput,
					}
					tr := tool.ExecuteToolCall(toolReq, o.toolOpts)

					status := "completed"
					if !tr.Success {
						status = "failed"
					}
					outputJSON := []byte(tool.FormatToolOutput(tr))

					// Update tool call status and output
					_, _ = o.repo.UpdateToolCallOutput(toolCallID, outputJSON, status)

					// Create tool message with the result
					toolResult := tool.FormatToolOutput(tr)
					toolMsg := &model.AgentMessage{
						RunID:    runID,
						StepID:   step.ID,
						Role:     "tool",
						Content:  toolResult,
						Metadata: []byte("{}"),
					}
					_ = o.repo.CreateMessage(toolMsg)

					// Append to LLM conversation context so LLM can see the result
					messages = append(messages, llm.Message{Role: "tool", Content: toolResult})
				}
				// After executing approved tools, continue loop to let LLM process the results
				// (don't call LLM here — the regular flow below will call LLM)
			}

			// Convert tool definitions for LLM
			toolDefs := make([]llm.ToolDefinition, 0, len(tool.ListTools()))
			for _, t := range tool.ListTools() {
				toolDefs = append(toolDefs, llm.ToolDefinition{
					Name:        t.Name,
					Description: t.Description,
					InputSchema: t.InputSchema,
				})
			}

			// Apply step-level timeout
			stepCtx := ctx
			var stepCancel context.CancelFunc
			if o.toolOpts.StepTimeout > 0 {
				stepCtx, stepCancel = context.WithTimeout(ctx, time.Duration(o.toolOpts.StepTimeout)*time.Second)
			}

			resp, err := o.llm.ChatCompletionWithTools(stepCtx, messages, toolDefs, o.toolOpts.ToolChoice)

			if stepCancel != nil {
				stepCancel()
			}

			if err != nil {
				if ctx.Err() != nil {
					_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
					_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
					return
				}
				_, _ = o.repo.MarkStepFailed(step.ID, fmt.Sprintf("Run failed: LLM error: %s", err.Error()))
				_, _ = o.repo.UpdateRunStatus(runID, "failed")
				return
			}

			// Store assistant message in DB
			asstContent := resp.Content
			asstMsg := &model.AgentMessage{
				RunID:     runID,
				StepID:    step.ID,
				ProfileID: step.ProfileID,
				Role:      "assistant",
				Content:   asstContent,
				Metadata:  []byte("{}"),
			}
			_ = o.repo.CreateMessage(asstMsg)

			// Add to LLM conversation context
			messages = append(messages, llm.Message{Role: "assistant", Content: asstContent})

			// Determine tool calls: native or fallback to text parsing
			toolCalls := resp.ToolCalls
			if len(toolCalls) == 0 {
				// Fall back to text parsing for backward compat (e.g., FakeProvider)
				parsedCalls, _ := tool.ParseToolCalls(resp.Content)
				for _, pc := range parsedCalls {
					toolCalls = append(toolCalls, llm.ToolCall{
						ToolName: pc.ToolName,
						Input:    pc.Input,
					})
				}
			}

			if len(toolCalls) == 0 {
				finalResponse = asstContent
				break
			}

			// Execute tool calls and record results
			for _, tc := range toolCalls {
				if o.toolOpts.RequiresApproval(tc.ToolName) {
					// --- APPROVAL GATE: PAUSE RUN ---
					toolCall := &model.AgentToolCall{
						RunID:    runID,
						StepID:   step.ID,
						ToolName: tc.ToolName,
						Input:    tc.Input,
						Status:   "recorded",
					}
					_ = o.repo.CreateToolCall(toolCall)

					approval := &model.HumanApproval{
						RunID:         runID,
						StepID:        step.ID,
						ApprovalType:  "tool:" + tc.ToolName,
						Status:        "pending",
						RequestedBy:   "system",
						RequestNotes:  string(tc.Input),
						ToolCallID:    &toolCall.ID,
					}
					_ = o.repo.CreateApproval(approval)

					// Mark step as waiting_approval and run as paused, then exit
					_, _ = o.repo.MarkStepWaitingApproval(step.ID)
					_, _ = o.repo.UpdateRunStatus(runID, "paused")
					_, _ = o.repo.UpdateRunSummary(runID, "Run paused: waiting for human approval on tool: "+tc.ToolName)

					// Exit the goroutine - the defer will clean up the running map
					return
				} else {
					// --- NORMAL EXECUTION ---
					toolReq := tool.ToolCallRequest{
						ToolName: tc.ToolName,
						Input:    tc.Input,
					}
					tr := tool.ExecuteToolCall(toolReq, o.toolOpts)
					status := "completed"
					if !tr.Success {
						status = "failed"
					}
					outputJSON := []byte(tool.FormatToolOutput(tr))
					toolCall := &model.AgentToolCall{
						RunID:    runID,
						StepID:   step.ID,
						ToolName: tc.ToolName,
						Input:    tc.Input,
						Output:   outputJSON,
						Status:   status,
					}
					_ = o.repo.CreateToolCall(toolCall)

					toolResult := tool.FormatToolOutput(tr)
					toolMsg := &model.AgentMessage{
						RunID:    runID,
						StepID:   step.ID,
						Role:     "tool",
						Content:  toolResult,
						Metadata: []byte("{}"),
					}
					_ = o.repo.CreateMessage(toolMsg)

					// Add tool result to LLM context for next iteration
					messages = append(messages, llm.Message{Role: "tool", Content: toolResult})
				}
			}

		}

		if finalResponse == "" {
			finalResponse = "Step completed."
		}

		select {
		case <-ctx.Done():
			_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
			_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
			return
		default:
		}

		_, err = o.repo.MarkStepCompleted(step.ID, finalResponse)
		if err != nil {
			if ctx.Err() != nil {
				_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
				_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
				return
			}
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		stepOutputs = append(stepOutputs, finalResponse)
	}

	select {
	case <-ctx.Done():
		_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
		_, _ = o.repo.UpdateRunSummary(runID, "Run cancelled.")
		return
	default:
	}

	_, _ = o.repo.UpdateRunStatus(runID, "completed")

	summary := fmt.Sprintf("Run completed successfully. Executed %d steps.", len(pending))
	_, _ = o.repo.UpdateRunSummary(runID, summary)
}

func (o *Orchestrator) ResumeRun(ctx context.Context, runID string) error {
	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	if run.Status != "paused" {
		return fmt.Errorf("run is not paused: current status is %s", run.Status)
	}

	// Atomic transition back to running
	_, err = o.repo.UpdateRunStatusIfIn(runID, "running", []string{"paused"})
	if err != nil {
		return fmt.Errorf("failed to resume run: %w", err)
	}

	// Register in running map
	var execCtx context.Context
	var cancel context.CancelFunc
	execCtx, cancel = context.WithTimeout(context.Background(), time.Duration(o.toolOpts.RunTimeout)*time.Second)
	o.mu.Lock()
	if _, exists := o.running[runID]; exists {
		o.mu.Unlock()
		cancel()
		_, _ = o.repo.UpdateRunStatusIfIn(runID, "paused", []string{"running"})
		return fmt.Errorf("run is already running")
	}
	o.running[runID] = cancel
	o.mu.Unlock()

	// Start execution goroutine
	go o.executeRun(execCtx, runID)

	return nil
}

func (o *Orchestrator) CancelRun(runID string) error {
	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	if run.Status == "completed" || run.Status == "failed" || run.Status == "cancelled" {
		return fmt.Errorf("run is already in a terminal state: %s", run.Status)
	}

	o.mu.Lock()
	if cancel, exists := o.running[runID]; exists {
		cancel()
		delete(o.running, runID)
	}
	o.mu.Unlock()

	_, err = o.repo.UpdateRunStatus(runID, "cancelled")
	if err != nil {
		return fmt.Errorf("failed to update run status to cancelled: %w", err)
	}

	return nil
}
