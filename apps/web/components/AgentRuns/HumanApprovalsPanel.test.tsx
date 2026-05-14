import { render, screen, waitFor, act, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { HumanApprovalsPanel } from "./HumanApprovalsPanel";
import type { HumanApproval, AgentToolCall, AgentRunStep } from "@/lib/api";

// ---------------------------------------------------------------------------
// Mocks – use vi.hoisted() so mock fns are created before vi.mock runs
// ---------------------------------------------------------------------------

const mockGetHumanApprovals = vi.hoisted(() => vi.fn());
const mockGetAgentToolCalls = vi.hoisted(() => vi.fn());
const mockUpdateHumanApprovalStatus = vi.hoisted(() => vi.fn());
const mockResumeAgentRun = vi.hoisted(() => vi.fn());
const mockUseWebSocket = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api", () => ({
  getHumanApprovals: mockGetHumanApprovals,
  getAgentToolCalls: mockGetAgentToolCalls,
  updateHumanApprovalStatus: mockUpdateHumanApprovalStatus,
  resumeAgentRun: mockResumeAgentRun,
}));

vi.mock("@/lib/hooks/useWebSocket", () => ({
  useWebSocket: mockUseWebSocket,
}));

vi.mock("@/components/ui/card", () => {
  const React = require("react");
  return {
    Card: ({ children, className }: any) =>
      React.createElement("div", { className, "data-testid": "card" }, children),
    CardHeader: ({ children, className }: any) =>
      React.createElement("div", { className }, children),
    CardTitle: ({ children, className }: any) =>
      React.createElement("div", { className }, children),
    CardContent: ({ children, className }: any) =>
      React.createElement("div", { className }, children),
  };
});

vi.mock("@/components/ui/button", () => {
  const React = require("react");
  return {
    Button: ({ children, onClick, disabled, type }: any) =>
      React.createElement(
        "button",
        { onClick, disabled, type },
        children
      ),
  };
});

vi.mock("./AgentRunBadges", () => {
  const React = require("react");
  return {
    ApprovalStatusBadge: ({ status }: any) =>
      React.createElement("span", { "data-testid": `badge-${status}` }, status),
  };
});

vi.mock("./ApprovalStatusSelector", () => ({
  ApprovalStatusSelector: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "approval-status-selector" });
  },
}));

vi.mock("./AddHumanApprovalForm", () => ({
  AddHumanApprovalForm: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "add-approval-form" });
  },
}));

vi.mock("@/lib/utils", () => ({
  formatDate: () => "Jan 1, 2024",
}));

// ---------------------------------------------------------------------------
// Shared test data
// ---------------------------------------------------------------------------

const mockSteps: AgentRunStep[] = [];

const pendingApproval: HumanApproval = {
  id: "a1",
  run_id: "r1",
  step_id: "s1",
  approval_type: "step_execution",
  status: "pending",
  request_notes: '{"key": "value"}',
  created_at: "2024-01-01T00:00:00Z",
};

const approvedApproval: HumanApproval = {
  id: "a2",
  run_id: "r1",
  approval_type: "run_completion",
  status: "approved",
  decided_by: "user123",
  decision_notes: "Looks good.",
  tool_call_id: "tc1",
  created_at: "2024-01-01T00:00:00Z",
  decided_at: "2024-01-02T00:00:00Z",
};

const rejectedApproval: HumanApproval = {
  id: "a3",
  run_id: "r1",
  approval_type: "step_execution",
  status: "rejected",
  decided_by: "user456",
  created_at: "2024-01-01T00:00:00Z",
  decided_at: "2024-01-02T00:00:00Z",
};

const mockToolCalls: AgentToolCall[] = [
  {
    id: "tc1",
    run_id: "r1",
    step_id: "s1",
    tool_name: "read_file",
    input: { path: "src/main.go" },
    output: { content: "test output" },
    status: "completed",
    created_at: "2024-01-01T00:00:00Z",
    completed_at: "2024-01-01T00:00:01Z",
  },
];

