import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AdminPageHeader } from "./AdminPageHeader";
import { Shield } from "lucide-react";

describe("AdminPageHeader", () => {
  it("renders the icon and title", () => {
    render(<AdminPageHeader icon={Shield} title="Dashboard" />);
    // The icon renders inside an SVG element
    const svg = document.querySelector("svg");
    expect(svg).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1, name: "Dashboard" })).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(<AdminPageHeader icon={Shield} title="Dashboard" description="A test description" />);
    expect(screen.getByText("A test description")).toBeInTheDocument();
  });

  it("does not render description when omitted", () => {
    render(<AdminPageHeader icon={Shield} title="Dashboard" />);
    expect(screen.queryByRole("paragraph")).not.toBeInTheDocument();
  });

  it("renders actions when provided", () => {
    render(
      <AdminPageHeader icon={Shield} title="Dashboard" actions={<button>Action</button>} />
    );
    expect(screen.getByRole("button", { name: "Action" })).toBeInTheDocument();
  });

  it("does not render actions section when omitted", () => {
    render(<AdminPageHeader icon={Shield} title="Dashboard" />);
    // The actions div should not exist
    expect(screen.queryByText("Action")).not.toBeInTheDocument();
  });
});
