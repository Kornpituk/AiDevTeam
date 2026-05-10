import { cn, formatDate } from "@/lib/utils";

describe("cn", () => {
  it("merges class names", () => {
    expect(cn("foo", "bar")).toBe("foo bar");
  });

  it("handles conditional classes", () => {
    expect(cn("foo", false && "bar", "baz")).toBe("foo baz");
  });

  it("resolves tailwind conflicts", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
  });

  it("handles undefined and null", () => {
    expect(cn("foo", undefined, null, "bar")).toBe("foo bar");
  });
});

describe("formatDate", () => {
  it("formats ISO date string", () => {
    const result = formatDate("2024-01-15T10:30:00Z");
    expect(result).toContain("2024");
    expect(result).toContain("Jan");
    expect(result).toContain("15");
  });
});
