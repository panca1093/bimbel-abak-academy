import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import TestDetailPage from "./page";
import { useParams } from "next/navigation";
import type { TestDetail, QuestionWithOptions, Test } from "@/lib/types";

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

let saveState = { mutateAsync: mockMutateAsync, isPending: false };
let deleteState = { mutateAsync: mockMutateAsync, isPending: false };

vi.mock("@/lib/hooks/admin-tests", () => ({
  useTestDetail: () => testDetailState,
  useTestQuestions: () => questionsState,
  useSaveQuestion: () => saveState,
  useDeleteQuestion: () => deleteState,
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

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
      test_id: "test-1",
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
      test_id: "test-1",
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
    saveState = { mutateAsync: mockMutateAsync, isPending: false };
    deleteState = { mutateAsync: mockMutateAsync, isPending: false };
    mockMutateAsync.mockReset();
    mockMutateAsync.mockResolvedValue(undefined);
  });

  it("renders the test metadata header", async () => {
    render(<TestDetailPage />);

    await waitFor(() => {
      // title is the i18n page title
      expect(screen.getByRole("heading", { level: 1, name: /detail tes/i })).toBeInTheDocument();
    });
    // subtitle holds the test metadata: subject · topic · duration
    expect(screen.getByText(/Matematika/)).toBeInTheDocument();
    expect(screen.getByText(/Aljabar/)).toBeInTheDocument();
    expect(screen.getByText(/90/)).toBeInTheDocument();
  });

  it("renders the question list from useTestQuestions", async () => {
    render(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });
    expect(screen.getByText("Sebutkan 1+1")).toBeInTheDocument();
  });

  it("Add question button opens an inline editor in create mode", async () => {
    render(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /tambah soal/i }));

    // Inline editor should render a new QuestionEditor with empty body
    const bodyInputs = screen.getAllByLabelText(/badan soal/i);
    expect(bodyInputs.length).toBeGreaterThan(0);
    // The first (topmost) one is the new empty editor
    expect(bodyInputs[0]).toHaveValue("");
  });

  it("clicking a question's row expands its QuestionEditor", async () => {
    render(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    // The row header is a single toggle button wrapping the body text.
    const row = screen.getByText("Apa ibu kota Indonesia?").closest("[data-question-row]");
    expect(row).toBeTruthy();
    const toggleButton = within(row as HTMLElement).getByRole("button", {
      name: /ibu kota/i,
    });
    fireEvent.click(toggleButton);

    // After expanding, the QuestionEditor renders with the body prefilled.
    await waitFor(() => {
      const bodyInputs = screen.getAllByLabelText(/badan soal/i);
      const prefilled = bodyInputs.find(
        (el) => (el as HTMLTextAreaElement).value === "Apa ibu kota Indonesia?"
      );
      expect(prefilled).toBeTruthy();
    });
  });

  it("delete question calls useDeleteQuestion after confirm", async () => {
    vi.stubGlobal("confirm", () => true);
    mockMutateAsync.mockResolvedValueOnce(undefined);

    render(<TestDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("Apa ibu kota Indonesia?")).toBeInTheDocument();
    });

    const row = screen.getByText("Apa ibu kota Indonesia?").closest("[data-question-row]");
    expect(row).toBeTruthy();
    const deleteButton = within(row as HTMLElement).getByRole("button", { name: /hapus/i });
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith("q1");
    });

    vi.unstubAllGlobals();
  });
});
