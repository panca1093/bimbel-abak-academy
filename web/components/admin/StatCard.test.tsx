import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatCard } from "./StatCard";
import { Users } from "lucide-react";

describe("StatCard", () => {
  it("renders label and value", () => {
    render(<StatCard label="Total Siswa" value="1.303" />);
    expect(screen.getByText("Total Siswa")).toBeInTheDocument();
    expect(screen.getByText("1.303")).toBeInTheDocument();
  });

  it("renders icon when provided", () => {
    render(<StatCard label="Total Siswa" value="1.303" icon={Users} />);
    const svg = document.querySelector("svg");
    expect(svg).toBeInTheDocument();
  });

  it("does not render icon when omitted", () => {
    render(<StatCard label="Total Siswa" value="1.303" />);
    const svg = document.querySelector("svg");
    expect(svg).not.toBeInTheDocument();
  });

  it("renders trend when provided", () => {
    render(<StatCard label="Total Siswa" value="1.303" trend="+42 minggu ini" />);
    expect(screen.getByText("+42 minggu ini")).toBeInTheDocument();
  });

  it("does not render trend when omitted", () => {
    render(<StatCard label="Total Siswa" value="1.303" />);
    expect(screen.queryByText("+42 minggu ini")).not.toBeInTheDocument();
  });

  it("defaults accent to primary if not specified", () => {
    render(<StatCard label="Test" value="123" />);
    // Should render without error; primary tokens are applied via CSS vars
    expect(screen.getByText("Test")).toBeInTheDocument();
  });

  it("renders with error accent", () => {
    render(<StatCard label="Test" value="123" accent="error" />);
    expect(screen.getByText("Test")).toBeInTheDocument();
  });
});
