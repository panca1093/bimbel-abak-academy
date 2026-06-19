import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import DashboardPage from "./page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
}));

let authStore = {
  user: null as { name?: string; username?: string } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

let dashboardState = {
  data: null as {
    enrolled_courses: Array<Record<string, unknown>>;
    pending_order?: { id: string; product?: string; amount: number };
  } | null,
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
      data: { enrolled_courses: [] },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("renders the three exam-dependent coming-soon stub cards", async () => {
    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: /Peringkat saya|My Ranking/i })).toBeInTheDocument();
      expect(screen.getByRole("heading", { name: /Total jam belajar|Total Hours/i })).toBeInTheDocument();
      expect(screen.getByRole("heading", { name: /Progress ujian|Exam Progress/i })).toBeInTheDocument();
    });
  });

  it("does not fetch fake dashboard data for the stub cards", async () => {
    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.queryByText(/#1|rank 1/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/120 jam|45h/i)).not.toBeInTheDocument();
    });
  });
});
