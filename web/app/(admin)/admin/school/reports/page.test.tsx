import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import SchoolReportsPage from "./page";
import type { Product, AdminResultRow, AdminResultDetail } from "@/lib/types";

const mockExport = vi.fn();

// Mock the exam products hook
const mockProducts = vi.fn();

// Mock admin results hooks
let resultsState = {
  data: null as { data: AdminResultRow[]; next_cursor?: string } | null,
  isLoading: false,
  isFetching: false,
  isError: false,
  error: null as Error | null,
};

let detailState = {
  data: null as AdminResultDetail | null,
  isLoading: false,
  isFetching: false,
  isError: false,
  error: null as Error | null,
};

vi.mock("@/lib/hooks/products", () => ({
  useProducts: () => mockProducts(),
}));

vi.mock("@/lib/hooks/admin-results", () => ({
  useAdminResults: () => resultsState,
  useAdminResultDetail: (sessionId: string) => detailState,
  exportAdminResults: (...args: Parameters<typeof mockExport>) => mockExport(...args),
}));

const sampleExamProducts = [
  {
    id: "exam-1",
    type: "exam",
    name: "Tryout Matematika",
    price: 50000,
    status: "published",
  },
  {
    id: "exam-2",
    type: "exam",
    name: "Tryout Fisika",
    price: 0,
    status: "published",
  },
] as unknown as Product[];

const sampleResultRows: AdminResultRow[] = [
  {
    session_id: "s1",
    name: "Budi Santoso",
    nis: "12345",
    score: 85,
    submitted_at: "2026-01-15T00:00:00Z",
  },
  {
    session_id: "s2",
    name: "Siti Aisyah",
    nis: "67890",
    score: 92,
    submitted_at: "2026-02-20T00:00:00Z",
  },
];

const paginatedResponse = (rows: AdminResultRow[]) => ({
  data: rows,
  next_cursor: undefined,
});

const scoreOnlyDetail: AdminResultDetail = {
  session_id: "s1",
  name: "Budi Santoso",
  nis: "12345",
  score: 85,
  submitted_at: "2026-01-15T00:00:00Z",
  result_config: "score_only",
  correct_count: 10,
  wrong_count: 2,
  empty_count: 1,
};

const scorePembahasanDetail: AdminResultDetail = {
  session_id: "s2",
  name: "Siti Aisyah",
  nis: "67890",
  score: 92,
  submitted_at: "2026-02-20T00:00:00Z",
  result_config: "score_pembahasan",
  correct_count: 12,
  wrong_count: 1,
  empty_count: 0,
  breakdown: [
    { test_id: "t1", title: "Aljabar", subject: "Matematika", topic: "Aljabar", earned: 8, max: 10 },
  ],
  pembahasan: [
    {
      question_id: "q1",
      body: "Berapa 2+2?",
      format: "mcq",
      your_answer: "4",
      correct_answer: "4",
      is_correct: true,
      explanation: "2+2=4",
    },
  ],
};

