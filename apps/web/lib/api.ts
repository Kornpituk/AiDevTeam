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
