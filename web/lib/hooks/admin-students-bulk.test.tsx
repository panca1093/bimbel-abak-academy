import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  usePresignStudentBulkUpload,
  useEnqueueStudentBulkImport,
  putFileToPresignedURL,
} from "./admin-students-bulk";

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

describe("admin-students-bulk hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
    vi.restoreAllMocks();
  });

  describe("usePresignStudentBulkUpload", () => {
    it("posts to /admin/students/bulk/presign with filename and content_type", async () => {
      const presignResp = {
        url: "http://minio.local/student-bulk/school-1/uuid-x.csv?sig=abc",
        method: "PUT",
        key: "student-bulk/school-1/uuid-x.csv",
      };
      mockAuthFetch.mockResolvedValueOnce(presignResp);

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => usePresignStudentBulkUpload(), {
        wrapper,
      });

      let returned: { url: string; method: string; key: string } | undefined;
      await act(async () => {
        returned = await result.current.mutateAsync({
          filename: "students.csv",
          contentType: "text/csv",
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith(
        "/admin/students/bulk/presign?filename=students.csv&content_type=text%2Fcsv",
        { method: "POST" },
      );
      expect(returned).toEqual(presignResp);
    });

    it("encodes filename and content_type (spaces, parens, semicolons safe in the URL)", async () => {
      mockAuthFetch.mockResolvedValueOnce({
        url: "http://minio/k",
        method: "PUT",
        key: "k",
      });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => usePresignStudentBulkUpload(), {
        wrapper,
      });

      await act(async () => {
        await result.current.mutateAsync({
          filename: "my students (1).csv",
          contentType: "text/csv; charset=utf-8",
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledTimes(1);
      const callArg = mockAuthFetch.mock.calls[0][0] as string;
      // URLSearchParams uses form-encoding (+ for space, %XX for others); the
      // backend's c.QueryParam accepts both, so the contract is just "no raw
      // unencoded chars in the query string". Spot-check the dangerous ones.
      expect(callArg).toContain("filename=my+students+%281%29.csv");
      expect(callArg).toContain("content_type=text%2Fcsv%3B+charset%3Dutf-8");
      // And nothing raw of the special chars.
      expect(callArg).not.toMatch(/filename=my students/);
      expect(callArg).not.toMatch(/content_type=text\/csv;/);
    });
  });

  describe("putFileToPresignedURL", () => {
    it("sends a raw PUT with the file body and Content-Type, no Authorization header", async () => {
      const fakeFetch = vi.fn().mockResolvedValueOnce({ ok: true });
      vi.stubGlobal("fetch", fakeFetch);

      const file = new Blob(["a,b,c\n1,2,3"], { type: "text/csv" });
      await putFileToPresignedURL(
        "http://minio.local/key?sig=xxx",
        file,
        "text/csv",
      );

      expect(fakeFetch).toHaveBeenCalledTimes(1);
      const [url, init] = fakeFetch.mock.calls[0];
      expect(url).toBe("http://minio.local/key?sig=xxx");
      expect(init.method).toBe("PUT");
      expect(init.body).toBe(file);
      expect(init.headers).toEqual({ "Content-Type": "text/csv" });
      // CRITICAL: no Authorization header -- this is a direct MinIO PUT.
      expect(init.headers).not.toHaveProperty("Authorization");

      vi.unstubAllGlobals();
    });

    it("throws when MinIO returns a non-OK status", async () => {
      const fakeFetch = vi
        .fn()
        .mockResolvedValueOnce({ ok: false, status: 403 });
      vi.stubGlobal("fetch", fakeFetch);

      await expect(
        putFileToPresignedURL("http://minio/k", new Blob(["x"]), "text/csv"),
      ).rejects.toThrow(/MinIO presigned PUT failed.*403/);

      vi.unstubAllGlobals();
    });
  });

  describe("useEnqueueStudentBulkImport", () => {
    it("posts {file_key} to /admin/students/bulk and returns job_id", async () => {
      mockAuthFetch.mockResolvedValueOnce({ job_id: "job-42" });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useEnqueueStudentBulkImport(), {
        wrapper,
      });

      let returned: { job_id: string } | undefined;
      await act(async () => {
        returned = await result.current.mutateAsync({
          fileKey: "student-bulk/school-1/uuid-x.csv",
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/students/bulk", {
        method: "POST",
        body: JSON.stringify({ file_key: "student-bulk/school-1/uuid-x.csv" }),
      });
      expect(returned).toEqual({ job_id: "job-42" });
    });
  });
});
