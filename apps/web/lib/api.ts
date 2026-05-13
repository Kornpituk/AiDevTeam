import { API_BASE_URL } from "./env";

export interface AgentProfile {
  id: string;
  name: string;
  role: string;
  description?: string;
  system_prompt?: string;
  default_model?: string;
  created_at: string;
  updated_at: string;
}

export interface AgentTeam {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface AgentTeamMember {
  id: string;
  team_id: string;
  profile_id: string;
  member_role: string;
  position: number;
  created_at: string;
}

export type TaskStatus =
  | "pending"
  | "planning"
  | "approved"
  | "in_progress"
  | "reviewing"
  | "completed"
  | "failed";

export const TASK_STATUSES: TaskStatus[] = [
  "pending",
  "planning",
  "approved",
  "in_progress",
  "reviewing",
  "completed",
  "failed",
];

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  plan?: string;
  review_notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Event {
  id: string;
  task_id: string;
  event_type: string;
  message?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface Artifact {
  id: string;
  task_id: string;
  name: string;
  artifact_type: string;
  content?: string;
  content_type?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    cache: "no-store",
  });

  const json = (await response.json()) as ApiResponse<T>;

  if (!response.ok) {
    throw new Error(json.error || `HTTP ${response.status}`);
  }

  if (json.error) {
    throw new Error(json.error);
  }

  return json.data as T;
}

export async function getTasks(): Promise<Task[]> {
  return fetchApi<Task[]>("/tasks");
}

export async function getTask(id: string): Promise<Task> {
  return fetchApi<Task>(`/tasks/${id}`);
}

