import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAdminRevenue, adminRevenueKeys } from "./admin-revenue";
import type { AdminRevenue } from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
  ApiError: class extends Error {
    code: string;
    status: number;
    constructor(code: string, message: string, status: number) {
      super(message);
      this.code = code;
      this.status = status;
    }
  },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: {
    getState: () => ({ token: "test-token" }),
  },
}));

describe("admin-revenue hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminRevenue fetches GET /admin/revenue and returns data", async () => {
    const revenue: AdminRevenue = {
      total: 1_500_000,
      by_type: {
        book: { total: 500_000, count: 5 },
        course: { total: 1_000_000, count: 2 },
      },
    };
    mockAuthFetch.mockResolvedValueOnce(revenue);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminRevenue(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/revenue");
    expect(result.current.data).toEqual(revenue);
  });

  it("uses stable query key", () => {
    expect(adminRevenueKeys.list()).toEqual(["admin", "revenue", "list"]);
  });
});

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
    queryClient,
  };
}
