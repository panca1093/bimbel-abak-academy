import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useJobStatus } from "./jobs";
import type { JobStatus } from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) =>
    mockAuthFetch(...args),
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

describe("useJobStatus", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("does not fetch when enabled=false", async () => {
    const { wrapper } = wrapperFactory();
    renderHook(() => useJobStatus("job-1", { enabled: false }), { wrapper });

    // Give react-query a tick to potentially fire.
    await new Promise((r) => setTimeout(r, 50));
    expect(mockAuthFetch).not.toHaveBeenCalled();
  });

  it("fetches GET /admin/jobs/:id and returns the JobStatus row", async () => {
    const job: JobStatus = {
      id: "job-1",
      type: "student_bulk",
      status: "running",
      progress: 40,
      result_url: null,
      error: null,
      created_at: "2026-07-18T00:00:00Z",
      updated_at: "2026-07-18T00:01:00Z",
    };
    mockAuthFetch.mockResolvedValueOnce(job);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useJobStatus("job-1"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/jobs/job-1");
    expect(result.current.data).toEqual(job);
  });

  it("stops polling once status reaches a terminal value (succeeded)", async () => {
    const terminal: JobStatus = {
      id: "job-1",
      type: "student_bulk",
      status: "succeeded",
      progress: 100,
      result_url: "http://minio/result.csv?sig=zzz",
      error: null,
      created_at: "2026-07-18T00:00:00Z",
      updated_at: "2026-07-18T00:02:00Z",
    };
    // First call: still running. Second call: succeeded (terminal). No third call expected.
    const running: JobStatus = { ...terminal, status: "running", progress: 80, result_url: null };
    mockAuthFetch.mockResolvedValueOnce(running).mockResolvedValueOnce(terminal);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useJobStatus("job-1"), { wrapper });

    await waitFor(() => expect(result.current.data?.status).toBe("running"));

    // Wait past the 2s poll interval so react-query fires the next refetch.
    await new Promise((r) => setTimeout(r, 2500));
    await waitFor(() => expect(result.current.data?.status).toBe("succeeded"));

    // After hitting terminal, no further fetches should be scheduled. Wait
    // well past the poll interval and confirm call count is unchanged.
    const callCountAfterTerminal = mockAuthFetch.mock.calls.length;
    await new Promise((r) => setTimeout(r, 5000));
    expect(mockAuthFetch.mock.calls.length).toBe(callCountAfterTerminal);
    // Specifically, the data should expose progress + result_url to the caller.
    expect(result.current.data?.progress).toBe(100);
    expect(result.current.data?.result_url).toBe(
      "http://minio/result.csv?sig=zzz",
    );
  }, 15000);

  it("stops polling once status reaches a terminal value (failed)", async () => {
    const terminal: JobStatus = {
      id: "job-2",
      type: "student_bulk",
      status: "failed",
      progress: 30,
      result_url: null,
      error: "boom",
      created_at: "2026-07-18T00:00:00Z",
      updated_at: "2026-07-18T00:00:30Z",
    };
    mockAuthFetch.mockResolvedValueOnce(terminal);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useJobStatus("job-2"), { wrapper });

    await waitFor(() => expect(result.current.data?.status).toBe("failed"));

    expect(result.current.data?.error).toBe("boom");
    const callCount = mockAuthFetch.mock.calls.length;
    await new Promise((r) => setTimeout(r, 50));
    expect(mockAuthFetch.mock.calls.length).toBe(callCount);
  });

  it("does not poll when jobId is empty/falsy", async () => {
    const { wrapper } = wrapperFactory();
    renderHook(() => useJobStatus(""), { wrapper });

    await new Promise((r) => setTimeout(r, 50));
    expect(mockAuthFetch).not.toHaveBeenCalled();
  });
});
