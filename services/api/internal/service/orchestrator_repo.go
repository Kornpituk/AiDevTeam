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
}

func NewOrchestratorRepoImpl(
	runRepo *repository.AgentRunRepository,
	stepRepo *repository.AgentRunStepRepository,
	teamMemberRepo *repository.AgentTeamMemberRepository,
	profileRepo *repository.AgentProfileRepository,
	messageRepo *repository.AgentMessageRepository,
	taskRepo *repository.TaskRepository,
) *OrchestratorRepoImpl {
	return &OrchestratorRepoImpl{
		runRepo:        runRepo,
		stepRepo:       stepRepo,
		teamMemberRepo: teamMemberRepo,
		profileRepo:    profileRepo,
		messageRepo:    messageRepo,
		taskRepo:       taskRepo,
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
