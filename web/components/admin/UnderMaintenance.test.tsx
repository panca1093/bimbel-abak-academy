import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { UnderMaintenance } from "./UnderMaintenance";
import { Wrench, Clock } from "lucide-react";

describe("UnderMaintenance", () => {
  it("renders the icon in a rounded container", () => {
    render(<UnderMaintenance icon={Wrench} title="Tryout / Ujian" />);
    const svg = document.querySelector("svg");
    expect(svg).toBeInTheDocument();
    const container = svg?.closest("div");
    expect(container).toHaveClass("rounded-[20px]");
  });

  it("renders the heading with title", () => {
    render(<UnderMaintenance icon={Wrench} title="Tryout / Ujian" />);
    expect(screen.getByRole("heading", { name: "Tryout / Ujian" })).toBeInTheDocument();
  });

  it("renders description paragraph with default text when not provided", () => {
    render(<UnderMaintenance icon={Wrench} title="Tryout / Ujian" />);
    expect(screen.getByText("Fitur ini sedang dalam pengembangan")).toBeInTheDocument();
  });

  it("renders custom description when provided", () => {
    render(
      <UnderMaintenance
        icon={Wrench}
        title="Tryout / Ujian"
        description="Custom message"
      />
    );
    expect(screen.getByText("Custom message")).toBeInTheDocument();
    expect(screen.queryByText("Fitur ini sedang dalam pengembangan")).not.toBeInTheDocument();
  });

  it("renders estimated timeline note in a subdued style when provided", () => {
    render(
      <UnderMaintenance
        icon={Wrench}
        title="Tryout / Ujian"
        estimatedTimeline="Estimasi: Q3 2026"
      />
    );
    const timeline = screen.getByText("Estimasi: Q3 2026");
    expect(timeline).toBeInTheDocument();
    expect(timeline.className).toMatch(/on-surface-variant|subdued|muted/i);
  });

  it("does not render timeline note when omitted", () => {
    render(<UnderMaintenance icon={Wrench} title="Tryout / Ujian" />);
    expect(screen.queryByText(/estimasi|timeline/i)).not.toBeInTheDocument();
  });

  it("renders all parts together: icon, heading, description, timeline", () => {
    render(
      <UnderMaintenance
        icon={Clock}
        title="Bank Soal"
        description="Fitur ini sedang dalam pengembangan"
        estimatedTimeline="Estimasi: Q3 2026"
      />
    );
    expect(document.querySelector("svg")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Bank Soal" })).toBeInTheDocument();
    expect(screen.getByText("Fitur ini sedang dalam pengembangan")).toBeInTheDocument();
    expect(screen.getByText("Estimasi: Q3 2026")).toBeInTheDocument();
  });

  it("does not contain hardcoded mock data", () => {
    render(<UnderMaintenance icon={Wrench} title="Some Title" />);
    // Neither the heading nor the body should reference real exam data
    expect(screen.queryByText(/tryout|ujian|bank soal|jadwal|analitik/i)).not.toBeInTheDocument();
    // The component should only render what's passed
    expect(screen.getByRole("heading", { name: "Some Title" })).toBeInTheDocument();
  });
});