describe("SchoolReportsPage", () => {
  beforeEach(() => {
    mockProducts.mockReturnValue({ data: sampleExamProducts, isLoading: false });
    mockExport.mockReset();

    resultsState = {
      data: paginatedResponse(sampleResultRows),
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    detailState = {
      data: null,
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    URL.createObjectURL = vi.fn().mockReturnValue("blob:test-url");
    URL.revokeObjectURL = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ── Test 1: Picker renders exam options from useProducts("exam") — mocked in test ──
  it("renders exam picker with options from useProducts", async () => {
    render(<SchoolReportsPage />);

    // Open the select dropdown to show options
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);

    await waitFor(() => {
      expect(screen.getByText("Tryout Matematika")).toBeInTheDocument();
      expect(screen.getByText("Tryout Fisika")).toBeInTheDocument();
    });
  });

  // ── Test 2: Table renders rows, search updates query, "load more" appends without duplicating ──
  it("renders results table when an exam is selected", async () => {
    render(<SchoolReportsPage />);

    // Select an exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);

    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
      expect(screen.getByText("Siti Aisyah")).toBeInTheDocument();
    });

    // Check NIS and score columns
    expect(screen.getByText("12345")).toBeInTheDocument();
    expect(screen.getByText("67890")).toBeInTheDocument();
    expect(screen.getByText("85")).toBeInTheDocument();
    expect(screen.getByText("92")).toBeInTheDocument();
  });

  it("shows load more button when next_cursor exists", async () => {
    resultsState = {
      data: { data: sampleResultRows, next_cursor: "cursor-next" },
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    render(<SchoolReportsPage />);

    // Select an exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);

    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Muat lebih")).toBeInTheDocument();
    });
  });

  it("clicking load more does not crash", async () => {
    resultsState = {
      data: { data: sampleResultRows, next_cursor: "cursor-2" },
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    render(<SchoolReportsPage />);

    // Select exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });

    // Click "load more" — should not crash
    const loadMoreBtn = screen.getByText("Muat lebih");
    fireEvent.click(loadMoreBtn);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
      expect(screen.getByText("Siti Aisyah")).toBeInTheDocument();
    });
  });

  it("search input updates query", async () => {
    render(<SchoolReportsPage />);

    // Select an exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    // Find search input
    const searchInput = screen.getByPlaceholderText(/cari/i);
    expect(searchInput).toBeInTheDocument();

    fireEvent.input(searchInput, { target: { value: "Budi" } });

    await waitFor(() => {
      expect(searchInput).toHaveValue("Budi");
    });
  });

  // ── Test 3: Mocked empty list renders neutral empty state, NOT an error banner ──
  it("renders neutral empty state when results list is empty", async () => {
    resultsState = {
      data: { data: [], next_cursor: undefined },
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    render(<SchoolReportsPage />);

    // Select an exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText(/belum ada hasil/i)).toBeInTheDocument();
    });

    // Verify it's NOT an error banner
    expect(screen.queryByText(/gagal/i)).not.toBeInTheDocument();
  });

  // ── Test 4: Drill-down dialog shows breakdown/pembahasan only for score_pembahasan ──
  it("shows score_only detail in drill-down dialog (no breakdown/pembahasan)", async () => {
    detailState = {
      data: scoreOnlyDetail,
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    render(<SchoolReportsPage />);

    // Select exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });

    // Click a row to drill down
    const row = screen.getByText("Budi Santoso").closest("tr") || screen.getByText("Budi Santoso");
    fireEvent.click(row);

    await waitFor(() => {
      expect(screen.getByText("Detail Hasil")).toBeInTheDocument();
    });

    // For score_only: no breakdown section, no pembahasan section
    expect(screen.queryByText("Berdasarkan Topik")).not.toBeInTheDocument();
    expect(screen.queryByText("Pembahasan")).not.toBeInTheDocument();

    // Close dialog
    const closeBtn = screen.getByRole("button", { name: /batal/i });
    fireEvent.click(closeBtn);
  });

  it("shows breakdown and pembahasan for score_pembahasan drill-down", async () => {
    detailState = {
      data: scorePembahasanDetail,
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
    };

    render(<SchoolReportsPage />);

    // Select exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Siti Aisyah")).toBeInTheDocument();
    });

    // Click a row
    const row = screen.getByText("Siti Aisyah").closest("tr") || screen.getByText("Siti Aisyah");
    fireEvent.click(row);

    await waitFor(() => {
      expect(screen.getByText("Detail Hasil")).toBeInTheDocument();
    });

    // score_pembahasan: breakdown and pembahasan should be visible
    expect(screen.getByText("Berdasarkan Topik")).toBeInTheDocument();
    expect(screen.getByText("Pembahasan")).toBeInTheDocument();
  });

  // ── Test 5: No rendered DOM contains substring "rank" — grep test ──
  it("does not render the word 'rank' anywhere in the DOM", async () => {
    render(<SchoolReportsPage />);

    // Select exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    const html = document.body.innerHTML;
    expect(html.toLowerCase()).not.toContain("rank");
  });

  // ── Test 6: Export button calls export function with current examId, doesn't crash ──
  it("renders export button disabled when no exam selected", async () => {
    render(<SchoolReportsPage />);

    const exportBtn = screen.getByRole("button", { name: /ekspor/i });
    expect(exportBtn).toBeDisabled();
  });

  it("calls exportAdminResults with selected examId on export click", async () => {
    mockExport.mockResolvedValueOnce(undefined);

    render(<SchoolReportsPage />);

    // Select exam
    const selectTrigger = screen.getByRole("combobox");
    fireEvent.click(selectTrigger);
    const examOption = await screen.findByText("Tryout Matematika");
    fireEvent.click(examOption);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });

    const exportBtn = screen.getByRole("button", { name: /ekspor/i });
    expect(exportBtn).not.toBeDisabled();

    fireEvent.click(exportBtn);

    expect(mockExport).toHaveBeenCalledWith("exam-1");
  });
});
