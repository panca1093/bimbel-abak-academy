import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useGoogleLogin, useLogout } from "./auth";

const apiFetch = vi.fn();
const setSession = vi.fn();
const clear = vi.fn();

vi.mock("@/lib/api", () => ({
  apiFetch: (...args: unknown[]) => apiFetch(...args),
  authFetch: vi.fn(),
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: (
    selector: (state: { setSession: typeof setSession; clear: typeof clear }) => unknown,
  ) => selector({ setSession, clear }),
}));

describe("useGoogleLogin", () => {
  beforeEach(() => {
    apiFetch.mockReset();
    setSession.mockReset();
  });

  it("posts the id token and stores the returned session", async () => {
    const user = { id: "user-1", role: "student", auth_provider: "google" };
    apiFetch.mockResolvedValue({
      access_token: "access",
      refresh_token: "refresh",
      user,
    });
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useGoogleLogin(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id_token: "google-token" });
    });

    expect(apiFetch).toHaveBeenCalledWith("/auth/google", {
      method: "POST",
      body: JSON.stringify({ id_token: "google-token" }),
    });
    expect(setSession).toHaveBeenCalledWith("access", "refresh", user);
  });
});

describe("useLogout", () => {
  beforeEach(() => {
    clear.mockReset();
  });

  it("clears the react-query cache so a prior session's data can't leak to the next user", async () => {
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    // Seed a cached profile from the previous session.
    queryClient.setQueryData(["students", "profile"], { id: "prev-user" });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useLogout(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync();
    });

    expect(clear).toHaveBeenCalled();
    expect(queryClient.getQueryData(["students", "profile"])).toBeUndefined();
  });
});
