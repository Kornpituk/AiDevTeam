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
  type Task,
  type TaskStatus,
  type AgentProfile,
  type AgentTeam,
  type AgentTeamMember,
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
});
