import { http, HttpResponse } from "msw";
import { API_BASE_URL } from "@/lib/env";
import type {
  AgentRun,
  AgentRunStatus,
  HumanApproval,
  HumanApprovalStatus,
} from "@/lib/api";

const mockRuns: AgentRun[] = [
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
    status: "paused",
    goal: "Review the code",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
];

const mockApprovals: HumanApproval[] = [
  {
    id: "a1",
    run_id: "r2",
    approval_type: "step_execution",
    status: "pending",
    request_notes: '{"key": "value"}',
    created_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "a2",
    run_id: "r2",
    approval_type: "run_completion",
    status: "approved",
    decided_by: "user123",
    decision_notes: "Looks good.",
    created_at: "2024-01-01T00:00:00Z",
    decided_at: "2024-01-02T00:00:00Z",
  },
];

const mockNotFoundError = { error: "Run not found" };

export const handlers = [
  // Resume agent run
  http.post(`${API_BASE_URL}/agent-runs/:id/resume`, ({ params }) => {
    const run = mockRuns.find((r) => r.id === params.id);
    if (run) {
      return HttpResponse.json({
        data: { ...run, status: "running" as AgentRunStatus },
      });
    }
    return HttpResponse.json(mockNotFoundError, { status: 404 });
  }),

  // Start agent run
  http.post(`${API_BASE_URL}/agent-runs/:id/start`, ({ params }) => {
    const run = mockRuns.find((r) => r.id === params.id);
    if (run) {
      return HttpResponse.json({
        data: { ...run, status: "running" as AgentRunStatus },
      });
    }
    return HttpResponse.json(mockNotFoundError, { status: 404 });
  }),

  // Cancel agent run
  http.post(`${API_BASE_URL}/agent-runs/:id/cancel`, ({ params }) => {
    const run = mockRuns.find((r) => r.id === params.id);
    if (run) {
      return HttpResponse.json({
        data: { ...run, status: "cancelled" as AgentRunStatus },
      });
    }
    return HttpResponse.json(mockNotFoundError, { status: 404 });
  }),

  // Update human approval status
  http.patch(
    `${API_BASE_URL}/human-approvals/:id/status`,
    async ({ params, request }) => {
      const body = (await request.json()) as { status: HumanApprovalStatus };
      const approval = mockApprovals.find((a) => a.id === params.id);
      if (approval) {
        return HttpResponse.json({
          data: { ...approval, status: body.status },
        });
      }
      return HttpResponse.json({ error: "Not found" }, { status: 404 });
    }
  ),
];
