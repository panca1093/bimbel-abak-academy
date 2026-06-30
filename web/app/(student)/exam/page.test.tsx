import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import ExamPage from "./page";
import type { RegistrationListItem } from "@/lib/types";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

let registrationsState = {
  data: null as RegistrationListItem[] | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/exam", () => ({
  useRegistrations: () => registrationsState,
}));

const sample: RegistrationListItem[] = [
  {
    id: "reg-1",
    student_id: "s-1",
    exam_id: "e-1",
    token: "ABC12345",
    card_pdf_url: null,
    checked_in_at: null,
    attempts_used: 0,
    status: "registered",
    created_at: "2026-06-01T00:00:00Z",
    exam_title: "Try Out UTBK Gratis #12",
    scheduled_at: "2026-07-15T09:00:00Z",
  },
  {
    id: "reg-2",
    student_id: "s-1",
    exam_id: "e-2",
    token: "XYZ98765",
    card_pdf_url: null,
    checked_in_at: null,
    attempts_used: 0,
    status: "registered",
    created_at: "2026-06-02T00:00:00Z",
    exam_title: "Ujian Akhir Matematika",
    scheduled_at: null,
  },
];

describe("ExamPage", () => {
  beforeEach(() => {
    uiStore = { lang: "id" };
    registrationsState = {
      data: sample,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("renders the competition list with exam titles", async () => {
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText("Try Out UTBK Gratis #12")).toBeInTheDocument();
      expect(screen.getByText("Ujian Akhir Matematika")).toBeInTheDocument();
    });
  });

  it("translates copy when language is en", () => {
    uiStore = { lang: "en" };
    render(<ExamPage />);

    expect(
      screen.getByRole("heading", { name: /Competition & Tryout/i })
    ).toBeInTheDocument();
  });
});
