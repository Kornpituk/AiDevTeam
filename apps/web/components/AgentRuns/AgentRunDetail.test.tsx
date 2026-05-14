import { render, screen, waitFor, act, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { AgentRunDetail } from "./AgentRunDetail";
import type { AgentRun, AgentRunStep, AgentProfile, AgentTeam } from "@/lib/api";

// ---------------------------------------------------------------------------
// Mocks – use vi.hoisted() so mock fns are created before vi.mock runs
// ---------------------------------------------------------------------------

const mockGetAgentRun = vi.hoisted(() => vi.fn());
const mockGetAgentRunSteps = vi.hoisted(() => vi.fn());
const mockGetAgentProfiles = vi.hoisted(() => vi.fn());
const mockGetAgentTeams = vi.hoisted(() => vi.fn());
const mockStartAgentRun = vi.hoisted(() => vi.fn());
const mockCancelAgentRun = vi.hoisted(() => vi.fn());
const mockResumeAgentRun = vi.hoisted(() => vi.fn());
const mockUseWebSocket = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api", () => ({
  getAgentRun: mockGetAgentRun,
  getAgentRunSteps: mockGetAgentRunSteps,
  getAgentProfiles: mockGetAgentProfiles,
  getAgentTeams: mockGetAgentTeams,
  startAgentRun: mockStartAgentRun,
  cancelAgentRun: mockCancelAgentRun,
  resumeAgentRun: mockResumeAgentRun,
}));

vi.mock("@/lib/hooks/useWebSocket", () => ({
  useWebSocket: mockUseWebSocket,
}));

// next/link
vi.mock("next/link", () => {
  const React = require("react");
  return {
    default: ({
      children,
      href,
    }: {
      children: React.ReactNode;
      href: string;
    }) => React.createElement("a", { href }, children),
  };
});

// Sub-components (render minimal placeholders)
vi.mock("./RunStepsList", () => ({
  RunStepsList: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "run-steps-list" });
  },
}));

vi.mock("./AddRunStepForm", () => ({
  AddRunStepForm: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "add-step-form" });
  },
}));

vi.mock("./AgentMessagesPanel", () => ({
  AgentMessagesPanel: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "agent-messages-panel" });
  },
}));

vi.mock("./HumanApprovalsPanel", () => {
  const React = require("react");
  return {
    HumanApprovalsPanel: (props: { runStatus?: string }) =>
      React.createElement("div", {
        "data-testid": "human-approvals-panel",
        "data-run-status": props.runStatus,
      }),
  };
});

vi.mock("./ToolCallsPanel", () => ({
  ToolCallsPanel: () => {
    const React = require("react");
    return React.createElement("div", { "data-testid": "tool-calls-panel" });
  },
}));

// ---------------------------------------------------------------------------
// Shared test data
// ---------------------------------------------------------------------------

