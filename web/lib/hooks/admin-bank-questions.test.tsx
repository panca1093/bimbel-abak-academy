import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useBankQuestions,
  useCreateBankQuestion,
  useUpdateBankQuestion,
  useDeleteBankQuestion,
  adminBankQuestionsKeys,
} from "./admin-bank-questions";

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

describe("admin-bank-questions hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useBankQuestions_fetches_with_filters", async () => {
    mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: "" });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(
      () => useBankQuestions({ format: "mcq", topic_id: "topic-1", search: "foo" }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const called = mockAuthFetch.mock.calls[0][0] as string;
    expect(called).toContain("format=mcq");
    expect(called).toContain("topic_id=topic-1");
    expect(called).toContain("search=foo");
  });

  it("useCreateBankQuestion_calls_POST_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce({ question: { id: "q1" }, options: [] });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateBankQuestion(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        format: "mcq",
        body: "Bank Q",
        sort_order: 1,
        topic_id: "topic-1",
      });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/questions", {
      method: "POST",
      body: JSON.stringify({
        format: "mcq",
        body: "Bank Q",
        sort_order: 1,
        topic_id: "topic-1",
      }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminBankQuestionsKeys.lists() });
  });

  it("useUpdateBankQuestion_calls_PATCH_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce({ question: { id: "q1" }, options: [] });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateBankQuestion("q1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        format: "mcq",
        body: "Updated Bank Q",
        sort_order: 1,
      });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/questions/q1", {
      method: "PATCH",
      body: JSON.stringify({
        format: "mcq",
        body: "Updated Bank Q",
        sort_order: 1,
      }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminBankQuestionsKeys.lists() });
    expect(spy).toHaveBeenCalledWith({
      queryKey: adminBankQuestionsKeys.detail("q1"),
    });
  });

  it("useDeleteBankQuestion_calls_DELETE_and_invalidates", async () => {
    mockAuthFetch.mockResolvedValueOnce(undefined);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteBankQuestion(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("q1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/questions/q1", {
      method: "DELETE",
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminBankQuestionsKeys.lists() });
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
