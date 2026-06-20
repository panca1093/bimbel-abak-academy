import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import DashboardPage from "./page";
import type { Dashboard } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
}));

let authStore = {
  user: null as { name?: string; username?: string } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

interface DashboardHookState {
  data: Dashboard | null;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: ReturnType<typeof vi.fn>;
}

const baseData: Dashboard = {
  enrolled_courses: [
    {
      id: "c1",
      title: "Matematika Dasar",
      progress: 0.5,
      total_lessons: 10,
      done_lessons: 5,
    },
    {
      id: "c2",
      title: "Fisika Lanjutan",
      progress: 0.3,
      total_lessons: 8,
      done_lessons: 2,
    },
  ],
  study_summary: {
    visited_lectures: 7,
    total_lectures: 18,
    enrolled_courses_count: 2,
    completed_courses: 0,
    total_minutes: 420,
  },
  ranking: {
    position: 3,
    points: 844,
    leaderboard: [
      { rank: 1, name: "Nadia Salsabila", points: 871, is_me: false },
      { rank: 2, name: "Reza Pratama", points: 858, is_me: false },
      { rank: 3, name: "Budi Santoso", points: 844, is_me: true },
    ],
  },
  exam_progress: [
    { label: "Tryout 1", completed: 148, in_progress: 82 },
    { label: "Tryout 2", completed: 162, in_progress: 104 },
  ],
  popular_lessons: [
    { title: "Penalaran Matematika HOTS", topics: 12, students: 240, duration: "1j 30m", progress: 0.72 },
    { title: "Literasi Bahasa Inggris", topics: 9, students: 198, duration: "45m", progress: 0.54 },
  ],
};

let dashboardState: DashboardHookState = {
  data: baseData,
  isLoading: false,
  isError: false,
  error: null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/students", () => ({
  useDashboard: () => dashboardState,
  useProfile: () => ({ data: null, isLoading: false, isError: false, refetch: vi.fn() }),
  useUpdateProfile: () => ({ mutate: vi.fn(), isPending: false }),
  useChangePassword: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("DashboardPage", () => {
  beforeEach(() => {
    authStore = { user: { name: "Budi Santoso" } };
    dashboardState = {
      data: { ...baseData },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("renders greeting with user name", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText(/Halo, Budi|Hello, Budi/)).toBeInTheDocument();
    });
  });

  it("renders course cards from enrolled_courses", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText("Matematika Dasar")).toBeInTheDocument();
      expect(screen.getByText("Fisika Lanjutan")).toBeInTheDocument();
    });
  });

  it("shows my-ranking card with leaderboard entries", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText("Nadia Salsabila")).toBeInTheDocument();
      expect(screen.getByText("Reza Pratama")).toBeInTheDocument();
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
  });

  it("shows empty state for ranking when leaderboard is empty", async () => {
    dashboardState = {
      ...dashboardState,
      data: {
        ...baseData,
        ranking: { position: null, points: null, leaderboard: [] },
      },
    };
    render(<DashboardPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/Belum ada data peringkat|No ranking data yet/),
      ).toBeInTheDocument();
    });
  });

  it("shows empty state for exam_progress when empty", async () => {
    dashboardState = {
      ...dashboardState,
      data: { ...baseData, exam_progress: [] },
    };
    render(<DashboardPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/Belum ada data progres ujian|No exam progress data yet/),
      ).toBeInTheDocument();
    });
  });

  it("shows empty state for popular_lessons when empty", async () => {
    dashboardState = {
      ...dashboardState,
      data: { ...baseData, popular_lessons: [] },
    };
    render(<DashboardPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/Belum ada data pelajaran populer|No popular lessons data yet/),
      ).toBeInTheDocument();
    });
  });

  it("does not show prototype placeholder mock numbers", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      // Prototype had hardcoded "120 jam" and "45m" but these come from real data now
      expect(screen.queryByText(/120 jam|45h/)).not.toBeInTheDocument();
    });
  });

  it("renders loading skeleton when isLoading is true", async () => {
    dashboardState = { data: null, isLoading: true, isError: false, error: null, refetch: vi.fn() };
    render(<DashboardPage />);
    expect(screen.getByTestId("dashboard-skeleton")).toBeInTheDocument();
  });

  it("renders error card when isError is true", async () => {
    dashboardState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("Network Error"),
      refetch: vi.fn(),
    };
    render(<DashboardPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/Gagal memuat dashboard|Failed to load dashboard/),
      ).toBeInTheDocument();
      expect(screen.getByText(/Network Error/)).toBeInTheDocument();
    });
  });

  it("renders pending banner when pending_order exists", async () => {
    dashboardState = {
      ...dashboardState,
      data: {
        ...baseData,
        pending_order: { id: "ord1", product: "Buku Soal SNBT", amount: 75000 },
      },
    };
    render(<DashboardPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/Pembayaran tertunda|Payment pending/),
      ).toBeInTheDocument();
    });
  });
});
