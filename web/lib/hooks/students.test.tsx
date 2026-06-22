import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useUpdatePhoto, studentsKeys } from "./students";
import type { User } from "@/lib/types";

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

const mockSetSession = vi.fn();

vi.mock("@/stores/auth", () => ({
  useAuthStore: {
    getState: () => ({
      token: "test-token-123",
      setSession: mockSetSession,
    }),
  },
}));

describe("useUpdatePhoto", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
    mockSetSession.mockClear();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("calls authFetch PATCH and updates auth store on success", async () => {
    const updatedUser: User = {
      id: "u1",
      name: "Budi Santoso",
      email: "budi@test.com",
      photo_url: "https://example.com/new-photo.jpg",
    };
    mockAuthFetch.mockResolvedValueOnce(updatedUser);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdatePhoto(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("https://example.com/new-photo.jpg");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/students/photo", {
      method: "PATCH",
      body: JSON.stringify({ photo_url: "https://example.com/new-photo.jpg" }),
    });

    // Auth store should be updated with token, refreshToken, and returned user
    expect(mockSetSession).toHaveBeenCalledWith("test-token-123", "", updatedUser);

    // Profile query should still be invalidated
    expect(spy).toHaveBeenCalledWith({ queryKey: studentsKeys.profile() });
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
