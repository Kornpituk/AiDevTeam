import { setupServer } from "msw/node";
import { http, HttpResponse } from "msw";
import { API_BASE_URL } from "@/lib/env";
import {
  getTasks,
  getTask,
  createTask,
  updateTaskStatus,
  getEvents,
  getArtifacts,
  getAgentProfiles,
  getAgentProfile,
  createAgentProfile,
  getAgentTeams,
  getAgentTeam,
  createAgentTeam,
  getAgentTeamMembers,
  addAgentTeamMember,
  getAgentRuns,
  getAgentRun,
  createAgentRun,
  getAgentRunSteps,
  createAgentRunStep,
  updateAgentRunStepStatus,
  type Task,
  type TaskStatus,
  type AgentProfile,
  type AgentTeam,
  type AgentTeamMember,
  type AgentRun,
  type AgentRunStep,
  type AgentRunStatus,
  type AgentRunStepStatus,
} from "@/lib/api";

const mockAgentProfiles: AgentProfile[] = [
  {
    id: "p1",
    name: "Senior Developer",
    role: "developer",
    description: "Expert in Go and backend development",
    system_prompt: "You are a senior developer...",
    default_model: "gpt-4",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "p2",
    name: "Code Reviewer",
    role: "reviewer",
    created_at: "2024-01-02T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z",
  },
];

