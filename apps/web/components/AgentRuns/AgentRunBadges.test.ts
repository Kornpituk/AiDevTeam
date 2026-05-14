import { describe, it, expect } from "vitest";
import {
  formatRunStatus,
  formatStepStatus,
  formatApprovalStatus,
  formatToolCallStatus,
  formatMessageRole,
  getRunStatusVariant,
  getStepStatusVariant,
  getApprovalStatusVariant,
  getToolCallStatusVariant,
  getMessageRoleVariant,
} from "./AgentRunBadges";

describe("formatRunStatus", () => {
  it("formats all 9 statuses correctly", () => {
    expect(formatRunStatus("draft")).toBe("Draft");
    expect(formatRunStatus("planned")).toBe("Planned");
    expect(formatRunStatus("waiting_approval")).toBe("Waiting Approval");
    expect(formatRunStatus("approved")).toBe("Approved");
    expect(formatRunStatus("running")).toBe("Running");
    expect(formatRunStatus("paused")).toBe("Paused");
    expect(formatRunStatus("completed")).toBe("Completed");
    expect(formatRunStatus("failed")).toBe("Failed");
    expect(formatRunStatus("cancelled")).toBe("Cancelled");
  });
});

describe("formatStepStatus", () => {
  it("formats all 7 statuses correctly", () => {
    expect(formatStepStatus("pending")).toBe("Pending");
    expect(formatStepStatus("waiting_approval")).toBe("Waiting Approval");
    expect(formatStepStatus("running")).toBe("Running");
    expect(formatStepStatus("completed")).toBe("Completed");
    expect(formatStepStatus("failed")).toBe("Failed");
    expect(formatStepStatus("skipped")).toBe("Skipped");
    expect(formatStepStatus("cancelled")).toBe("Cancelled");
  });
});

describe("formatApprovalStatus", () => {
  it("formats all 4 statuses correctly", () => {
    expect(formatApprovalStatus("pending")).toBe("Pending");
    expect(formatApprovalStatus("approved")).toBe("Approved");
    expect(formatApprovalStatus("rejected")).toBe("Rejected");
    expect(formatApprovalStatus("cancelled")).toBe("Cancelled");
  });
});

describe("formatToolCallStatus", () => {
  it("formats all 5 statuses correctly", () => {
    expect(formatToolCallStatus("recorded")).toBe("Recorded");
    expect(formatToolCallStatus("approved")).toBe("Approved");
    expect(formatToolCallStatus("rejected")).toBe("Rejected");
    expect(formatToolCallStatus("completed")).toBe("Completed");
    expect(formatToolCallStatus("failed")).toBe("Failed");
  });
});

describe("formatMessageRole", () => {
  it("formats all 5 roles correctly", () => {
    expect(formatMessageRole("system")).toBe("System");
    expect(formatMessageRole("user")).toBe("User");
    expect(formatMessageRole("assistant")).toBe("Assistant");
    expect(formatMessageRole("tool")).toBe("Tool");
    expect(formatMessageRole("reviewer")).toBe("Reviewer");
  });
});

describe("getRunStatusVariant", () => {
  it("returns success for completed and approved", () => {
    expect(getRunStatusVariant("completed")).toBe("success");
    expect(getRunStatusVariant("approved")).toBe("success");
  });

  it("returns error for failed and cancelled", () => {
    expect(getRunStatusVariant("failed")).toBe("error");
    expect(getRunStatusVariant("cancelled")).toBe("error");
  });

  it("returns info for running", () => {
    expect(getRunStatusVariant("running")).toBe("info");
  });

  it("returns warning for waiting_approval and paused", () => {
    expect(getRunStatusVariant("waiting_approval")).toBe("warning");
    expect(getRunStatusVariant("paused")).toBe("warning");
  });

  it("returns default for draft and planned", () => {
    expect(getRunStatusVariant("draft")).toBe("default");
    expect(getRunStatusVariant("planned")).toBe("default");
  });

  it("covers all 9 statuses", () => {
    const statuses = [
      "draft",
      "planned",
      "waiting_approval",
      "approved",
      "running",
      "paused",
      "completed",
      "failed",
      "cancelled",
    ] as const;
    statuses.forEach((status) => {
      const variant = getRunStatusVariant(status);
      expect(["default", "success", "warning", "error", "info"]).toContain(
        variant
      );
    });
  });
});

