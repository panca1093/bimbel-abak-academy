import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAdminSystemConfig, useUpdateSystemConfig, adminConfigKeys } from "./admin-config";
import type { SystemConfig } from "@/lib/types";

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

describe("admin-config hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminSystemConfig fetches GET /admin/system/config and returns data", async () => {
    const config: SystemConfig = { app_name: "Akademi Bimbel", app_address: "Jakarta" };
    mockAuthFetch.mockResolvedValueOnce(config);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminSystemConfig(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/config");
    expect(result.current.data).toEqual(config);
  });

  it("useUpdateSystemConfig puts /admin/system/config and invalidates config", async () => {
    const updated: SystemConfig = { app_name: "Akademi Bimbel Updated", app_address: "Jakarta" };
    mockAuthFetch.mockResolvedValueOnce(updated);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateSystemConfig(), { wrapper });

    const input = { app_name: "Akademi Bimbel Updated" };

    await act(async () => {
      await result.current.mutateAsync(input);
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/config", {
      method: "PUT",
      body: JSON.stringify(input),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminConfigKeys.all });
  });

  it("uses stable query key", () => {
    expect(adminConfigKeys.detail()).toEqual(["admin", "config", "detail"]);
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
