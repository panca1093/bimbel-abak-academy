import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useTopics,
  useCreateTopic,
  useUpdateTopic,
  useDeleteTopic,
  adminTopicsKeys,
} from "./admin-topics";

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

describe("admin-topics hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useTopics_fetches_with_subject_filter", async () => {
    mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: "" });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useTopics("math"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/topics?subject=math");
  });

  it("useCreateTopic_calls_POST_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce({
      id: "topic-1",
      name: "Algebra",
      subject: "math",
    });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateTopic(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ name: "Algebra", subject: "math" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/topics", {
      method: "POST",
      body: JSON.stringify({ name: "Algebra", subject: "math" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminTopicsKeys.lists() });
  });

  it("useUpdateTopic_calls_PATCH_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce({
      id: "topic-1",
      name: "Geometry",
      subject: "math",
    });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateTopic("topic-1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ name: "Geometry" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/topics/topic-1", {
      method: "PATCH",
      body: JSON.stringify({ name: "Geometry" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminTopicsKeys.lists() });
    expect(spy).toHaveBeenCalledWith({
      queryKey: adminTopicsKeys.detail("topic-1"),
    });
  });

  it("useDeleteTopic_calls_DELETE_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce(undefined);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteTopic(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("topic-1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/topics/topic-1", {
      method: "DELETE",
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminTopicsKeys.lists() });
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
