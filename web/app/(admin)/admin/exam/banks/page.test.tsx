import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ExamBanksPage from "./page";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("ExamBanksPage", () => {
  it("renders UnderMaintenance with Bank Soal title", async () => {
    render(<ExamBanksPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Bank Soal/i })
      ).toBeInTheDocument();
    });

    expect(
      screen.getByText("Fitur ini sedang dalam pengembangan")
    ).toBeInTheDocument();
  });
});