function setupDefaults(
  approvals: HumanApproval[] = [pendingApproval],
  toolCalls: AgentToolCall[] = mockToolCalls
) {
  mockGetHumanApprovals.mockResolvedValue(approvals);
  mockGetAgentToolCalls.mockResolvedValue(toolCalls);
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("HumanApprovalsPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseWebSocket.mockReturnValue({ isConnected: false, reconnect: vi.fn() });
  });

  // ---- Loading state -------------------------------------------------------

  describe("loading state", () => {
    it("shows loading indicator on mount", () => {
      mockGetHumanApprovals.mockReturnValue(new Promise(() => {}));
      mockGetAgentToolCalls.mockReturnValue(new Promise(() => {}));
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      expect(screen.getByText("Loading approvals...")).toBeInTheDocument();
    });
  });

  // ---- Empty state ---------------------------------------------------------

  describe("empty state", () => {
    it('shows "No approvals yet" when there are no approvals', async () => {
      setupDefaults([]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("No approvals yet.")).toBeInTheDocument();
      });
    });
  });

  // ---- Pending approval buttons --------------------------------------------

  describe("pending approval buttons", () => {
    it("shows Approve and Reject buttons for pending approvals", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("✅ Approve")).toBeInTheDocument();
        expect(screen.getByText("❌ Reject")).toBeInTheDocument();
      });
    });

    it("does NOT show Approve/Reject buttons for non-pending approvals", async () => {
      setupDefaults([approvedApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.queryByText("✅ Approve")).toBeNull();
        expect(screen.queryByText("❌ Reject")).toBeNull();
      });
    });

    it("shows status badge instead of buttons for non-pending approvals", async () => {
      setupDefaults([approvedApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByTestId("badge-approved")).toBeInTheDocument();
      });
    });
  });

  // ---- Click actions -------------------------------------------------------

  describe("clicking approve / reject", () => {
    it('calls updateHumanApprovalStatus with "approved" on Approve click', async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockResolvedValue({
        ...pendingApproval,
        status: "approved",
      });
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("✅ Approve")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("✅ Approve"));
      await waitFor(() => {
        expect(mockUpdateHumanApprovalStatus).toHaveBeenCalledWith(
          "a1",
          "approved"
        );
      });
    });

    it('calls updateHumanApprovalStatus with "rejected" on Reject click', async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockResolvedValue({
        ...pendingApproval,
        status: "rejected",
      });
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("❌ Reject")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("❌ Reject"));
      await waitFor(() => {
        expect(mockUpdateHumanApprovalStatus).toHaveBeenCalledWith(
          "a1",
          "rejected"
        );
      });
    });
  });

  // ---- Loading spinner on button -------------------------------------------

  describe("loading spinner on clicked button", () => {
    it("shows Approving... text on clicked Approve button during API call", async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockReturnValue(new Promise(() => {}));
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("✅ Approve")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("✅ Approve"));
      await waitFor(() => {
        expect(screen.getByText("Approving...")).toBeInTheDocument();
      });
    });

    it("shows Rejecting... text on clicked Reject button during API call", async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockReturnValue(new Promise(() => {}));
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("❌ Reject")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("❌ Reject"));
      await waitFor(() => {
        expect(screen.getByText("Rejecting...")).toBeInTheDocument();
      });
    });
  });

  // ---- Error banner --------------------------------------------------------

  describe("error banner", () => {
    it("shows error banner when API call fails", async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockRejectedValue(
        new Error("Update failed")
      );
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("✅ Approve")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("✅ Approve"));
      await waitFor(() => {
        expect(screen.getByText("Update failed")).toBeInTheDocument();
      });
    });
  });

  // ---- Auto-resume ---------------------------------------------------------

  describe("auto-resume on paused run", () => {
    it("calls resumeAgentRun when run is paused and approval is acted upon", async () => {
      setupDefaults([pendingApproval]);
      mockUpdateHumanApprovalStatus.mockResolvedValue({
        ...pendingApproval,
        status: "approved",
      });
      mockResumeAgentRun.mockResolvedValue({});
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="paused"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("✅ Approve")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("✅ Approve"));
      await waitFor(() => {
        expect(mockResumeAgentRun).toHaveBeenCalledWith("r1");
      });
    });
  });

  // ---- Tool call input (request_notes) -------------------------------------

  describe("tool call input display", () => {
    it("shows expandable Tool Input details for approval with request_notes", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("Tool Input")).toBeInTheDocument();
      });
    });

    it("shows formatted JSON content in Tool Input", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(
          screen.getByText(/"key"/)
        ).toBeInTheDocument();
        expect(
          screen.getByText(/"value"/)
        ).toBeInTheDocument();
      });
    });
  });

  // ---- Tool call result ----------------------------------------------------

  describe("tool call result display", () => {
    it("shows Tool Result for approved approval with tool_call_id", async () => {
      setupDefaults([approvedApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText("Tool Result")).toBeInTheDocument();
        expect(screen.getByText(/test output/)).toBeInTheDocument();
      });
    });
  });

  // ---- Decision notes ------------------------------------------------------

  describe("decision notes", () => {
    it("shows decision notes when present", async () => {
      setupDefaults([approvedApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(screen.getByText(/Decision:/)).toBeInTheDocument();
        expect(screen.getByText(/Looks good/)).toBeInTheDocument();
      });
    });
  });

  // ---- Polling behavior ----------------------------------------------------

  describe("polling behavior", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("polls every 3 seconds while runStatus is 'running'", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );

      // Initial load
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);

      // Advance 3s
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(2);

      // Advance another 3s
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(3);
    });

    it("polls while runStatus is 'paused'", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="paused"
        />
      );

      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(2);
    });

    it("does NOT poll when runStatus is terminal", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="completed"
        />
      );

      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);
    });

    it("cleans up interval on unmount", async () => {
      setupDefaults([pendingApproval]);
      const { unmount } = render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );

      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);

      unmount();

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);
    });

    it("does NOT poll when WebSocket is connected", async () => {
      mockUseWebSocket.mockReturnValue({ isConnected: true, reconnect: vi.fn() });
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );

      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);

      // Advance 3s – should NOT poll because WebSocket is connected
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetHumanApprovals).toHaveBeenCalledTimes(1);
    });
  });

  // ---- WebSocket integration -----------------------------------------------

  describe("WebSocket integration", () => {
    it("connects WebSocket when runStatus is running", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="running"
        />
      );
      await waitFor(() => {
        expect(mockUseWebSocket).toHaveBeenCalledWith(
          expect.objectContaining({ runId: "r1" })
        );
      });
    });

    it("connects WebSocket when runStatus is paused", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="paused"
        />
      );
      await waitFor(() => {
        expect(mockUseWebSocket).toHaveBeenCalledWith(
          expect.objectContaining({ runId: "r1" })
        );
      });
    });

    it("does not connect WebSocket when runStatus is completed", async () => {
      setupDefaults([pendingApproval]);
      render(
        <HumanApprovalsPanel
          runId="r1"
          steps={mockSteps}
          runStatus="completed"
        />
      );
      await waitFor(() => {
        const lastCall = mockUseWebSocket.mock.calls.at(-1)?.[0];
        expect(lastCall?.runId).toBeNull();
      });
    });
  });
});
