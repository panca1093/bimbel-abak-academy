import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { act } from "react";

import SessionPage from "./page";
import type { SessionState } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "session-1" }),
}));

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

// ── Mock hooks ────────────────────────────────────────────────────────────

let sessionState = {
  data: null as SessionState | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

const saveAnswersMutate = vi.fn();
const saveAnswersMutateAsync = vi.fn();
const submitSessionMutate = vi.fn();
const logViolationMutate = vi.fn();

vi.mock("@/lib/hooks/exam", () => ({
  useReconnectSession: () => sessionState,
  useSaveAnswers: () => ({
    mutate: saveAnswersMutate,
    mutateAsync: saveAnswersMutateAsync,
    isPending: false,
  }),
  useSubmitSession: () => ({
    mutate: submitSessionMutate,
    isPending: false,
  }),
  useLogViolation: () => ({
    mutate: logViolationMutate,
  }),
}));

// ── Sample data ───────────────────────────────────────────────────────────

const sampleSession: SessionState = {
  session_id: "session-1",
  registration_id: "reg-1",
  status: "in_progress",
  remaining_seconds: 3600,
  timer_mode: "overall",
  duration_minutes: 60,
  started_at: "2026-07-15T09:00:00Z",
  answers: [],
  tests: [
    {
      id: "test-1",
      title: "Tes Matematika",
      subject: "Matematika",
      questions: [
        {
          id: "q-mcq",
          test_id: "test-1",
          format: "mcq",
          body: "Berapa 2+2?",
          sort_order: 1,
          options: [
            { key: "A", text: "3", sort_order: 1 },
            { key: "B", text: "4", sort_order: 2 },
            { key: "C", text: "5", sort_order: 3 },
          ],
        },
        {
          id: "q-multi",
          test_id: "test-1",
          format: "multi_answer",
          body: "Pilih bilangan genap",
          sort_order: 2,
          options: [
            { key: "A", text: "1", sort_order: 1 },
            { key: "B", text: "2", sort_order: 2 },
            { key: "C", text: "4", sort_order: 3 },
          ],
        },
        {
          id: "q-short",
          test_id: "test-1",
          format: "short",
          body: "Ibu kota Indonesia adalah?",
          sort_order: 3,
          options: [],
        },
        {
          id: "q-fill",
          test_id: "test-1",
          format: "fill_blank",
          body: "Bendera Indonesia berwarna ___ dan putih.",
          sort_order: 4,
          options: [],
        },
        {
          id: "q-essay",
          test_id: "test-1",
          format: "essay",
          body: "Jelaskan penyebab Perang Diponegoro.",
          sort_order: 5,
          options: [],
        },
      ],
    },
  ],
};

const submittedSession: SessionState = {
  ...sampleSession,
  status: "submitted",
  submitted_at: "2026-07-15T10:00:00Z",
  remaining_seconds: 0,
};

// Helper: click fullscreen gate button to start the exam
async function enterFullscreen() {
  document.documentElement.requestFullscreen = vi
    .fn()
    .mockResolvedValue(undefined);
  const btn = screen.getByTestId("enter-fullscreen");
  fireEvent.click(btn);
  await waitFor(() => {
    expect(screen.getByText(/Berapa 2\+2\?/)).toBeInTheDocument();
  });
}

describe("SessionPage", () => {
  beforeEach(() => {
    uiStore = { lang: "id" };
    sessionState = {
      data: sampleSession,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    saveAnswersMutate.mockReset();
    saveAnswersMutateAsync.mockReset();
    submitSessionMutate.mockReset();
    logViolationMutate.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // ── Loading state ───────────────────────────────────────────────────────

  it("shows loading skeleton while reconnecting (FR29 reconnect)", () => {
    sessionState = { ...sessionState, data: null, isLoading: true };
    render(<SessionPage />);
    expect(screen.getByText("Memuat…")).toBeInTheDocument();
  });

  // ── Error state ─────────────────────────────────────────────────────────

  it("shows error card when reconnect fails (FR29 reconnect)", () => {
    sessionState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("not found"),
      refetch: vi.fn(),
    };
    render(<SessionPage />);
    expect(screen.getByText(/gagal memuat data/i)).toBeInTheDocument();
  });

  // ── Submitted state ─────────────────────────────────────────────────────

  it("shows submitted state when session is already submitted (FR29)", () => {
    sessionState = { ...sessionState, data: submittedSession };
    render(<SessionPage />);
    expect(screen.getByText("Terkumpul")).toBeInTheDocument();
  });

  // ── Fullscreen gate ─────────────────────────────────────────────────────

  it("shows fullscreen gate when not yet in fullscreen (FR29)", () => {
    render(<SessionPage />);
    expect(
      screen.getByText(/mode layar penuh diperlukan/i)
    ).toBeInTheDocument();
    expect(screen.getByTestId("enter-fullscreen")).toBeInTheDocument();
  });

  // ── Question rendering per format ───────────────────────────────────────

  it("renders MCQ question with radio inputs (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    // MCQ radio options
    const radios = screen.getAllByRole("radio");
    expect(radios).toHaveLength(3);
  });

  it("renders multi_answer with checkboxes (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    // Navigate to multi_answer question (index 1)
    fireEvent.click(screen.getByTestId("session-nav-1"));

    await waitFor(() => {
      expect(screen.getByText(/pilih bilangan genap/i)).toBeInTheDocument();
    });

    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes).toHaveLength(3);
  });

  it("renders short answer with text input (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    // Navigate to short answer (index 2)
    fireEvent.click(screen.getByTestId("session-nav-2"));

    await waitFor(() => {
      expect(
        screen.getByText(/ibu kota indonesia adalah/i)
      ).toBeInTheDocument();
    });

    // The text input should be visible
    const textInputs = screen
      .getAllByRole("textbox")
      .filter((tb) => tb.tagName === "INPUT");
    expect(textInputs.length).toBeGreaterThan(0);
  });

  it("renders essay with textarea (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    // Navigate to essay (index 4)
    fireEvent.click(screen.getByTestId("session-nav-4"));

    await waitFor(() => {
      expect(
        screen.getByText(/jelaskan penyebab perang diponegoro/i)
      ).toBeInTheDocument();
    });

    // Textarea should exist
    const textareas = screen
      .getAllByRole("textbox")
      .filter((tb) => tb.tagName === "TEXTAREA");
    expect(textareas.length).toBeGreaterThan(0);
  });

  // ── Flag toggle ─────────────────────────────────────────────────────────

  it("toggles flag for review (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    const flagBtn = screen.getByRole("button", { name: /tandai/i });
    fireEvent.click(flagBtn);

    expect(
      screen.getByRole("button", { name: /hapus tanda/i })
    ).toBeInTheDocument();
  });

  // ── Timer ───────────────────────────────────────────────────────────────

  it("shows countdown timer display (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    expect(screen.getByText(/60:00/)).toBeInTheDocument();
  });

  // ── Submit confirmation dialog ──────────────────────────────────────────

  it("shows submit confirmation dialog (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    const submitBtn = screen.getByRole("button", { name: /kumpulkan/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(
        screen.getByText(/yakin ingin mengumpulkan jawaban/i)
      ).toBeInTheDocument();
    });
  });

  // ── Answer updates state ────────────────────────────────────────────────

  it("updates answer state when MCQ option is selected (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    const radios = screen.getAllByRole("radio");
    expect(radios[0]).not.toBeChecked();
    expect(radios[1]).not.toBeChecked();

    fireEvent.click(radios[1]);

    expect(radios[0]).not.toBeChecked();
    expect(radios[1]).toBeChecked();
  });

  // ── Submit flow (also tests save is triggered) ──────────────────────────

  it("submit saves answers, calls hook, and shows result (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    // Answer a question first so save is triggered
    const radios = screen.getAllByRole("radio");
    fireEvent.click(radios[1]);

    // Open confirmation dialog
    fireEvent.click(screen.getByRole("button", { name: /kumpulkan/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/yakin ingin mengumpulkan jawaban/i)
      ).toBeInTheDocument();
    });

    // Click submit in dialog (last Kumpulkan button, inside the dialog)
    const btns = screen.getAllByRole("button", { name: /kumpulkan/i });
    fireEvent.click(btns[btns.length - 1]);

    // Verify save was triggered before submit
    expect(saveAnswersMutateAsync).toHaveBeenCalledWith(
      [{ question_id: "q-mcq", answer: "B" }],
    );

    // Verify submitSession was called
    expect(submitSessionMutate).toHaveBeenCalled();

    // Simulate success response inside act to flush React state updates
    await act(async () => {
      const [, opts] = submitSessionMutate.mock.calls[0];
      opts.onSuccess({ submitted: true, score: 75 });
    });

    // Submitted state with score
    expect(screen.getByText(/skor/i)).toBeInTheDocument();
    expect(screen.getByText(/75/)).toBeInTheDocument();
  });
});
