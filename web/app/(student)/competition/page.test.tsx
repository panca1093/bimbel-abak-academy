import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import CompetitionPage from "./page";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), message: vi.fn() },
}));

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("CompetitionPage", () => {
  it("renders title and package sections", async () => {
    render(<CompetitionPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Kompetisi & Tryout/i })
      ).toBeInTheDocument();
      expect(screen.getByText(/Paket gratis/i)).toBeInTheDocument();
      expect(screen.getByText(/Paket saya/i)).toBeInTheDocument();
    });
  });

  it("lists free and purchased packages", async () => {
    render(<CompetitionPage />);

    await waitFor(() => {
      expect(screen.getByText("Try Out UTBK Gratis #12")).toBeInTheDocument();
      expect(
        screen.getByText("Simulasi SNBT Nasional 2026")
      ).toBeInTheDocument();
    });
  });

  it("shows check-in token input for checkin package", async () => {
    render(<CompetitionPage />);

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(/Token dari kartu ujian/i)
      ).toBeInTheDocument();
    });
  });

  it("shows score for submitted package", async () => {
    render(<CompetitionPage />);

    await waitFor(() => {
      expect(screen.getByText("7.5")).toBeInTheDocument();
    });
  });
});
