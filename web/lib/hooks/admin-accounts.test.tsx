import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminAccounts,
  useCreateAdminAccount,
  useChangeAccountRole,
  useChangeAccountStatus,
  useResetAccountPassword,
  adminAccountsKeys,
} from "./admin-accounts";
import type { AdminAccount } from "@/lib/types";

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

describe("admin-accounts hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminAccounts fetches GET /admin/system/accounts and returns data", async () => {
    const accounts: AdminAccount[] = [
      { id: "a1", name: "Admin A", email: "admin@test.com", role: "admin_store", status: "active", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
    ];
    mockAuthFetch.mockResolvedValueOnce({ data: accounts });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminAccounts(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts");
    expect(result.current.data).toEqual(accounts);
  });

  it("useAdminAccounts filters by role and status", async () => {
    mockAuthFetch.mockResolvedValueOnce({ data: [] });

    const { wrapper } = wrapperFactory();
    renderHook(() => useAdminAccounts("admin_store", "active"), { wrapper });

    await waitFor(() => expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts?role=admin_store&status=active"));
  });

  it("useCreateAdminAccount posts to /admin/system/accounts and invalidates list", async () => {
    const account: AdminAccount = { id: "a2", name: "Admin B", email: "adminb@test.com", role: "admin_school", status: "active", created_at: "2026-02-01T00:00:00Z", updated_at: "2026-02-01T00:00:00Z" };
    mockAuthFetch.mockResolvedValueOnce(account);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateAdminAccount(), { wrapper });

    const input = { email: "adminb@test.com", name: "Admin B", role: "admin_school" as const, password: "secret123" };

    await act(async () => {
      await result.current.mutateAsync(input);
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts", {
      method: "POST",
      body: JSON.stringify(input),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminAccountsKeys.all });
  });

  it("useChangeAccountRole patches /admin/system/accounts/:id/role and invalidates list", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "role updated" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useChangeAccountRole(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id: "a1", role: "super_admin" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts/a1/role", {
      method: "PATCH",
      body: JSON.stringify({ role: "super_admin" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminAccountsKeys.all });
  });

  it("useChangeAccountStatus patches /admin/system/accounts/:id/status and invalidates list", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "status updated" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useChangeAccountStatus(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id: "a1", status: "deactivated" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts/a1/status", {
      method: "PATCH",
      body: JSON.stringify({ status: "deactivated" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminAccountsKeys.all });
  });

  it("useResetAccountPassword posts to /admin/system/accounts/:id/reset-password and does not invalidate list", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "password reset triggered" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useResetAccountPassword(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("a1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/system/accounts/a1/reset-password", {
      method: "POST",
    });
    expect(spy).not.toHaveBeenCalled();
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
