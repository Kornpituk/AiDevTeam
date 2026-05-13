package service

import (
	"github.com/Kornpituk/AiDevTeam/services/api/internal/model"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/repository"
)

type OrchestratorRepoImpl struct {
	runRepo        *repository.AgentRunRepository
	stepRepo       *repository.AgentRunStepRepository
	teamMemberRepo *repository.AgentTeamMemberRepository
	profileRepo    *repository.AgentProfileRepository
	messageRepo    *repository.AgentMessageRepository
	taskRepo       *repository.TaskRepository
	toolCallRepo   *repository.AgentToolCallRepository
	approvalRepo   *repository.HumanApprovalRepository
}

func NewOrchestratorRepoImpl(
	runRepo *repository.AgentRunRepository,
	stepRepo *repository.AgentRunStepRepository,
	teamMemberRepo *repository.AgentTeamMemberRepository,
	profileRepo *repository.AgentProfileRepository,
	messageRepo *repository.AgentMessageRepository,
	taskRepo *repository.TaskRepository,
	toolCallRepo *repository.AgentToolCallRepository,
	approvalRepo *repository.HumanApprovalRepository,
) *OrchestratorRepoImpl {
	return &OrchestratorRepoImpl{
		runRepo:        runRepo,
		stepRepo:       stepRepo,
		teamMemberRepo: teamMemberRepo,
		profileRepo:    profileRepo,
		messageRepo:    messageRepo,
		taskRepo:       taskRepo,
		toolCallRepo:   toolCallRepo,
		approvalRepo:   approvalRepo,
	}
}

func (r *OrchestratorRepoImpl) GetRunByID(id string) (*model.AgentRun, error) {
	return r.runRepo.GetByID(id)
}

func (r *OrchestratorRepoImpl) UpdateRunStatus(id string, status string) (*model.AgentRun, error) {
	return r.runRepo.UpdateStatus(id, status)
}

func (r *OrchestratorRepoImpl) UpdateRunStatusIfIn(id string, newStatus string, allowedStatuses []string) (*model.AgentRun, error) {
	return r.runRepo.UpdateRunStatusIfIn(id, newStatus, allowedStatuses)
}

func (r *OrchestratorRepoImpl) GetTeamMembers(teamID string) ([]model.AgentTeamMember, error) {
	return r.teamMemberRepo.GetByTeamID(teamID)
}

func (r *OrchestratorRepoImpl) GetProfileByID(id string) (*model.AgentProfile, error) {
	return r.profileRepo.GetByID(id)
}

func (r *OrchestratorRepoImpl) CreateStep(step *model.AgentRunStep) error {
	return r.stepRepo.Create(step)
}

func (r *OrchestratorRepoImpl) GetStepsByRunID(runID string) ([]model.AgentRunStep, error) {
	return r.stepRepo.GetByRunID(runID)
}

func (r *OrchestratorRepoImpl) MarkStepStarted(id string) (*model.AgentRunStep, error) {
	return r.stepRepo.MarkStarted(id)
}

func (r *OrchestratorRepoImpl) MarkStepCompleted(id string, output string) (*model.AgentRunStep, error) {
	return r.stepRepo.MarkCompleted(id, output)
}

func (r *OrchestratorRepoImpl) MarkStepFailed(id string, output string) (*model.AgentRunStep, error) {
	return r.stepRepo.MarkFailed(id, output)
}

func (r *OrchestratorRepoImpl) CreateMessage(message *model.AgentMessage) error {
	return r.messageRepo.Create(message)
}

func (r *OrchestratorRepoImpl) UpdateRunSummary(id string, summary string) (*model.AgentRun, error) {
	return r.runRepo.UpdateSummary(id, summary)
}

func (r *OrchestratorRepoImpl) GetTaskByID(id string) (*model.Task, error) {
	return r.taskRepo.GetByID(id)
}

func (r *OrchestratorRepoImpl) CreateToolCall(toolCall *model.AgentToolCall) error {
	return r.toolCallRepo.Create(toolCall)
}

func (r *OrchestratorRepoImpl) UpdateToolCallStatus(id string, status string) (*model.AgentToolCall, error) {
	return r.toolCallRepo.UpdateStatus(id, status)
}

func (r *OrchestratorRepoImpl) CreateApproval(approval *model.HumanApproval) error {
	return r.approvalRepo.Create(approval)
}

func (r *OrchestratorRepoImpl) GetApprovedToolApprovals(runID, stepID string) ([]model.HumanApproval, error) {
	return r.approvalRepo.GetApprovedToolApprovals(runID, stepID)
}

func (r *OrchestratorRepoImpl) UpdateToolCallOutput(id string, output []byte, status string) (*model.AgentToolCall, error) {
	return r.toolCallRepo.UpdateOutput(id, output, status)
}

func (r *OrchestratorRepoImpl) MarkStepWaitingApproval(id string) (*model.AgentRunStep, error) {
	return r.stepRepo.MarkWaitingApproval(id)
}

func (r *OrchestratorRepoImpl) GetMessagesByStepID(stepID string) ([]model.AgentMessage, error) {
	return r.messageRepo.GetByStepID(stepID)
}
