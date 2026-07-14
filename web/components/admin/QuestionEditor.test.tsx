import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent, within } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { QuestionEditor } from "./QuestionEditor";
import type { QuestionWithOptions, QuestionFormat, Question, QuestionOption, ExamTopic } from "@/lib/types";

const mockTestSaveAsync = vi.fn();
let testSaveState = { mutateAsync: mockTestSaveAsync, isPending: false };

const mockCreateBankAsync = vi.fn();
let createBankState = { mutateAsync: mockCreateBankAsync, isPending: false };

const mockUpdateBankAsync = vi.fn();
let updateBankState = { mutateAsync: mockUpdateBankAsync, isPending: false };

const mockTopics: ExamTopic[] = [
  { id: "topic-1", name: "Aljabar", subject: "Matematika" },
  { id: "topic-2", name: "Fisika Dasar", subject: "Fisika" },
];

vi.mock("@/lib/hooks/admin-tests", () => ({
  useSaveQuestion: () => testSaveState,
}));

vi.mock("@/lib/hooks/admin-bank-questions", () => ({
  useCreateBankQuestion: () => createBankState,
  useUpdateBankQuestion: () => updateBankState,
}));

vi.mock("@/lib/hooks/admin-topics", () => ({
  useTopics: () => ({ data: { data: mockTopics }, isLoading: false }),
}));

