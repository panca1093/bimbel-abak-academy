import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { act } from "react";

import SessionPage from "./page";
import type { SessionState } from "@/lib/types";

const routerReplace = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "session-1" }),
  useRouter: () => ({ replace: routerReplace }),
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
const advanceSectionMutate = vi.fn();
const advanceSectionMutateAsync = vi.fn();

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
  useAdvanceSection: () => ({
    mutate: advanceSectionMutate,
    mutateAsync: advanceSectionMutateAsync,
    isPending: false,
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

// ── Sectioned session samples ───────────────────────────────────────────────

const sectionedSession: SessionState = {
  ...sampleSession,
  mode: "utbk",
  active_test_id: "test-section-1",
  duration_minutes: null,
  remaining_seconds: 0,
  answers: [],
  tests: [
    {
      id: "test-section-1",
      title: "TPS",
      subject: "TPS",
      status: "active",
      remaining_seconds: 1800,
      duration_minutes: 30,
      questions: [
        {
          id: "q-sec1-mcq",
          test_id: "test-section-1",
          format: "mcq",
          body: "TPS Question 1?",
          sort_order: 1,
          options: [
            { key: "A", text: "Three", sort_order: 1 },
            { key: "B", text: "Four", sort_order: 2 },
          ],
        },
        {
          id: "q-sec1-essay",
          test_id: "test-section-1",
          format: "essay",
          body: "TPS Essay?",
          sort_order: 2,
          options: [],
        },
      ],
    },
    {
      id: "test-section-2",
      title: "Literasi",
      subject: "Literasi",
      status: "pending",
      remaining_seconds: 0,
      duration_minutes: 45,
      questions: [
        {
          id: "q-sec2-mcq",
          test_id: "test-section-2",
          format: "mcq",
          body: "Literasi Question 1?",
          sort_order: 1,
          options: [
            { key: "A", text: "Choice A", sort_order: 1 },
            { key: "B", text: "Choice B", sort_order: 2 },
          ],
        },
      ],
    },
  ],
};

const ieltsSession: SessionState = {
  ...sectionedSession,
  mode: "ielts",
  active_test_id: "test-listening",
  duration_minutes: null,
  tests: [
    {
      id: "test-listening",
      title: "Listening",
      subject: "Listening",
      section_type: "listening",
      status: "active",
      remaining_seconds: 2400,
      duration_minutes: 40,
      audio_url: "https://example.com/audio.mp3",
      audio_play_limit: 2,
      questions: [
        {
          id: "q-listening",
          test_id: "test-listening",
          format: "mcq",
          body: "Listening Q1?",
          sort_order: 1,
          options: [
            { key: "A", text: "Opt A", sort_order: 1 },
            { key: "B", text: "Opt B", sort_order: 2 },
          ],
        },
      ],
    },
    {
      id: "test-reading",
      title: "Reading",
      subject: "Reading",
      section_type: "reading",
      status: "pending",
      remaining_seconds: 0,
      duration_minutes: 60,
      questions: [],
    },
    {
      id: "test-writing",
      title: "Writing",
      subject: "Writing",
      section_type: "writing",
      status: "pending",
      remaining_seconds: 0,
      duration_minutes: 60,
      questions: [
        {
          id: "q-writing",
          test_id: "test-writing",
          format: "essay",
          body: "Writing Task?",
          sort_order: 1,
          options: [],
        },
      ],
    },
  ],
};

// Helper: click fullscreen gate button to start the exam (standard mode)
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

// Helper: enter fullscreen for sectioned exam (UTBK)
async function enterFullscreenSectioned() {
  document.documentElement.requestFullscreen = vi
    .fn()
    .mockResolvedValue(undefined);
  const btn = screen.getByTestId("enter-fullscreen");
  fireEvent.click(btn);
  await waitFor(() => {
    expect(screen.getByText(/TPS Question 1\?/)).toBeInTheDocument();
  });
}

// Helper: enter fullscreen for IELTS (listening section)
async function enterFullscreenIELTS() {
  document.documentElement.requestFullscreen = vi
    .fn()
    .mockResolvedValue(undefined);
  const btn = screen.getByTestId("enter-fullscreen");
  fireEvent.click(btn);
  await waitFor(() => {
    expect(screen.getByText(/Listening Q1\?/)).toBeInTheDocument();
  });
}

// Helper: enter fullscreen and wait for a specific question text
async function enterFullscreenUntil(text: RegExp) {
  document.documentElement.requestFullscreen = vi
    .fn()
    .mockResolvedValue(undefined);
  const btn = screen.getByTestId("enter-fullscreen");
  fireEvent.click(btn);
  await waitFor(() => {
    expect(screen.getByText(text)).toBeInTheDocument();
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
    advanceSectionMutate.mockReset();
    advanceSectionMutateAsync.mockReset();
    routerReplace.mockReset();
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

  it("redirects to the result route when session is already submitted (FR29, FR-S5-25)", () => {
    sessionState = { ...sessionState, data: submittedSession };
    render(<SessionPage />);
    expect(routerReplace).toHaveBeenCalledWith(
      "/exam/sessions/session-1/result",
    );
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

  it("renders rich body via RichContent (LaTeX + bold HTML) on the question card", async () => {
    sessionState = {
      ...sessionState,
      data: {
        ...sampleSession,
        tests: [
          {
            ...sampleSession.tests[0],
            questions: [
              {
                id: "q-rich",
                test_id: "test-1",
                format: "mcq",
                body: "Hitung \\(x^2\\) dan buat <b>tebal</b>",
                sort_order: 1,
                options: [
                  { key: "A", text: "Ya", sort_order: 1 },
                  { key: "B", text: "Tidak", sort_order: 2 },
                ],
              },
            ],
          },
        ],
      },
    };
    render(<SessionPage />);
    document.documentElement.requestFullscreen = vi
      .fn()
      .mockResolvedValue(undefined);
    fireEvent.click(screen.getByTestId("enter-fullscreen"));

    // Body should be wrapped in RichContent; KaTeX renders \(x^2\) and <b> renders bold.
    const richNode = await waitFor(
      () => {
        const el = document.querySelector("[data-rich-content] .katex");
        if (!el) throw new Error("not yet");
        return el.closest("[data-rich-content]") as HTMLElement;
      },
      { timeout: 3000 }
    );
    expect(richNode).not.toBeNull();
    const b = richNode.querySelector("b");
    expect(b).not.toBeNull();
    expect(b?.textContent).toBe("tebal");
    // Literal LaTeX delimiters are replaced by KaTeX — not visible as text.
    expect(richNode.textContent).not.toContain("\\(");
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

  it("rehydrates flagged_for_review from session answers on reconnect (FR29)", async () => {
    sessionState = {
      ...sessionState,
      data: {
        ...sampleSession,
        answers: [
          { question_id: "q-mcq", answer: "B", flagged_for_review: true },
        ],
      },
    };
    render(<SessionPage />);
    await enterFullscreen();

    expect(
      screen.getByRole("button", { name: /hapus tanda/i })
    ).toBeInTheDocument();
  });

  it("includes flagged_for_review in the submit save payload (FR29)", async () => {
    saveAnswersMutateAsync.mockResolvedValue(undefined);
    render(<SessionPage />);
    await enterFullscreen();

    const flagBtn = screen.getByRole("button", { name: /tandai/i });
    fireEvent.click(flagBtn);

    fireEvent.click(screen.getByRole("button", { name: /kumpulkan/i }));
    await waitFor(() => {
      expect(
        screen.getByText(/yakin ingin mengumpulkan jawaban/i)
      ).toBeInTheDocument();
    });
    const btns = screen.getAllByRole("button", { name: /kumpulkan/i });
    fireEvent.click(btns[btns.length - 1]);

    await waitFor(() => {
      expect(saveAnswersMutateAsync).toHaveBeenCalled();
    });
    const payload = saveAnswersMutateAsync.mock.calls[0][0];
    expect(payload).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          question_id: "q-mcq",
          flagged_for_review: true,
        }),
      ])
    );
  });

  // ── Timer ───────────────────────────────────────────────────────────────

  it("shows countdown timer display (FR29)", async () => {
    render(<SessionPage />);
    await enterFullscreen();

    expect(screen.getByText(/60:00/)).toBeInTheDocument();
  });

  it("untimed exam (per_test, null duration) never auto-submits and hides the countdown", async () => {
    sessionState = {
      ...sessionState,
      data: {
        ...sampleSession,
        timer_mode: "per_test",
        duration_minutes: null,
        remaining_seconds: 0,
      },
    };
    render(<SessionPage />);
    await enterFullscreen();

    expect(screen.getByText(/Berapa 2\+2\?/)).toBeInTheDocument();
    expect(screen.queryByText(/00:00/)).not.toBeInTheDocument();
    await waitFor(() => {
      expect(submitSessionMutate).not.toHaveBeenCalled();
    });
    expect(routerReplace).not.toHaveBeenCalled();
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

  it("submit saves answers, calls hook, and redirects to result (FR29, FR-S5-25)", async () => {
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
      [{ question_id: "q-mcq", answer: "B", flagged_for_review: false }],
    );

    // Verify submitSession was called
    expect(submitSessionMutate).toHaveBeenCalled();

    // Simulate success response inside act to flush React state updates
    await act(async () => {
      const [, opts] = submitSessionMutate.mock.calls[0];
      opts.onSuccess({ submitted: true, score: 75 });
    });

    // Redirects to the result route instead of rendering an inline card
    expect(routerReplace).toHaveBeenCalledWith(
      "/exam/sessions/session-1/result",
    );
  });

  // ── Sectioned mode (FR-23) ────────────────────────────────────────────

  it("sectioned mode shows only active section questions (FR-23)", async () => {
    sessionState = { ...sessionState, data: sectionedSession };
    render(<SessionPage />);
    await enterFullscreenSectioned();

    // Active section's first question is visible
    expect(screen.getByText(/TPS Question 1\?/)).toBeInTheDocument();

    // Navigate to second question in the same section
    fireEvent.click(screen.getByTestId("session-nav-1"));
    await waitFor(() => {
      expect(screen.getByText(/TPS Essay\?/)).toBeInTheDocument();
    });

    // Non-active section questions are NOT visible anywhere
    expect(screen.queryByText(/Literasi Question 1\?/)).not.toBeInTheDocument();
  });

  it("sectioned mode renders section rail with all sections (FR-23)", async () => {
    sessionState = { ...sessionState, data: sectionedSession };
    render(<SessionPage />);
    await enterFullscreenSectioned();

    // Section rail container exists
    expect(screen.getByTestId("section-rail")).toBeInTheDocument();

    // Each section title appears in the rail
    const rail = screen.getByTestId("section-rail");
    expect(rail).toHaveTextContent("TPS");
    expect(rail).toHaveTextContent("Literasi");
  });

  it("sectioned mode shows per-section countdown (FR-23)", async () => {
    sessionState = { ...sessionState, data: sectionedSession };
    render(<SessionPage />);
    await enterFullscreenSectioned();

    // Active section has 1800 seconds remaining = 30:00
    expect(screen.getByText(/30:00/)).toBeInTheDocument();
  });

  it("timer zero in sectioned mode calls save then advance (FR-24)", async () => {
    saveAnswersMutateAsync.mockResolvedValue(undefined);
    advanceSectionMutateAsync.mockResolvedValue({
      mode: "utbk",
      active_test_id: "test-section-2",
      completed: false,
      tests: sectionedSession.tests,
    });

    // Set remaining to 0 so the auto-advance fires immediately
    // Include a pre-existing answer so buildSavePayload returns a non-empty array
    const expiredSession = {
      ...sectionedSession,
      answers: [{ question_id: "q-sec1-mcq", answer: "A" }],
      tests: sectionedSession.tests.map((t, i) =>
        i === 0 ? { ...t, remaining_seconds: 0 } : t,
      ),
    };
    sessionState = { ...sessionState, data: expiredSession };
    render(<SessionPage />);
    await enterFullscreenSectioned();

    await waitFor(() => {
      expect(saveAnswersMutateAsync).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(advanceSectionMutateAsync).toHaveBeenCalledWith("test-section-1");
    });

    // Submit should NOT be called for a non-last section advance
    expect(submitSessionMutate).not.toHaveBeenCalled();
  });

  it("advancing last section triggers submit and redirect (FR-24)", async () => {
    saveAnswersMutateAsync.mockResolvedValue(undefined);
    advanceSectionMutateAsync.mockResolvedValue({
      mode: "utbk",
      active_test_id: null,
      completed: true,
      tests: sectionedSession.tests,
    });

    // Active = last section (test-section-2), remaining=0
    const lastSectionActive = {
      ...sectionedSession,
      active_test_id: "test-section-2",
      tests: [
        {
          ...sectionedSession.tests[0],
          status: "submitted" as const,
          remaining_seconds: 0,
        },
        {
          ...sectionedSession.tests[1],
          status: "active" as const,
          remaining_seconds: 0,
        },
      ],
    };
    sessionState = { ...sessionState, data: lastSectionActive };
    render(<SessionPage />);
    await enterFullscreenUntil(/Literasi Question 1\?/);

    await waitFor(() => {
      expect(saveAnswersMutateAsync).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(advanceSectionMutateAsync).toHaveBeenCalledWith("test-section-2");
    });
    await waitFor(() => {
      expect(submitSessionMutate).toHaveBeenCalled();
    });

    // Simulate submit success
    await act(async () => {
      const [, opts] = submitSessionMutate.mock.calls[0];
      opts.onSuccess({ submitted: true, score: 85 });
    });
    expect(routerReplace).toHaveBeenCalledWith(
      "/exam/sessions/session-1/result",
    );
  });

  it("periodic save excludes a submitted section's answers (FR-14 seam — backend rejects locked-section saves)", async () => {
    vi.useFakeTimers();

    // Section 1 is already submitted; section 2 is active (not expired). Both
    // sections carry a persisted answer, rehydrated into state on reconnect.
    // The autosave must send ONLY the active section's answer — the backend
    // rejects the whole batch (ErrSectionLocked) if any answer targets a
    // non-active section, which would silently lose every section past the first.
    const section2Active = {
      ...sectionedSession,
      active_test_id: "test-section-2",
      answers: [
        { question_id: "q-sec1-mcq", answer: "A" },
        { question_id: "q-sec2-mcq", answer: "B" },
      ],
      tests: [
        {
          ...sectionedSession.tests[0],
          status: "submitted" as const,
          remaining_seconds: 0,
        },
        {
          ...sectionedSession.tests[1],
          status: "active" as const,
          remaining_seconds: 1800,
        },
      ],
    };
    sessionState = { ...sessionState, data: section2Active };
    render(<SessionPage />);
    // Flush the init effect so answers hydrate before the autosave fires.
    await act(async () => {
      await Promise.resolve();
    });
    // Fire the 30s periodic autosave.
    await act(async () => {
      vi.advanceTimersByTime(30000);
    });

    expect(saveAnswersMutate).toHaveBeenCalled();
    const sentIds = saveAnswersMutate.mock.calls.flatMap(([payload]) =>
      (payload as Array<{ question_id: string }>).map((p) => p.question_id),
    );
    expect(sentIds).toContain("q-sec2-mcq"); // active section answer is saved
    expect(sentIds).not.toContain("q-sec1-mcq"); // submitted section answer is not resent
  });

  it("resets the question index to 0 when advancing to a shorter section (FR-13)", async () => {
    sessionState = { ...sessionState, data: sectionedSession };
    const { rerender } = render(<SessionPage />);
    await enterFullscreenSectioned();

    // Move to section 1's 2nd question (index 1).
    fireEvent.click(screen.getByTestId("session-nav-1"));
    await waitFor(() => {
      expect(screen.getByText(/TPS Essay\?/)).toBeInTheDocument();
    });

    // Advance to section 2 (Literasi) which has only ONE question. If the index
    // is not reset, questionsToShow[1] is undefined and the panel renders blank.
    const section2Active = {
      ...sectionedSession,
      active_test_id: "test-section-2",
      tests: [
        { ...sectionedSession.tests[0], status: "submitted" as const },
        {
          ...sectionedSession.tests[1],
          status: "active" as const,
          remaining_seconds: 2700,
        },
      ],
    };
    sessionState = { ...sessionState, data: section2Active };
    rerender(<SessionPage />);

    await waitFor(() => {
      expect(screen.getByText(/Literasi Question 1\?/)).toBeInTheDocument();
    });
  });

  it("pending section rail items are not clickable (FR-23)", async () => {
    sessionState = { ...sessionState, data: sectionedSession };
    render(<SessionPage />);
    await enterFullscreenSectioned();

    // Active section is the first one (TPS) — clicking Literasi rail item
    // should not change the visible questions
    const literasiRail = screen.getByText("Literasi");
    fireEvent.click(literasiRail);

    // Active section questions still shown (no navigation)
    expect(screen.getByText(/TPS Question 1\?/)).toBeInTheDocument();
    expect(screen.queryByText(/Literasi Question 1\?/)).not.toBeInTheDocument();
  });

  // ── IELTS skill rendering (FR-25) ──────────────────────────────────────

  it("renders audio player for listening sections (FR-25)", async () => {
    sessionState = { ...sessionState, data: ieltsSession };
    render(<SessionPage />);
    await enterFullscreenIELTS();

    const audio = screen.getByTestId("section-audio-player");
    expect(audio).toBeInTheDocument();
    expect(audio).toHaveAttribute("src", "https://example.com/audio.mp3");
  });

  it("renders writing section questions as essay (FR-25)", async () => {
    // Set writing as active section
    const writingActive = {
      ...ieltsSession,
      active_test_id: "test-writing",
      tests: ieltsSession.tests.map((t) => {
        if (t.id === "test-listening")
          return { ...t, status: "submitted" as const, remaining_seconds: 0 };
        if (t.id === "test-writing")
          return { ...t, status: "active" as const, remaining_seconds: 3600, duration_minutes: 60 };
        return t;
      }),
    };
    sessionState = { ...sessionState, data: writingActive };
    render(<SessionPage />);
    await enterFullscreenUntil(/Writing Task\?/);

    // Writing section uses essay format (textarea)
    expect(screen.getByText(/Writing Task\?/)).toBeInTheDocument();
    const textareas = screen
      .getAllByRole("textbox")
      .filter((tb) => tb.tagName === "TEXTAREA");
    expect(textareas.length).toBeGreaterThan(0);
  });
});
