import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

import SessionResultPage from "./page";
import type { SessionResult } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "session-1" }),
}));

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

let resultState = {
  data: null as SessionResult | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/exam", () => ({
  useSessionResult: () => resultState,
}));

describe("SessionResultPage", () => {
  beforeEach(() => {
    uiStore = { lang: "id" };
    resultState = {
      data: null,
      isLoading: true,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("shows loading skeleton while fetching (FR-S5-26)", () => {
    render(<SessionResultPage />);
    expect(screen.getByText("Memuat…")).toBeInTheDocument();
  });

  it("shows error card with retry when the fetch fails (FR-S5-26)", () => {
    resultState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("not found"),
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText(/gagal memuat data/i)).toBeInTheDocument();
  });

  it("renders hidden state (FR-S5-27)", () => {
    resultState = {
      data: { state: "hidden" },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText("Hasil belum dipublikasikan.")).toBeInTheDocument();
  });

  it("renders grading state (FR-S5-27)", () => {
    resultState = {
      data: { state: "grading" },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText("Sedang dinilai")).toBeInTheDocument();
  });

  it("renders locked state with formatted result_release_at (FR-S5-27)", () => {
    resultState = {
      data: {
        state: "locked",
        result_release_at: "2026-08-01T10:00:00Z",
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText(/Hasil tersedia pada/)).toBeInTheDocument();
    expect(screen.getByText(/2026/)).toBeInTheDocument();
  });

  it("renders score card for score_only result without breakdown (FR-S5-26)", () => {
    resultState = {
      data: {
        state: "result",
        result_config: "score_only",
        score: 80,
        correct_count: 8,
        wrong_count: 1,
        empty_count: 1,
        rank: 3,
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText("80")).toBeInTheDocument();
    expect(screen.getByText("#3")).toBeInTheDocument();
    expect(screen.queryByText("Berdasarkan Topik")).not.toBeInTheDocument();
    expect(screen.queryByText("Pembahasan")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sertifikat" })).toBeDisabled();
  });

  it("renders breakdown and pembahasan for score_pembahasan result (FR-S5-26)", () => {
    resultState = {
      data: {
        state: "result",
        result_config: "score_pembahasan",
        score: 90,
        correct_count: 9,
        wrong_count: 0,
        empty_count: 1,
        rank: 1,
        breakdown: [
          {
            test_id: "test-1",
            title: "Tes Matematika",
            subject: "Matematika",
            topic: "Aljabar",
            earned: 8,
            max: 10,
          },
        ],
        pembahasan: [
          {
            question_id: "q-1",
            body: "Berapa 2+2?",
            format: "mcq",
            your_answer: "4",
            correct_answer: "4",
            is_correct: true,
            explanation: "2+2 sama dengan 4.",
          },
        ],
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<SessionResultPage />);
    expect(screen.getByText("Berdasarkan Topik")).toBeInTheDocument();
    expect(screen.getByText("Tes Matematika")).toBeInTheDocument();
    expect(screen.getByText("8/10")).toBeInTheDocument();
    expect(screen.getByText("Pembahasan")).toBeInTheDocument();
    expect(screen.getByText(/Berapa 2\+2\?/)).toBeInTheDocument();
  });
});