const baseRun: AgentRun = {
  id: "test-run-1",
  task_id: "task-1",
  team_id: "team-1",
  status: "draft",
  goal: "Test goal",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const emptySteps: AgentRunStep[] = [];
const emptyProfiles: AgentProfile[] = [];
const testTeams: AgentTeam[] = [
  {
    id: "team-1",
    name: "Test Team",
    description: "",
    created_at: "",
    updated_at: "",
  },
];

function setupDefaults(overrides: Partial<AgentRun> = {}) {
  mockGetAgentRun.mockResolvedValue({ ...baseRun, ...overrides });
  mockGetAgentRunSteps.mockResolvedValue(emptySteps);
  mockGetAgentProfiles.mockResolvedValue(emptyProfiles);
  mockGetAgentTeams.mockResolvedValue(testTeams);
  mockUseWebSocket.mockReturnValue({ isConnected: false, reconnect: vi.fn() });
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("AgentRunDetail", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseWebSocket.mockReturnValue({ isConnected: false, reconnect: vi.fn() });
  });

  // ---- Loading state -------------------------------------------------------

  describe("loading state", () => {
    it("shows loading indicator on mount", () => {
      mockGetAgentRun.mockReturnValue(new Promise(() => {}));
      mockGetAgentRunSteps.mockReturnValue(new Promise(() => {}));
      mockGetAgentProfiles.mockReturnValue(new Promise(() => {}));
      mockGetAgentTeams.mockReturnValue(new Promise(() => {}));

      render(<AgentRunDetail runId="test-run-1" />);
      expect(screen.getByText("Loading run...")).toBeInTheDocument();
    });
  });

  // ---- Button visibility ---------------------------------------------------

  describe("button visibility based on run status", () => {
    it.each([
      ["draft", "Start Run"],
      ["planned", "Start Run"],
      ["waiting_approval", "Start Run"],
      ["approved", "Start Run"],
    ])("shows Start Run button when status is '%s'", async (status, label) => {
      setupDefaults({ status: status as AgentRun["status"] });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: label })
        ).toBeInTheDocument();
      });
    });

    it("shows Cancel Run button when status is 'running'", async () => {
      setupDefaults({ status: "running" });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Cancel Run" })
        ).toBeInTheDocument();
      });
    });

    it("shows Resume Run button when status is 'paused'", async () => {
      setupDefaults({ status: "paused" });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /Resume Run/ })
        ).toBeInTheDocument();
      });
    });

    it.each(["completed", "failed", "cancelled"] as const)(
      "shows no action buttons when status is '%s'",
      async (status) => {
        setupDefaults({ status });
        render(<AgentRunDetail runId="test-run-1" />);
        await waitFor(() => {
          expect(
            screen.queryByRole("button", { name: "Start Run" })
          ).toBeNull();
          expect(
            screen.queryByRole("button", { name: "Cancel Run" })
          ).toBeNull();
          expect(
            screen.queryByRole("button", { name: /Resume Run/ })
          ).toBeNull();
        });
      }
    );
  });

  // ---- Loading states on buttons -------------------------------------------

  describe("action loading states", () => {
    it("disables buttons when actionLoading is not null (Start)", async () => {
      setupDefaults({ status: "draft" });
      mockStartAgentRun.mockReturnValue(new Promise(() => {}));
      render(<AgentRunDetail runId="test-run-1" />);

      const startBtn = await screen.findByRole("button", {
        name: "Start Run",
      });
      fireEvent.click(startBtn);

      await waitFor(() => {
        const btn = screen.getByRole("button", { name: /Starting/ });
        expect(btn).toBeDisabled();
      });
    });

    it("shows 'Starting...' text when starting", async () => {
      setupDefaults({ status: "draft" });
      mockStartAgentRun.mockReturnValue(new Promise(() => {}));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: "Start Run" })
      );

      await waitFor(() => {
        expect(screen.getByText("Starting...")).toBeInTheDocument();
      });
    });

    it("shows 'Cancelling...' text when cancelling", async () => {
      setupDefaults({ status: "running" });
      mockCancelAgentRun.mockReturnValue(new Promise(() => {}));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: "Cancel Run" })
      );

      await waitFor(() => {
        expect(screen.getByText("Cancelling...")).toBeInTheDocument();
      });
    });

    it("shows 'Resuming...' text when resuming", async () => {
      setupDefaults({ status: "paused" });
      mockResumeAgentRun.mockReturnValue(new Promise(() => {}));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: /Resume Run/ })
      );

      await waitFor(() => {
        expect(screen.getByText("Resuming...")).toBeInTheDocument();
      });
    });
  });

  // ---- Error handling -------------------------------------------------------

  describe("error handling on actions", () => {
    it("shows error message when start fails", async () => {
      setupDefaults({ status: "draft" });
      mockStartAgentRun.mockRejectedValue(new Error("Start failed"));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: "Start Run" })
      );

      await waitFor(() => {
        expect(screen.getByText("Start failed")).toBeInTheDocument();
      });
    });

    it("shows error message when cancel fails", async () => {
      setupDefaults({ status: "running" });
      mockCancelAgentRun.mockRejectedValue(new Error("Cancel failed"));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: "Cancel Run" })
      );

      await waitFor(() => {
        expect(screen.getByText("Cancel failed")).toBeInTheDocument();
      });
    });

    it("shows error message when resume fails", async () => {
      setupDefaults({ status: "paused" });
      mockResumeAgentRun.mockRejectedValue(new Error("Resume failed"));
      render(<AgentRunDetail runId="test-run-1" />);

      fireEvent.click(
        await screen.findByRole("button", { name: /Resume Run/ })
      );

      await waitFor(() => {
        expect(screen.getByText("Resume failed")).toBeInTheDocument();
      });
    });
  });

  // ---- Approval panel status prop -------------------------------------------

  describe("HumanApprovalsPanel receives runStatus prop", () => {
    it("passes the run status to the HumanApprovalsPanel", async () => {
      setupDefaults({ status: "running" });
      render(<AgentRunDetail runId="test-run-1" />);

      await waitFor(() => {
        const panel = screen.getByTestId("human-approvals-panel");
        expect(panel.getAttribute("data-run-status")).toBe("running");
      });
    });
  });

  // ---- Polling behavior -----------------------------------------------------

  describe("polling behavior", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    function setupPolling(status: AgentRun["status"]) {
      setupDefaults({ status });
      mockGetAgentRun.mockResolvedValue({ ...baseRun, status });
    }

    it("polls every 3 seconds when status is 'running'", async () => {
      setupPolling("running");
      render(<AgentRunDetail runId="test-run-1" />);

      // Let initial useEffect complete
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);

      // Advance 3s to trigger first poll
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(2);

      // Advance another 3s
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(3);
    });

    it("polls when status is 'paused'", async () => {
      setupPolling("paused");
      render(<AgentRunDetail runId="test-run-1" />);

      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(2);
    });

    it("stops polling when status becomes terminal", async () => {
      mockGetAgentRun
        .mockResolvedValueOnce({ ...baseRun, status: "running" })
        .mockResolvedValue({ ...baseRun, status: "completed" });
      mockGetAgentRunSteps.mockResolvedValue(emptySteps);
      mockGetAgentProfiles.mockResolvedValue(emptyProfiles);
      mockGetAgentTeams.mockResolvedValue(testTeams);

      render(<AgentRunDetail runId="test-run-1" />);

      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);

      // Advance 3s – poll fires, but now status is "completed" so interval clears
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(2);

      // Advance another 3s – should NOT poll again
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(2);
    });

    it("cleans up interval on unmount", async () => {
      setupPolling("running");
      const { unmount } = render(<AgentRunDetail runId="test-run-1" />);

      await act(async () => {});
      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);

      unmount();

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});

      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);
    });
  });

  // ---- WebSocket integration -----------------------------------------------

  describe("WebSocket integration", () => {
    it("shows Live indicator when WebSocket is connected", async () => {
      setupDefaults({ status: "running" });
      mockUseWebSocket.mockReturnValue({ isConnected: true, reconnect: vi.fn() });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(screen.getByText("Live")).toBeInTheDocument();
      });
    });

    it("does not show Live indicator when WebSocket is disconnected", async () => {
      setupDefaults({ status: "running" });
      mockUseWebSocket.mockReturnValue({ isConnected: false, reconnect: vi.fn() });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(screen.queryByText("Live")).toBeNull();
      });
    });

    it("connects WebSocket when run status is running", async () => {
      setupDefaults({ status: "running" });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(mockUseWebSocket).toHaveBeenCalledWith(
          expect.objectContaining({ runId: "test-run-1" })
        );
      });
    });

    it("connects WebSocket when run status is paused", async () => {
      setupDefaults({ status: "paused" });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        expect(mockUseWebSocket).toHaveBeenCalledWith(
          expect.objectContaining({ runId: "test-run-1" })
        );
      });
    });

    it("does not connect WebSocket when run status is terminal", async () => {
      setupDefaults({ status: "completed" });
      render(<AgentRunDetail runId="test-run-1" />);
      await waitFor(() => {
        const lastCall = mockUseWebSocket.mock.calls.at(-1)?.[0];
        expect(lastCall?.runId).toBeNull();
      });
    });

    it("does not poll when WebSocket is connected", async () => {
      vi.useFakeTimers();
      mockGetAgentRun.mockResolvedValue({ ...baseRun, status: "running" });
      mockGetAgentRunSteps.mockResolvedValue(emptySteps);
      mockGetAgentProfiles.mockResolvedValue(emptyProfiles);
      mockGetAgentTeams.mockResolvedValue(testTeams);
      mockUseWebSocket.mockReturnValue({ isConnected: true, reconnect: vi.fn() });

      render(<AgentRunDetail runId="test-run-1" />);
      await act(async () => {});

      // Advance 3s – should NOT poll because WebSocket is connected
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      await act(async () => {});

      // Only 1 call from the initial fetch
      expect(mockGetAgentRun).toHaveBeenCalledTimes(1);

      vi.useRealTimers();
    });
  });
});