describe("getStepStatusVariant", () => {
  it("returns success for completed", () => {
    expect(getStepStatusVariant("completed")).toBe("success");
  });

  it("returns error for failed and cancelled", () => {
    expect(getStepStatusVariant("failed")).toBe("error");
    expect(getStepStatusVariant("cancelled")).toBe("error");
  });

  it("returns info for running", () => {
    expect(getStepStatusVariant("running")).toBe("info");
  });

  it("returns warning for waiting_approval", () => {
    expect(getStepStatusVariant("waiting_approval")).toBe("warning");
  });

  it("returns default for pending and skipped", () => {
    expect(getStepStatusVariant("pending")).toBe("default");
    expect(getStepStatusVariant("skipped")).toBe("default");
  });

  it("covers all 7 statuses", () => {
    const statuses = [
      "pending",
      "waiting_approval",
      "running",
      "completed",
      "failed",
      "skipped",
      "cancelled",
    ] as const;
    statuses.forEach((status) => {
      const variant = getStepStatusVariant(status);
      expect(["default", "success", "warning", "error", "info"]).toContain(
        variant
      );
    });
  });
});

describe("getApprovalStatusVariant", () => {
  it("returns success for approved", () => {
    expect(getApprovalStatusVariant("approved")).toBe("success");
  });

  it("returns error for rejected and cancelled", () => {
    expect(getApprovalStatusVariant("rejected")).toBe("error");
    expect(getApprovalStatusVariant("cancelled")).toBe("error");
  });

  it("returns warning for pending", () => {
    expect(getApprovalStatusVariant("pending")).toBe("warning");
  });

  it("covers all 4 statuses", () => {
    const statuses = ["pending", "approved", "rejected", "cancelled"] as const;
    statuses.forEach((status) => {
      const variant = getApprovalStatusVariant(status);
      expect(["default", "success", "warning", "error", "info"]).toContain(
        variant
      );
    });
  });
});

describe("getToolCallStatusVariant", () => {
  it("returns success for completed and approved", () => {
    expect(getToolCallStatusVariant("completed")).toBe("success");
    expect(getToolCallStatusVariant("approved")).toBe("success");
  });

  it("returns error for failed and rejected", () => {
    expect(getToolCallStatusVariant("failed")).toBe("error");
    expect(getToolCallStatusVariant("rejected")).toBe("error");
  });

  it("returns info for recorded", () => {
    expect(getToolCallStatusVariant("recorded")).toBe("info");
  });

  it("covers all 5 statuses", () => {
    const statuses = [
      "recorded",
      "approved",
      "rejected",
      "completed",
      "failed",
    ] as const;
    statuses.forEach((status) => {
      const variant = getToolCallStatusVariant(status);
      expect(["default", "success", "warning", "error", "info"]).toContain(
        variant
      );
    });
  });
});

describe("getMessageRoleVariant", () => {
  it("returns info for assistant", () => {
    expect(getMessageRoleVariant("assistant")).toBe("info");
  });

  it("returns success for user and reviewer", () => {
    expect(getMessageRoleVariant("user")).toBe("success");
    expect(getMessageRoleVariant("reviewer")).toBe("success");
  });

  it("returns warning for system", () => {
    expect(getMessageRoleVariant("system")).toBe("warning");
  });

  it("returns default for tool", () => {
    expect(getMessageRoleVariant("tool")).toBe("default");
  });

  it("covers all 5 roles", () => {
    const roles = [
      "system",
      "user",
      "assistant",
      "tool",
      "reviewer",
    ] as const;
    roles.forEach((role) => {
      const variant = getMessageRoleVariant(role);
      expect(["default", "success", "warning", "error", "info"]).toContain(
        variant
      );
    });
  });
});
