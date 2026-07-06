import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ExamMonitorPage from "./page";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("ExamMonitorPage", () => {
  it("renders UnderMaintenance with Session Monitor title", async () => {
    render(<ExamMonitorPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Monitor Sesi/i })
      ).toBeInTheDocument();
    });

    expect(
      screen.getByText("Fitur ini sedang dalam pengembangan")
    ).toBeInTheDocument();
  });
});