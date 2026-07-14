import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import TestDetailPage from "./page";
import { useParams } from "next/navigation";
import type { TestDetail, QuestionWithOptions, Test, ExamTopic } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useParams: vi.fn(),
}));

const mockMutateAsync = vi.fn();

let testDetailState: {
  data: TestDetail | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
} = { data: undefined, isLoading: true, isError: false, error: null };

let questionsState: {
  data: { data: QuestionWithOptions[] } | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
} = { data: undefined, isLoading: true, isError: false, error: null };

let updateState = { mutateAsync: mockMutateAsync, isPending: false };
let detachState = { mutateAsync: mockMutateAsync, isPending: false };
let reorderState = { mutateAsync: mockMutateAsync, isPending: false };
let attachState = { mutateAsync: mockMutateAsync, isPending: false };
let saveQuestionState = { mutateAsync: mockMutateAsync, isPending: false };

const mockTopics: ExamTopic[] = [{ id: "topic-1", name: "Aljabar", subject: "Matematika" }];
const mockBankQuestions: { data: QuestionWithOptions[] } = {
  data: [
    {
      question: {
        id: "bq1",
        format: "mcq",
        body: "Bank question one",
        sort_order: 1,
        point_correct: 1,
        point_wrong: 0,
      },
      options: [],
    },
  ],
};

vi.mock("@/lib/hooks/admin-tests", () => ({
  useTestDetail: () => testDetailState,
  useTestQuestions: () => questionsState,
  useUpdateTest: () => updateState,
  useDetachQuestion: () => detachState,
  useReorderTestQuestions: () => reorderState,
  useAttachQuestions: () => attachState,
  useSaveQuestion: () => saveQuestionState,
}));

vi.mock("@/lib/hooks/admin-topics", () => ({
  useTopics: () => ({ data: { data: mockTopics }, isLoading: false }),
}));

vi.mock("@/lib/hooks/admin-bank-questions", () => ({
  useBankQuestions: () => ({ data: mockBankQuestions, isLoading: false, isError: false }),
  useCreateBankQuestion: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateBankQuestion: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

function renderWithClient(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

const sampleTest: Test = {
  id: "test-1",
  title: "Tryout UTBK Saintek",
  subject: "Matematika",
  topic: "Aljabar",
  duration_minutes: 90,
};

const sampleQuestions: QuestionWithOptions[] = [
  {
    question: {
      id: "q1",
      format: "mcq",
      body: "Apa ibu kota Indonesia?",
      sort_order: 1,
      point_correct: 1,
      point_wrong: 0,
    },
    options: [
      { question_id: "q1", key: "a", text: "Jakarta", is_correct: true, sort_order: 1 },
      { question_id: "q1", key: "b", text: "Bandung", is_correct: false, sort_order: 2 },
    ],
  },
  {
    question: {
      id: "q2",
      format: "short",
      body: "Sebutkan 1+1",
      sort_order: 2,
      correct_answer: "2",
      point_correct: 1,
      point_wrong: 0,
    },
    options: [],
  },
];

describe("TestDetailPage", () => {
  beforeEach(() => {
    (useParams as ReturnType<typeof vi.fn>).mockReturnValue({ id: "test-1" });
    testDetailState = {
      data: { test: sampleTest, questions: sampleQuestions },
      isLoading: false,
      isError: false,
      error: null,
    };
    questionsState = {
      data: { data: sampleQuestions },
      isLoading: false,
      isError: false,
      error: null,
    };
    updateState = { mutateAsync: mockMutateAsync, isPending: false };
    detachState = { mutateAsync: mockMutateAsync, isPending: false };
    reorderState = { mutateAsync: mockMutateAsync, isPending: false };
    attachState = { mutateAsync: mockMutateAsync, isPending: false };
    saveQuestionState = { mutateAsync: mockMutateAsync, isPending: false };
    mockMutateAsync.mockReset();
    mockMutateAsync.mockResolvedValue(undefined);
  });

  it("renders the test metadata header", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByRole("heading", { level: 1, name: /detail tes/i })).toBeInTheDocument();
    });
    expect(screen.getByText(/Matematika/)).toBeInTheDocument();
    expect(screen.getByText(/Aljabar/)).toBeInTheDocument();
    expect(screen.getByText(/90/)).toBeInTheDocument();
  });

  it("renders two columns with test details form and questions panel", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByLabelText(/judul/i)).toBeInTheDocument();
    });
    expect(screen.getByLabelText(/durasi/i)).toBeInTheDocument();
    expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    expect(screen.getByText("Sebutkan 1+1")).toBeInTheDocument();
  });

  it("saves test metadata via useUpdateTest", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByLabelText(/judul/i)).toHaveValue("Tryout UTBK Saintek");
    });

    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Geometri" } });
    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ topic: "Geometri" })
      );
    });
  });

  it("New question button opens an inline QuestionEditor", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /soal baru/i }));

    const bodyInputs = screen.getAllByLabelText(/badan soal/i);
    expect(bodyInputs.length).toBeGreaterThan(0);
    expect(bodyInputs[0]).toHaveValue("");
  });

  it("detach button calls useDetachQuestion after confirm", async () => {
    vi.stubGlobal("confirm", () => true);
    mockMutateAsync.mockResolvedValueOnce(undefined);

    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    const row = screen.getByText("Apa ibu kota Indonesia?").closest("[data-question-row]");
    expect(row).toBeTruthy();
    const detachButton = within(row as HTMLElement).getByRole("button", { name: /lepas/i });
    fireEvent.click(detachButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith("q1");
    });

    vi.unstubAllGlobals();
  });

  it("reorder down persists the new full question_ids order", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    const row = screen.getByText("Apa ibu kota Indonesia?").closest("[data-question-row]");
    const downButton = within(row as HTMLElement).getByRole("button", { name: /turun/i });
    fireEvent.click(downButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        question_ids: ["q2", "q1"],
      });
    });
  });

  it("From bank opens the picker and attaches selected questions", async () => {
    renderWithClient(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /dari bank/i }));

    await waitFor(() => {
      expect(screen.getByRole("dialog", { name: /pilih soal dari bank/i })).toBeInTheDocument();
    });

    const row = screen.getByText("Bank question one").closest("label");
    fireEvent.click(within(row as HTMLElement).getByRole("checkbox"));

    fireEvent.click(screen.getByRole("button", { name: /tambahkan 1 soal/i }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({ question_ids: ["bq1"] });
    });
  });
});
