package service

import (
	"context"
	"fmt"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
)

type OrchestratorRepository interface {
	GetRunByID(id string) (*model.AgentRun, error)
	UpdateRunStatus(id string, status string) (*model.AgentRun, error)
	GetTeamMembers(teamID string) ([]model.AgentTeamMember, error)
	GetProfileByID(id string) (*model.AgentProfile, error)
	CreateStep(step *model.AgentRunStep) error
	GetStepsByRunID(runID string) ([]model.AgentRunStep, error)
	MarkStepStarted(id string) (*model.AgentRunStep, error)
	MarkStepCompleted(id string, output string) (*model.AgentRunStep, error)
	MarkStepFailed(id string, output string) (*model.AgentRunStep, error)
	CreateMessage(message *model.AgentMessage) error
	UpdateRunSummary(id string, summary string) (*model.AgentRun, error)
}

type Orchestrator struct {
	repo   OrchestratorRepository
	llm    llm.LLMProvider
	cancel context.CancelFunc
}

func NewOrchestrator(repo OrchestratorRepository, llmProvider llm.LLMProvider) *Orchestrator {
	return &Orchestrator{
		repo: repo,
		llm:  llmProvider,
	}
}

func (o *Orchestrator) StartRun(ctx context.Context, runID string) error {
	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	validStartStatuses := map[string]bool{
		"draft":           true,
		"planned":         true,
		"waiting_approval": true,
		"approved":        true,
	}
	if !validStartStatuses[run.Status] {
		return fmt.Errorf("run is not in startable state: %s", run.Status)
	}

	_, err = o.repo.UpdateRunStatus(runID, "running")
	if err != nil {
		return fmt.Errorf("failed to update run status to running: %w", err)
	}

	if run.TeamID != "" {
		err = o.createStepsFromTeam(runID, run.TeamID, run.Goal)
		if err != nil {
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return fmt.Errorf("failed to create steps from team: %w", err)
		}
	}

	go o.executeRun(ctx, runID)

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

		stepType := member.MemberRole
		if stepType == "" {
			stepType = "agent"
		}

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
	steps, err := o.repo.GetStepsByRunID(runID)
	if err != nil {
		_, _ = o.repo.UpdateRunStatus(runID, "failed")
		return
	}

	if len(steps) == 0 {
		_, _ = o.repo.UpdateRunStatus(runID, "completed")
		return
	}

	var stepOutputs []string

	for _, step := range steps {
		select {
		case <-ctx.Done():
			_, _ = o.repo.UpdateRunStatus(runID, "cancelled")
			return
		default:
		}

		_, err := o.repo.MarkStepStarted(step.ID)
		if err != nil {
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		profile, err := o.repo.GetProfileByID(step.ProfileID)
		if err != nil {
			_, _ = o.repo.MarkStepFailed(step.ID, fmt.Sprintf("Failed to get profile: %v", err))
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		var previousOutputs string
		for i, out := range stepOutputs {
			previousOutputs += fmt.Sprintf("Step %d Output:\n%s\n\n", i+1, out)
		}

		messages := []llm.Message{
			{
				Role:    "system",
				Content: profile.SystemPrompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Instructions: %s\n\nPrevious Step Outputs:\n%s", step.Instructions, previousOutputs),
			},
		}

		response, err := o.llm.ChatCompletion(ctx, messages)
		if err != nil {
			_, _ = o.repo.MarkStepFailed(step.ID, fmt.Sprintf("LLM call failed: %v", err))
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		message := &model.AgentMessage{
			RunID:     runID,
			StepID:    step.ID,
			ProfileID: step.ProfileID,
			Role:      "assistant",
			Content:   response,
			Metadata:  []byte("{}"),
		}
		_ = o.repo.CreateMessage(message)

		_, err = o.repo.MarkStepCompleted(step.ID, response)
		if err != nil {
			_, _ = o.repo.UpdateRunStatus(runID, "failed")
			return
		}

		stepOutputs = append(stepOutputs, response)
	}

	_, _ = o.repo.UpdateRunStatus(runID, "completed")

	summary := fmt.Sprintf("Run completed successfully. Executed %d steps.", len(steps))
	_, _ = o.repo.UpdateRunSummary(runID, summary)
}

func (o *Orchestrator) CancelRun(runID string) error {
	run, err := o.repo.GetRunByID(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	if run.Status == "completed" || run.Status == "failed" || run.Status == "cancelled" {
		return fmt.Errorf("run is already in a terminal state: %s", run.Status)
	}

	if o.cancel != nil {
		o.cancel()
	}

	_, err = o.repo.UpdateRunStatus(runID, "cancelled")
	if err != nil {
		return fmt.Errorf("failed to update run status to cancelled: %w", err)
	}

	return nil
}
