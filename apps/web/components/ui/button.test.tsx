import { render, screen } from "@testing-library/react";
import { Button } from "@/components/ui/button";
import { describe, it, expect, vi } from "vitest";

describe("Button", () => {
  it("renders with default props", () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole("button", { name: "Click me" });
    expect(button).toBeInTheDocument();
  });

  it("applies primary variant by default", () => {
    render(<Button>Primary</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toContain("from-blue-600");
  });

  it("applies secondary variant", () => {
    render(<Button variant="secondary">Secondary</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toContain("bg-gray-100");
  });

  it("applies outline variant", () => {
    render(<Button variant="outline">Outline</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toContain("border-gray-300");
  });

  it("applies ghost variant", () => {
    render(<Button variant="ghost">Ghost</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toContain("hover:bg-gray-100");
  });

  it("applies different sizes", () => {
    const { rerender } = render(<Button size="sm">Small</Button>);
    let button = screen.getByRole("button");
    expect(button.className).toContain("text-sm");

    rerender(<Button size="lg">Large</Button>);
    button = screen.getByRole("button");
    expect(button.className).toContain("text-lg");
  });

  it("handles disabled state", () => {
    render(<Button disabled>Disabled</Button>);
    const button = screen.getByRole("button");
    expect(button).toBeDisabled();
    expect(button.className).toContain("disabled:opacity-50");
  });
});
