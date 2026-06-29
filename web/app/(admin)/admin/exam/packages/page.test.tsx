import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ExamPackagesPage from "./page";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("ExamPackagesPage", () => {
  it("renders UnderMaintenance with Packages title", async () => {
    render(<ExamPackagesPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Paket/i })
      ).toBeInTheDocument();
    });

    expect(
      screen.getByText("Fitur ini sedang dalam pengembangan")
    ).toBeInTheDocument();
  });
});