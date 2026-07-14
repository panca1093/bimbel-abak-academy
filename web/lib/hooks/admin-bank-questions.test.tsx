import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useBankQuestions,
  useCreateBankQuestion,
  useUpdateBankQuestion,
  useDeleteBankQuestion,
  useImportBankQuestions,
  adminBankQuestionsKeys,
} from "./admin-bank-questions";
import type { AdminQuestionImportResponse } from "@/lib/types";

const mockAuthFetch = vi.fn();
const mockAuthFetchMultipart = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
  authFetchMultipart: (...args: Parameters<typeof mockAuthFetchMultipart>) =>
    mockAuthFetchMultipart(...args),
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
    mockAuthFetchMultipart.mockReset();
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

  it("useImportBankQuestion_posts_multipart_and_invalidates", async () => {
    const response: AdminQuestionImportResponse = {
      inserted: 1,
      rows: [{ row_number: 2, status: "error", error: "bad format" }],
    };
    mockAuthFetchMultipart.mockResolvedValueOnce(response);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useImportBankQuestions(), { wrapper });

    const file = new File(["csv"], "questions.csv", { type: "text/csv" });
    let returned: AdminQuestionImportResponse | undefined;
    await act(async () => {
      returned = await result.current.mutateAsync(file);
    });

    expect(returned).toEqual(response);
    const callInit = mockAuthFetchMultipart.mock.calls[0][1] as RequestInit | undefined;
    expect(callInit?.method).toBe("POST");
    const body = callInit?.body as FormData;
    expect(body.get("file")).toBe(file);
    expect(mockAuthFetchMultipart).toHaveBeenCalledWith("/admin/questions/import", expect.any(Object));
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
