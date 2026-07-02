import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import { useParams } from "next/navigation";
import ExamPackageDetailPage from "./page";
import type { ExamDetail, GradingSessionItem, GradingEssayItem, Test } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useParams: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const mockReplaceTests = vi.fn();
const mockUpdatePrice = vi.fn();
const mockPublish = vi.fn();
const mockGradeEssay = vi.fn();

let examState: {
  data: ExamDetail | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: ReturnType<typeof vi.fn>;
} = { data: undefined, isLoading: true, isError: false, error: null, refetch: vi.fn() };

let gradingSessionsState: {
  data: { data: GradingSessionItem[] } | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
} = { data: undefined, isLoading: true, isError: false, error: null };

let sessionEssaysState: {
  data: { data: GradingEssayItem[] } | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
} = { data: undefined, isLoading: false, isError: false, error: null };

let gradeEssayState: {
  mutateAsync: ReturnType<typeof vi.fn>;
  isPending: boolean;
  variables?: { question_id: string };
} = { mutateAsync: mockGradeEssay, isPending: false, variables: undefined };

vi.mock("@/lib/hooks/admin-exams", () => ({
  useExam: () => examState,
  useReplaceExamTests: () => ({ mutateAsync: mockReplaceTests, isPending: false }),
  useUpdateExamPrice: () => ({ mutateAsync: mockUpdatePrice, isPending: false }),
  usePublishExam: () => ({ mutateAsync: mockPublish, isPending: false }),
  useCreateExam: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateExam: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useGradingSessions: () => gradingSessionsState,
  useSessionEssays: () => sessionEssaysState,
  useGradeEssay: () => gradeEssayState,
}));

vi.mock("@/lib/hooks/admin-tests", () => ({
  useAdminTests: () => ({
    data: { data: [] as Test[] },
    isLoading: false,
  }),
}));

const sampleExam: ExamDetail = {
  id: "exam-1",
  title: "UTS Matematika",
  scheduled_at: "2026-07-01T08:00:00Z",
  timer_mode: "overall",
  duration_minutes: 90,
  is_free: false,
  requires_checkin: true,
  allow_leaderboard: false,
  randomize: false,
  status: "published",
  product_price: 50000,
  product_status: "published",
  tests: [],
};

const sampleSessions: GradingSessionItem[] = [
  {
    session_id: "session-1",
    student_id: "student-1",
    student_name: "Budi Santoso",
    submitted_at: "2026-06-30T10:00:00Z",
    ungraded_essay_count: 2,
  },
  {
    session_id: "session-2",
    student_id: "student-2",
    student_name: "Siti Aminah",
    submitted_at: "2026-06-30T11:00:00Z",
    ungraded_essay_count: 1,
  },
];

const sampleEssays: GradingEssayItem[] = [
  {
    question_id: "q-1",
    body: "Jelaskan penyebab perang Diponegoro",
    answer: "Karena penjajahan Belanda",
    point_correct: 10,
    score: null,
    grader_comment: null,
    graded_at: null,
  },
];

function openGradingTab() {
  fireEvent.click(screen.getByRole("button", { name: /^penilaian$/i }));
}

describe("ExamPackageDetailPage — grading tab", () => {
  beforeEach(() => {
    (useParams as ReturnType<typeof vi.fn>).mockReturnValue({ id: "exam-1" });
    examState = {
      data: sampleExam,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    gradingSessionsState = {
      data: { data: sampleSessions },
      isLoading: false,
      isError: false,
      error: null,
    };
    sessionEssaysState = {
      data: { data: sampleEssays },
      isLoading: false,
      isError: false,
      error: null,
    };
    gradeEssayState = { mutateAsync: mockGradeEssay, isPending: false, variables: undefined };
    mockGradeEssay.mockReset();
    mockGradeEssay.mockResolvedValue({ status: "graded", score: 8 });
  });

  it("renders the list of sessions needing grading", async () => {
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
    expect(screen.getByText("Siti Aminah")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("shows an empty state when no sessions need grading", async () => {
    gradingSessionsState = { data: { data: [] }, isLoading: false, isError: false, error: null };
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText(/tidak ada sesi yang perlu dinilai/i)).toBeInTheDocument();
    });
  });

  it("selecting a session shows its essays", async () => {
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });

    const row = screen.getByText("Budi Santoso").closest("tr");
    expect(row).toBeTruthy();
    fireEvent.click(within(row as HTMLElement).getByRole("button", { name: /lihat detail/i }));

    await waitFor(() => {
      expect(screen.getByText("Jelaskan penyebab perang Diponegoro")).toBeInTheDocument();
    });
    expect(screen.getByText("Karena penjajahan Belanda")).toBeInTheDocument();
  });

  it("grading an essay calls useGradeEssay with the score and comment payload", async () => {
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
    fireEvent.click(
      within(screen.getByText("Budi Santoso").closest("tr") as HTMLElement).getByRole("button", {
        name: /lihat detail/i,
      }),
    );

    await waitFor(() => {
      expect(screen.getByText("Jelaskan penyebab perang Diponegoro")).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/^skor$/i), { target: { value: "8" } });
    fireEvent.change(screen.getByLabelText(/komentar/i), {
      target: { value: "Jawaban cukup lengkap" },
    });
    fireEvent.click(screen.getByRole("button", { name: /simpan nilai/i }));

    await waitFor(() => {
      expect(mockGradeEssay).toHaveBeenCalledWith({
        question_id: "q-1",
        score: 8,
        comment: "Jawaban cukup lengkap",
      });
    });
  });

  it("blocks the save and does not call useGradeEssay when the score exceeds point_correct", async () => {
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
    fireEvent.click(
      within(screen.getByText("Budi Santoso").closest("tr") as HTMLElement).getByRole("button", {
        name: /lihat detail/i,
      }),
    );

    await waitFor(() => {
      expect(screen.getByText("Jelaskan penyebab perang Diponegoro")).toBeInTheDocument();
    });

    // point_correct for this essay is 10; 11 is out of bounds.
    fireEvent.change(screen.getByLabelText(/^skor$/i), { target: { value: "11" } });
    fireEvent.click(screen.getByRole("button", { name: /simpan nilai/i }));

    expect(mockGradeEssay).not.toHaveBeenCalled();
  });

  it("clears a session from the grading queue after it is fully graded (post-invalidation refetch)", async () => {
    render(<ExamPackageDetailPage />);
    openGradingTab();

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
    fireEvent.click(
      within(screen.getByText("Budi Santoso").closest("tr") as HTMLElement).getByRole("button", {
        name: /lihat detail/i,
      }),
    );

    await waitFor(() => {
      expect(screen.getByText("Jelaskan penyebab perang Diponegoro")).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/^skor$/i), { target: { value: "8" } });
    fireEvent.click(screen.getByRole("button", { name: /simpan nilai/i }));

    await waitFor(() => {
      expect(mockGradeEssay).toHaveBeenCalled();
    });

    // Simulate the queue refetch that useGradeEssay's onSuccess invalidation triggers.
    gradingSessionsState = {
      data: { data: sampleSessions.filter((s) => s.session_id !== "session-1") },
      isLoading: false,
      isError: false,
      error: null,
    };

    fireEvent.click(screen.getByRole("button", { name: /kembali ke daftar/i }));

    await waitFor(() => {
      expect(screen.queryByText("Budi Santoso")).not.toBeInTheDocument();
    });
    expect(screen.getByText("Siti Aminah")).toBeInTheDocument();
  });
});
