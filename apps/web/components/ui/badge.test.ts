import { formatStatus, getStatusVariant } from "@/components/ui/badge";
import { TASK_STATUSES, type TaskStatus } from "@/lib/api";

describe("formatStatus", () => {
  it("formats snake_case to Title Case", () => {
    expect(formatStatus("pending")).toBe("Pending");
    expect(formatStatus("in_progress")).toBe("In Progress");
    expect(formatStatus("completed")).toBe("Completed");
    expect(formatStatus("approved")).toBe("Approved");
    expect(formatStatus("failed")).toBe("Failed");
    expect(formatStatus("planning")).toBe("Planning");
    expect(formatStatus("reviewing")).toBe("Reviewing");
  });
});

describe("getStatusVariant", () => {
  it("returns success for completed and approved", () => {
    expect(getStatusVariant("completed")).toBe("success");
    expect(getStatusVariant("approved")).toBe("success");
  });

  it("returns error for failed", () => {
    expect(getStatusVariant("failed")).toBe("error");
  });

  it("returns info for planning and in_progress", () => {
    expect(getStatusVariant("planning")).toBe("info");
    expect(getStatusVariant("in_progress")).toBe("info");
  });

  it("returns warning for reviewing", () => {
    expect(getStatusVariant("reviewing")).toBe("warning");
  });

  it("returns default for pending", () => {
    expect(getStatusVariant("pending")).toBe("default");
  });

  it("covers all statuses", () => {
    const statuses: TaskStatus[] = ["pending", "planning", "approved", "in_progress", "reviewing", "completed", "failed"];
    statuses.forEach((status) => {
      const variant = getStatusVariant(status);
      expect(["default", "success", "warning", "error", "info"]).toContain(variant);
    });
  });
});

describe("TASK_STATUSES", () => {
  it("contains all expected statuses", () => {
    expect(TASK_STATUSES).toEqual([
      "pending",
      "planning",
      "approved",
      "in_progress",
      "reviewing",
      "completed",
      "failed",
    ]);
  });
});