export async function createTask(data: { title: string; description?: string }): Promise<Task> {
  return fetchApi<Task>("/tasks", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateTaskStatus(id: string, status: TaskStatus): Promise<Task> {
  return fetchApi<Task>(`/tasks/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export async function updateTaskPlan(id: string, plan: string): Promise<Task> {
  return fetchApi<Task>(`/tasks/${id}/plan`, {
    method: "PATCH",
    body: JSON.stringify({ plan }),
  });
}

export async function updateTaskReviewNotes(id: string, reviewNotes: string): Promise<Task> {
  return fetchApi<Task>(`/tasks/${id}/review-notes`, {
    method: "PATCH",
    body: JSON.stringify({ review_notes: reviewNotes }),
  });
}

export async function getEvents(taskId: string): Promise<Event[]> {
  return fetchApi<Event[]>(`/tasks/${taskId}/events`);
}

export async function createEvent(
  taskId: string,
  data: { event_type: string; message?: string; metadata?: Record<string, unknown> }
): Promise<Event> {
  return fetchApi<Event>(`/tasks/${taskId}/events`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getArtifacts(taskId: string): Promise<Artifact[]> {
  return fetchApi<Artifact[]>(`/tasks/${taskId}/artifacts`);
}

export async function createArtifact(
  taskId: string,
  data: {
    name: string;
    artifact_type: string;
    content?: string;
    content_type?: string;
    metadata?: Record<string, unknown>;
  }
): Promise<Artifact> {
  return fetchApi<Artifact>(`/tasks/${taskId}/artifacts`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getAgentProfiles(): Promise<AgentProfile[]> {
  return fetchApi<AgentProfile[]>("/agent-profiles");
}

export async function getAgentProfile(id: string): Promise<AgentProfile> {
  return fetchApi<AgentProfile>(`/agent-profiles/${id}`);
}

export async function createAgentProfile(data: {
  name: string;
  role: string;
  description?: string;
  system_prompt?: string;
  default_model?: string;
}): Promise<AgentProfile> {
  return fetchApi<AgentProfile>("/agent-profiles", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getAgentTeams(): Promise<AgentTeam[]> {
  return fetchApi<AgentTeam[]>("/agent-teams");
}

export async function getAgentTeam(id: string): Promise<AgentTeam> {
  return fetchApi<AgentTeam>(`/agent-teams/${id}`);
}

export async function createAgentTeam(data: {
  name: string;
  description?: string;
}): Promise<AgentTeam> {
  return fetchApi<AgentTeam>("/agent-teams", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getAgentTeamMembers(teamId: string): Promise<AgentTeamMember[]> {
  return fetchApi<AgentTeamMember[]>(`/agent-teams/${teamId}/members`);
}

export async function addAgentTeamMember(
  teamId: string,
  data: {
    profile_id: string;
    member_role: string;
    position?: number;
  }
): Promise<AgentTeamMember> {
  return fetchApi<AgentTeamMember>(`/agent-teams/${teamId}/members`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export type AgentRunStatus =
  | "draft"
  | "planned"
  | "waiting_approval"
  | "approved"
  | "running"
  | "paused"
  | "completed"
  | "failed"
  | "cancelled";

export const AGENT_RUN_STATUSES: AgentRunStatus[] = [
  "draft",
  "planned",
  "waiting_approval",
  "approved",
  "running",
  "paused",
  "completed",
  "failed",
  "cancelled",
];

export type AgentRunStepStatus =
  | "pending"
  | "waiting_approval"
  | "running"
  | "completed"
  | "failed"
  | "skipped"
  | "cancelled";

export const AGENT_RUN_STEP_STATUSES: AgentRunStepStatus[] = [
  "pending",
  "waiting_approval",
  "running",
  "completed",
  "failed",
  "skipped",
  "cancelled",
];

export interface AgentRun {
  id: string;
  task_id: string;
  team_id?: string;
  status: AgentRunStatus;
  goal?: string;
  summary?: string;
  created_at: string;
  updated_at: string;
}

export interface AgentRunStep {
  id: string;
  run_id: string;
  profile_id?: string;
  step_type: string;
  status: AgentRunStepStatus;
  title: string;
  instructions?: string;
  output?: string;
  position: number;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export async function getAgentRuns(taskId: string): Promise<AgentRun[]> {
  return fetchApi<AgentRun[]>(`/tasks/${taskId}/agent-runs`);
}

export async function getAgentRun(id: string): Promise<AgentRun> {
  return fetchApi<AgentRun>(`/agent-runs/${id}`);
}

export async function startAgentRun(id: string): Promise<AgentRun> {
  return fetchApi<AgentRun>(`/agent-runs/${id}/start`, {
    method: "POST",
  });
}

export async function cancelAgentRun(id: string): Promise<AgentRun> {
  return fetchApi<AgentRun>(`/agent-runs/${id}/cancel`, {
    method: "POST",
  });
}

export async function createAgentRun(
  taskId: string,
  data: {
    team_id?: string;
    goal?: string;
    status?: AgentRunStatus;
  }
): Promise<AgentRun> {
  return fetchApi<AgentRun>(`/tasks/${taskId}/agent-runs`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getAgentRunSteps(runId: string): Promise<AgentRunStep[]> {
  return fetchApi<AgentRunStep[]>(`/agent-runs/${runId}/steps`);
}

export async function createAgentRunStep(
  runId: string,
  data: {
    profile_id?: string;
    step_type: string;
    title: string;
    instructions?: string;
    position?: number;
    status?: AgentRunStepStatus;
  }
): Promise<AgentRunStep> {
  return fetchApi<AgentRunStep>(`/agent-runs/${runId}/steps`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateAgentRunStepStatus(
  stepId: string,
  status: AgentRunStepStatus
): Promise<AgentRunStep> {
  return fetchApi<AgentRunStep>(`/agent-run-steps/${stepId}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export type AgentMessageRole =
  | "system"
  | "user"
  | "assistant"
  | "tool"
  | "reviewer";

export const AGENT_MESSAGE_ROLES: AgentMessageRole[] = [
  "system",
  "user",
  "assistant",
  "tool",
  "reviewer",
];

export interface AgentMessage {
  id: string;
  run_id: string;
  step_id?: string;
  profile_id?: string;
  role: AgentMessageRole;
  content: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export type HumanApprovalStatus =
  | "pending"
  | "approved"
  | "rejected"
  | "cancelled";

export const HUMAN_APPROVAL_STATUSES: HumanApprovalStatus[] = [
  "pending",
  "approved",
  "rejected",
  "cancelled",
];

export interface HumanApproval {
  id: string;
  task_id?: string;
  run_id?: string;
  step_id?: string;
  approval_type: string;
  status: HumanApprovalStatus;
  requested_by?: string;
  decided_by?: string;
  request_notes?: string;
  decision_notes?: string;
  tool_call_id?: string;
  created_at: string;
  decided_at?: string;
}

export type AgentToolCallStatus =
  | "recorded"
  | "approved"
  | "rejected"
  | "completed"
  | "failed";

export const AGENT_TOOL_CALL_STATUSES: AgentToolCallStatus[] = [
  "recorded",
  "approved",
  "rejected",
  "completed",
  "failed",
];

export interface AgentToolCall {
  id: string;
  run_id: string;
  step_id?: string;
  tool_name: string;
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  status: AgentToolCallStatus;
  created_at: string;
  completed_at?: string;
}

export async function getAgentMessages(runId: string): Promise<AgentMessage[]> {
  return fetchApi<AgentMessage[]>(`/agent-runs/${runId}/messages`);
}

export async function createAgentMessage(
  runId: string,
  data: {
    step_id?: string;
    profile_id?: string;
    role: AgentMessageRole;
    content: string;
    metadata?: Record<string, unknown>;
  }
): Promise<AgentMessage> {
  return fetchApi<AgentMessage>(`/agent-runs/${runId}/messages`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getHumanApprovals(runId: string): Promise<HumanApproval[]> {
  return fetchApi<HumanApproval[]>(`/agent-runs/${runId}/approvals`);
}

export async function createHumanApproval(
  runId: string,
  data: {
    task_id?: string;
    step_id?: string;
    approval_type: string;
    status?: HumanApprovalStatus;
    request_notes?: string;
  }
): Promise<HumanApproval> {
  return fetchApi<HumanApproval>(`/agent-runs/${runId}/approvals`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateHumanApprovalStatus(
  approvalId: string,
  status: HumanApprovalStatus
): Promise<HumanApproval> {
  return fetchApi<HumanApproval>(`/human-approvals/${approvalId}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export async function getAgentToolCalls(runId: string): Promise<AgentToolCall[]> {
  return fetchApi<AgentToolCall[]>(`/agent-runs/${runId}/tool-calls`);
}

export async function createAgentToolCall(
  runId: string,
  data: {
    step_id?: string;
    tool_name: string;
    input?: Record<string, unknown>;
    output?: Record<string, unknown>;
    status?: AgentToolCallStatus;
  }
): Promise<AgentToolCall> {
  return fetchApi<AgentToolCall>(`/agent-runs/${runId}/tool-calls`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateAgentToolCallStatus(
  toolCallId: string,
  status: AgentToolCallStatus
): Promise<AgentToolCall> {
  return fetchApi<AgentToolCall>(`/agent-tool-calls/${toolCallId}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}