const mockAgentTeams: AgentTeam[] = [
  {
    id: "t1",
    name: "Development Team",
    description: "A team for development tasks",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
];

const mockTeamMembers: AgentTeamMember[] = [
  {
    id: "m1",
    team_id: "t1",
    profile_id: "p1",
    member_role: "lead_developer",
    position: 0,
    created_at: "2024-01-01T00:00:00Z",
  },
];

const mockAgentRuns: AgentRun[] = [
  {
    id: "r1",
    task_id: "1",
    team_id: "t1",
    status: "draft",
    goal: "Implement the feature",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "r2",
    task_id: "1",
    status: "running",
    goal: "Review the code",
    created_at: "2024-01-02T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z",
  },
];

const mockAgentRunSteps: AgentRunStep[] = [
  {
    id: "s1",
    run_id: "r1",
    profile_id: "p1",
    step_type: "plan",
    status: "completed",
    title: "Plan the implementation",
    instructions: "Create a detailed plan",
    output: "Plan created",
    position: 0,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "s2",
    run_id: "r1",
    profile_id: "p2",
    step_type: "review",
    status: "pending",
    title: "Review the code",
    instructions: "Review the implementation",
    position: 1,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
];

const mockTasks: Task[] = [
  {
    id: "1",
    title: "Test Task 1",
    description: "Description 1",
    status: "pending",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "2",
    title: "Test Task 2",
    status: "in_progress",
    created_at: "2024-01-02T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z",
  },
];

const handlers = [
  http.get(`${API_BASE_URL}/tasks`, () => {
    return HttpResponse.json({ data: mockTasks });
  }),

  http.get(`${API_BASE_URL}/tasks/:id`, ({ params }) => {
    const task = mockTasks.find((t) => t.id === params.id);
    if (task) {
      return HttpResponse.json({ data: task });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),

  http.post(`${API_BASE_URL}/tasks`, async ({ request }) => {
    const body = (await request.json()) as { title: string; description?: string };
    const newTask: Task = {
      id: "3",
      title: body.title,
      description: body.description,
      status: "pending",
      created_at: "2024-01-03T00:00:00Z",
      updated_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newTask }, { status: 201 });
  }),

  http.patch(`${API_BASE_URL}/tasks/:id/status`, async ({ params, request }) => {
    const body = (await request.json()) as { status: TaskStatus };
    const task = mockTasks.find((t) => t.id === params.id);
    if (task) {
      return HttpResponse.json({ data: { ...task, status: body.status } });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),

  http.get(`${API_BASE_URL}/tasks/:id/events`, ({ params }) => {
    return HttpResponse.json({
      data: [
        {
          id: "e1",
          task_id: params.id as string,
          event_type: "note",
          message: "Test event",
          created_at: "2024-01-01T00:00:00Z",
        },
      ],
    });
  }),

  http.get(`${API_BASE_URL}/tasks/:id/artifacts`, ({ params }) => {
    return HttpResponse.json({
      data: [
        {
          id: "a1",
          task_id: params.id as string,
          name: "test.go",
          artifact_type: "code",
          content: "package main",
          created_at: "2024-01-01T00:00:00Z",
        },
      ],
    });
  }),

  http.get(`${API_BASE_URL}/agent-profiles`, () => {
    return HttpResponse.json({ data: mockAgentProfiles });
  }),

  http.get(`${API_BASE_URL}/agent-profiles/:id`, ({ params }) => {
    const profile = mockAgentProfiles.find((p) => p.id === params.id);
    if (profile) {
      return HttpResponse.json({ data: profile });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),

  http.post(`${API_BASE_URL}/agent-profiles`, async ({ request }) => {
    const body = (await request.json()) as {
      name: string;
      role: string;
      description?: string;
      system_prompt?: string;
      default_model?: string;
    };
    const newProfile: AgentProfile = {
      id: "p3",
      name: body.name,
      role: body.role,
      description: body.description,
      system_prompt: body.system_prompt,
      default_model: body.default_model,
      created_at: "2024-01-03T00:00:00Z",
      updated_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newProfile }, { status: 201 });
  }),

  http.get(`${API_BASE_URL}/agent-teams`, () => {
    return HttpResponse.json({ data: mockAgentTeams });
  }),

  http.get(`${API_BASE_URL}/agent-teams/:id`, ({ params }) => {
    const team = mockAgentTeams.find((t) => t.id === params.id);
    if (team) {
      return HttpResponse.json({ data: team });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),

  http.post(`${API_BASE_URL}/agent-teams`, async ({ request }) => {
    const body = (await request.json()) as { name: string; description?: string };
    const newTeam: AgentTeam = {
      id: "t2",
      name: body.name,
      description: body.description,
      created_at: "2024-01-03T00:00:00Z",
      updated_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newTeam }, { status: 201 });
  }),

  http.get(`${API_BASE_URL}/agent-teams/:id/members`, ({ params }) => {
    const members = mockTeamMembers.filter((m) => m.team_id === params.id);
    return HttpResponse.json({ data: members });
  }),

  http.post(`${API_BASE_URL}/agent-teams/:id/members`, async ({ params, request }) => {
    const body = (await request.json()) as {
      profile_id: string;
      member_role: string;
      position?: number;
    };
    const newMember: AgentTeamMember = {
      id: "m2",
      team_id: params.id as string,
      profile_id: body.profile_id,
      member_role: body.member_role,
      position: body.position || 0,
      created_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newMember }, { status: 201 });
  }),

  http.get(`${API_BASE_URL}/tasks/:id/agent-runs`, ({ params }) => {
    const runs = mockAgentRuns.filter((r) => r.task_id === params.id);
    return HttpResponse.json({ data: runs });
  }),

  http.get(`${API_BASE_URL}/agent-runs/:id`, ({ params }) => {
    const run = mockAgentRuns.find((r) => r.id === params.id);
    if (run) {
      return HttpResponse.json({ data: run });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),

  http.post(`${API_BASE_URL}/tasks/:id/agent-runs`, async ({ params, request }) => {
    const body = (await request.json()) as {
      team_id?: string;
      goal?: string;
      status?: AgentRunStatus;
    };
    const newRun: AgentRun = {
      id: "r3",
      task_id: params.id as string,
      team_id: body.team_id,
      status: body.status || "draft",
      goal: body.goal,
      created_at: "2024-01-03T00:00:00Z",
      updated_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newRun }, { status: 201 });
  }),

  http.get(`${API_BASE_URL}/agent-runs/:id/steps`, ({ params }) => {
    const steps = mockAgentRunSteps.filter((s) => s.run_id === params.id);
    return HttpResponse.json({ data: steps });
  }),

  http.post(`${API_BASE_URL}/agent-runs/:id/steps`, async ({ params, request }) => {
    const body = (await request.json()) as {
      profile_id?: string;
      step_type: string;
      title: string;
      instructions?: string;
      position?: number;
      status?: AgentRunStepStatus;
    };
    const newStep: AgentRunStep = {
      id: "s3",
      run_id: params.id as string,
      profile_id: body.profile_id,
      step_type: body.step_type,
      status: body.status || "pending",
      title: body.title,
      instructions: body.instructions,
      position: body.position || 0,
      created_at: "2024-01-03T00:00:00Z",
      updated_at: "2024-01-03T00:00:00Z",
    };
    return HttpResponse.json({ data: newStep }, { status: 201 });
  }),

  http.patch(`${API_BASE_URL}/agent-run-steps/:id/status`, async ({ params, request }) => {
    const body = (await request.json()) as { status: AgentRunStepStatus };
    const step = mockAgentRunSteps.find((s) => s.id === params.id);
    if (step) {
      return HttpResponse.json({ data: { ...step, status: body.status } });
    }
    return HttpResponse.json({ error: "Not found" }, { status: 404 });
  }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe("API Functions", () => {
  describe("getTasks", () => {
    it("fetches all tasks", async () => {
      const tasks = await getTasks();
      expect(tasks).toHaveLength(2);
      expect(tasks[0].title).toBe("Test Task 1");
    });
  });

  describe("getTask", () => {
    it("fetches a single task by id", async () => {
      const task = await getTask("1");
      expect(task.id).toBe("1");
      expect(task.title).toBe("Test Task 1");
    });
  });

  describe("createTask", () => {
    it("creates a new task", async () => {
      const task = await createTask({ title: "New Task", description: "New description" });
      expect(task.id).toBe("3");
      expect(task.title).toBe("New Task");
      expect(task.status).toBe("pending");
    });
  });

  describe("updateTaskStatus", () => {
    it("updates task status", async () => {
      const task = await updateTaskStatus("1", "completed");
      expect(task.status).toBe("completed");
    });
  });

  describe("getEvents", () => {
    it("fetches task events", async () => {
      const events = await getEvents("1");
      expect(events).toHaveLength(1);
      expect(events[0].event_type).toBe("note");
    });
  });

  describe("getArtifacts", () => {
    it("fetches task artifacts", async () => {
      const artifacts = await getArtifacts("1");
      expect(artifacts).toHaveLength(1);
      expect(artifacts[0].name).toBe("test.go");
    });
  });

  describe("getAgentProfiles", () => {
    it("fetches all agent profiles", async () => {
      const profiles = await getAgentProfiles();
      expect(profiles).toHaveLength(2);
      expect(profiles[0].name).toBe("Senior Developer");
    });
  });

  describe("getAgentProfile", () => {
    it("fetches a single agent profile by id", async () => {
      const profile = await getAgentProfile("p1");
      expect(profile.id).toBe("p1");
      expect(profile.name).toBe("Senior Developer");
    });
  });

  describe("createAgentProfile", () => {
    it("creates a new agent profile", async () => {
      const profile = await createAgentProfile({
        name: "New Agent",
        role: "tester",
        description: "A test agent",
      });
      expect(profile.id).toBe("p3");
      expect(profile.name).toBe("New Agent");
      expect(profile.role).toBe("tester");
    });
  });

  describe("getAgentTeams", () => {
    it("fetches all agent teams", async () => {
      const teams = await getAgentTeams();
      expect(teams).toHaveLength(1);
      expect(teams[0].name).toBe("Development Team");
    });
  });

  describe("getAgentTeam", () => {
    it("fetches a single agent team by id", async () => {
      const team = await getAgentTeam("t1");
      expect(team.id).toBe("t1");
      expect(team.name).toBe("Development Team");
    });
  });

  describe("createAgentTeam", () => {
    it("creates a new agent team", async () => {
      const team = await createAgentTeam({
        name: "New Team",
        description: "A new team",
      });
      expect(team.id).toBe("t2");
      expect(team.name).toBe("New Team");
    });
  });

  describe("getAgentTeamMembers", () => {
    it("fetches team members", async () => {
      const members = await getAgentTeamMembers("t1");
      expect(members).toHaveLength(1);
      expect(members[0].profile_id).toBe("p1");
    });
  });

  describe("addAgentTeamMember", () => {
    it("adds a member to a team", async () => {
      const member = await addAgentTeamMember("t1", {
        profile_id: "p2",
        member_role: "reviewer",
        position: 1,
      });
      expect(member.id).toBe("m2");
      expect(member.profile_id).toBe("p2");
      expect(member.member_role).toBe("reviewer");
    });
  });

  describe("getAgentRuns", () => {
    it("fetches agent runs for a task", async () => {
      const runs = await getAgentRuns("1");
      expect(runs).toHaveLength(2);
      expect(runs[0].id).toBe("r1");
      expect(runs[0].task_id).toBe("1");
    });
  });

  describe("getAgentRun", () => {
    it("fetches a single agent run by id", async () => {
      const run = await getAgentRun("r1");
      expect(run.id).toBe("r1");
      expect(run.status).toBe("draft");
      expect(run.goal).toBe("Implement the feature");
    });
  });

  describe("createAgentRun", () => {
    it("creates a new agent run", async () => {
      const run = await createAgentRun("1", {
        team_id: "t1",
        goal: "New run goal",
        status: "draft",
      });
      expect(run.id).toBe("r3");
      expect(run.task_id).toBe("1");
      expect(run.team_id).toBe("t1");
      expect(run.goal).toBe("New run goal");
      expect(run.status).toBe("draft");
    });
  });

  describe("getAgentRunSteps", () => {
    it("fetches steps for an agent run", async () => {
      const steps = await getAgentRunSteps("r1");
      expect(steps).toHaveLength(2);
      expect(steps[0].id).toBe("s1");
      expect(steps[0].run_id).toBe("r1");
    });
  });

  describe("createAgentRunStep", () => {
    it("creates a new agent run step", async () => {
      const step = await createAgentRunStep("r1", {
        profile_id: "p1",
        step_type: "implement",
        title: "Implement the feature",
        instructions: "Write the code",
        position: 2,
        status: "pending",
      });
      expect(step.id).toBe("s3");
      expect(step.run_id).toBe("r1");
      expect(step.step_type).toBe("implement");
      expect(step.title).toBe("Implement the feature");
      expect(step.status).toBe("pending");
    });
  });

  describe("updateAgentRunStepStatus", () => {
    it("updates step status", async () => {
      const step = await updateAgentRunStepStatus("s1", "completed");
      expect(step.status).toBe("completed");
    });
  });
});
