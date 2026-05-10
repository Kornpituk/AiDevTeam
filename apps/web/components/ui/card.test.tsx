import { render, screen } from "@testing-library/react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";

describe("Card", () => {
  it("renders Card component", () => {
    render(<Card data-testid="card">Content</Card>);
    const card = screen.getByTestId("card");
    expect(card).toBeInTheDocument();
    expect(card.className).toContain("bg-white");
  });

  it("renders full card structure", () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>Title</CardTitle>
          <CardDescription>Description</CardDescription>
        </CardHeader>
        <CardContent>Content</CardContent>
        <CardFooter>Footer</CardFooter>
      </Card>
    );

    expect(screen.getByText("Title")).toBeInTheDocument();
    expect(screen.getByText("Description")).toBeInTheDocument();
    expect(screen.getByText("Content")).toBeInTheDocument();
    expect(screen.getByText("Footer")).toBeInTheDocument();
  });

  it("CardHeader has border styling", () => {
    render(<CardHeader data-testid="header">Header</CardHeader>);
    const header = screen.getByTestId("header");
    expect(header.className).toContain("border-b");
  });

  it("CardFooter has border styling", () => {
    render(<CardFooter data-testid="footer">Footer</CardFooter>);
    const footer = screen.getByTestId("footer");
    expect(footer.className).toContain("border-t");
  });
});