function renderWithClient(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

function makeQuestion(overrides: Partial<Question> = {}): Question {
  return {
    id: "q1",
    format: "mcq" as QuestionFormat,
    body: "Apa ibu kota Indonesia?",
    sort_order: 1,
    point_correct: 1,
    point_wrong: 0,
    topic_id: "topic-1",
    topic: "Aljabar",
    ...overrides,
  };
}

function makeOption(overrides: Partial<QuestionOption> = {}): QuestionOption {
  return {
    question_id: "q1",
    key: "a",
    text: "Jakarta",
    is_correct: true,
    sort_order: 1,
    ...overrides,
  };
}

function makeQuestionWithOptions(
  q?: Partial<Question>,
  opts?: QuestionOption[]
): QuestionWithOptions {
  return {
    question: makeQuestion(q),
    options: opts ?? [
      makeOption({ key: "a", text: "Jakarta", is_correct: true, sort_order: 1 }),
      makeOption({ question_id: "q1", key: "b", text: "Bandung", is_correct: false, sort_order: 2 }),
    ],
  };
}

function fillRequiredFields() {
  fireEvent.input(screen.getByLabelText(/badan soal/i), { target: { value: "Soal" } });
  fireEvent.change(screen.getByLabelText(/topik/i), { target: { value: "topic-1" } });
}

describe("QuestionEditor", () => {
  beforeEach(() => {
    mockTestSaveAsync.mockReset();
    mockTestSaveAsync.mockResolvedValue({ question: makeQuestion(), options: [] });
    testSaveState = { mutateAsync: mockTestSaveAsync, isPending: false };

    mockCreateBankAsync.mockReset();
    mockCreateBankAsync.mockResolvedValue({ question: makeQuestion(), options: [] });
    createBankState = { mutateAsync: mockCreateBankAsync, isPending: false };

    mockUpdateBankAsync.mockReset();
    mockUpdateBankAsync.mockResolvedValue({ question: makeQuestion(), options: [] });
    updateBankState = { mutateAsync: mockUpdateBankAsync, isPending: false };
  });

  it("renders create mode with format defaulting to mcq", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    expect(screen.getByLabelText(/badan soal/i)).toHaveValue("");
    const radios = screen.getAllByRole("radio");
    expect(radios.length).toBe(2);
  });

  it("renders edit mode with existing mcq options prefilled and correct radio set", () => {
    const qwo = makeQuestionWithOptions();
    renderWithClient(
      <QuestionEditor testId="test-1" question={qwo} onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    expect(screen.getByDisplayValue("Apa ibu kota Indonesia?")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Jakarta")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Bandung")).toBeInTheDocument();

    const radios = screen.getAllByRole("radio");
    const checked = radios.filter((r) => (r as HTMLInputElement).checked);
    expect(checked.length).toBe(1);
  });

  it("switching format to essay hides option editor and correct_answer input", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    expect(screen.getAllByRole("radio").length).toBeGreaterThan(0);

    const formatSelect = screen.getByLabelText(/format/i);
    fireEvent.change(formatSelect, { target: { value: "essay" } });

    expect(screen.queryAllByRole("radio").length).toBe(0);
    expect(screen.queryByLabelText(/jawaban benar/i)).not.toBeInTheDocument();
  });

  it("switching format to short shows correct_answer input and hides option editor", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    const formatSelect = screen.getByLabelText(/format/i);
    fireEvent.change(formatSelect, { target: { value: "short" } });

    expect(screen.getByLabelText(/jawaban benar/i)).toBeInTheDocument();
    expect(screen.queryAllByRole("radio").length).toBe(0);
  });

  it("switching format to fill_blank shows correct_answer input and hides option editor", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    const formatSelect = screen.getByLabelText(/format/i);
    fireEvent.change(formatSelect, { target: { value: "fill_blank" } });

    expect(screen.getByLabelText(/jawaban benar/i)).toBeInTheDocument();
    expect(screen.queryAllByRole("radio").length).toBe(0);
  });

  it("switching format to multi_answer shows checkboxes instead of radios", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    const formatSelect = screen.getByLabelText(/format/i);
    fireEvent.change(formatSelect, { target: { value: "multi_answer" } });

    expect(screen.queryAllByRole("radio").length).toBe(0);
    expect(screen.getAllByRole("checkbox").length).toBeGreaterThan(0);
  });

  it("submit calls save mutation with correct input shape (mcq)", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fireEvent.input(screen.getByLabelText(/badan soal/i), {
      target: { value: "Soal baru" },
    });
    fireEvent.input(screen.getByLabelText(/urutan/i), { target: { value: "2" } });
    fireEvent.change(screen.getByLabelText(/topik/i), { target: { value: "topic-1" } });

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          input: expect.objectContaining({
            format: "mcq",
            body: "Soal baru",
            sort_order: 2,
            topic_id: "topic-1",
            options: expect.any(Array),
          }),
        })
      );
    });
  });

  it("mcq submit with default 1-correct option passes validation", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalled();
    });
  });

  it("mcq submit with all options moved to a different one still passes (1 correct)", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();

    const radios = screen.getAllByRole("radio");
    fireEvent.change(radios[1], { target: { checked: true } });

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalled();
    });
  });

  it("multi_answer validation: 0 correct blocks submit", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.change(screen.getByLabelText(/format/i), { target: { value: "multi_answer" } });

    const checkboxes = screen.getAllByRole("checkbox");
    fireEvent.click(checkboxes[0]);

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/minimal satu opsi benar/i)
      ).toBeInTheDocument();
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
  });

  it("multi_answer validation: 1 correct allowed", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.change(screen.getByLabelText(/format/i), { target: { value: "multi_answer" } });

    const checkboxes = screen.getAllByRole("checkbox");
    fireEvent.click(checkboxes[1]);

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalled();
    });
  });

  it("short validation: empty correct_answer blocks submit", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.change(screen.getByLabelText(/format/i), { target: { value: "short" } });

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/jawaban benar wajib diisi/i)
      ).toBeInTheDocument();
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
  });

  it("empty body blocks submit with validation error", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fireEvent.change(screen.getByLabelText(/topik/i), { target: { value: "topic-1" } });
    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/badan soal wajib diisi/i)
      ).toBeInTheDocument();
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
  });

  it("edit mode includes question id in save payload", async () => {
    const qwo = makeQuestionWithOptions();
    renderWithClient(
      <QuestionEditor testId="test-1" question={qwo} onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalledWith(
        expect.objectContaining({ question: "q1" })
      );
    });
  });

  // ── Penilaian panel (FR-S5-03, FR-S5-29) ─────────────────────────────────

  it("renders the Penilaian panel with correct min/step attributes", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    expect(screen.getByText(/^penilaian$/i)).toBeInTheDocument();

    const pointCorrect = screen.getByLabelText(/poin benar/i);
    expect(pointCorrect).toHaveAttribute("min", "1");
    expect(pointCorrect).toHaveAttribute("step", "1");
    expect(pointCorrect).toHaveValue(1);

    const pointWrong = screen.getByLabelText(/poin salah/i);
    expect(pointWrong).toHaveAttribute("min", "0");
    expect(pointWrong).toHaveAttribute("step", "1");
    expect(pointWrong).toHaveValue(0);
  });

  it("save payload carries both point_correct and point_wrong", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.input(screen.getByLabelText(/poin benar/i), { target: { value: "4" } });
    fireEvent.input(screen.getByLabelText(/poin salah/i), { target: { value: "2" } });

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          input: expect.objectContaining({ point_correct: 4, point_wrong: 2 }),
        })
      );
    });
  });

  it("edit mode initializes points from question.point_correct/point_wrong, not the difficulty default", () => {
    const qwo = makeQuestionWithOptions({ difficulty: "easy", point_correct: 7, point_wrong: 3 });
    renderWithClient(
      <QuestionEditor testId="test-1" question={qwo} onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    expect(screen.getByLabelText(/poin benar/i)).toHaveValue(7);
    expect(screen.getByLabelText(/poin salah/i)).toHaveValue(3);
  });

  // ── Topic select (FR-34..FR-36) ─────────────────────────────────────────

  it("renders topic select populated from useTopics", () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    const topicSelect = screen.getByLabelText(/topik/i);
    expect(topicSelect).toBeInTheDocument();
    expect(within(topicSelect).getByText("Aljabar")).toBeInTheDocument();
    expect(within(topicSelect).getByText("Fisika Dasar")).toBeInTheDocument();
  });

  it("topic is required and blocks submit when empty", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fireEvent.input(screen.getByLabelText(/badan soal/i), { target: { value: "Soal" } });
    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(screen.getByText(/topik wajib dipilih/i)).toBeInTheDocument();
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
  });

  it("bank standalone create uses useCreateBankQuestion", async () => {
    renderWithClient(
      <QuestionEditor onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockCreateBankAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          format: "mcq",
          body: "Soal",
          topic_id: "topic-1",
          options: expect.any(Array),
        })
      );
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
    expect(mockUpdateBankAsync).not.toHaveBeenCalled();
  });

  it("bank standalone edit uses useUpdateBankQuestion", async () => {
    const qwo = makeQuestionWithOptions();
    renderWithClient(
      <QuestionEditor question={qwo} onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockUpdateBankAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          format: "mcq",
          body: "Apa ibu kota Indonesia?",
          topic_id: "topic-1",
        })
      );
    });
    expect(mockTestSaveAsync).not.toHaveBeenCalled();
    expect(mockCreateBankAsync).not.toHaveBeenCalled();
  });

  it("bank standalone create omits sort_order in payload", async () => {
    renderWithClient(
      <QuestionEditor onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockCreateBankAsync).toHaveBeenCalledWith(
        expect.not.objectContaining({ sort_order: expect.any(Number) })
      );
    });
  });

  it("test scoped new question hits create-and-attach via useSaveQuestion", async () => {
    renderWithClient(
      <QuestionEditor testId="test-1" onCancel={vi.fn()} onSaved={vi.fn()} />
    );

    fillRequiredFields();
    fireEvent.click(screen.getByRole("button", { name: /simpan soal/i }));

    await waitFor(() => {
      expect(mockTestSaveAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          question: undefined,
          input: expect.objectContaining({ topic_id: "topic-1" }),
        })
      );
    });
  });
});
