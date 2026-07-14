import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import ExamPackagesPage from "./page";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

let uiStore = { lang: "id" as "id" | "en" };

let examsState = {
  data: undefined as { data: unknown[]; next_cursor?: string } | undefined,
  isLoading: true,
  isError: false,
  error: null as Error | null,
};

vi.mock("@/lib/hooks/admin-exams", () => ({
  useExams: () => examsState,
  useCreateExam: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateExam: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

const sampleExams = [
  {
    id: "e1",
    title: "UTS Matematika",
    scheduled_at: "2026-07-01T08:00:00Z",
    is_free: false,
    product_status: "draft",
    product_price: 50000,
    timer_mode: "overall",
    duration_minutes: 90,
    requires_checkin: true,
    allow_leaderboard: true,
    randomize: false,
  },
  {
    id: "e2",
    title: "UAS IPA",
    scheduled_at: "2026-07-15T09:00:00Z",
    is_free: true,
    product_status: "published",
    product_price: 0,
    timer_mode: "per_test",
    duration_minutes: null,
    requires_checkin: false,
    allow_leaderboard: false,
    randomize: true,
  },
];

describe("ExamPackagesPage", () => {
  beforeEach(() => {
    push.mockReset();
  });

  it("navigates to the exam detail page when a row is clicked", async () => {
    examsState = {
      data: { data: sampleExams },
      isLoading: false,
      isError: false,
      error: null,
    };

    render(<ExamPackagesPage />);

    await waitFor(() => {
      expect(screen.getByText("UTS Matematika")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("UTS Matematika"));
    expect(push).toHaveBeenCalledWith("/admin/exam/packages/e1");
  });

  it("renders the packages table with exam titles and the create button", async () => {
    examsState = {
      data: { data: sampleExams },
      isLoading: false,
      isError: false,
      error: null,
    };

    render(<ExamPackagesPage />);

    await waitFor(() => {
      expect(screen.getByText("UTS Matematika")).toBeInTheDocument();
      expect(screen.getByText("UAS IPA")).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: /buat paket/i })).toBeInTheDocument();
  });

  it("shows skeleton rows while loading", () => {
    examsState = {
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    };

    render(<ExamPackagesPage />);

    expect(document.querySelectorAll("[data-slot=skeleton]").length).toBeGreaterThan(0);
  });

  it("surfaces an API error as inline error text", async () => {
    examsState = {
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error("gagal memuat paket"),
      refetch: vi.fn(),
    } as any;

    render(<ExamPackagesPage />);

    await waitFor(() => {
      expect(screen.getByText(/gagal memuat paket/i)).toBeInTheDocument();
    });
  });

  it("shows empty state when no packages exist", async () => {
    examsState = {
      data: { data: [] },
      isLoading: false,
      isError: false,
      error: null,
    };

    render(<ExamPackagesPage />);

    await waitFor(() => {
      expect(screen.getByText(/belum ada paket/i)).toBeInTheDocument();
    });
  });
});