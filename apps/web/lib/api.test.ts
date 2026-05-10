import { setupServer } from "msw/node";
import { http, HttpResponse } from "msw";
import { API_BASE_URL } from "@/lib/env";
import { getTasks, getTask, createTask, updateTaskStatus, getEvents, getArtifacts, type Task, type TaskStatus } from "@/lib/api";

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
});
