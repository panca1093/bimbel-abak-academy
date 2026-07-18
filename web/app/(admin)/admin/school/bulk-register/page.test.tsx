import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import BulkRegisterPage from "./page";

// ── Mutable mock state for the three bulk-upload hooks ──

const presignMutateAsync = vi.fn();
const enqueueMutateAsync = vi.fn();
const putFile = vi.fn();

const jobStatusState: {
  data: { id: string; type: string; status: string; progress: number; result_url: string | null; error: string | null; created_at: string; updated_at: string } | null;
} = {
  data: null,
};

// Re-render trigger so the test can flip the mock state and observe the page
// updating. We bump a counter inside a useState in the mock on every poll so
// the consuming component re-renders and reads the latest snapshot.
let pollTick = 0;

vi.mock("@/lib/hooks/admin-students-bulk", () => ({
  usePresignStudentBulkUpload: () => ({
    mutateAsync: presignMutateAsync,
    isPending: false,
  }),
  putFileToPresignedURL: (...args: Parameters<typeof putFile>) => putFile(...args),
  useEnqueueStudentBulkImport: () => ({
    mutateAsync: enqueueMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/lib/hooks/jobs", () => ({
  useJobStatus: () => {
    // useState would normally subscribe; with the global tick we trigger
    // re-renders manually in the test using `act()`.
    void pollTick;
    return {
      data: jobStatusState.data,
      isLoading: false,
      isError: false,
      error: null,
    };
  },
}));

vi.mock("@/stores/ui", () => ({
  useUIStore: (sel: any) => sel({ lang: "id", theme: "light", toggleTheme: vi.fn(), setLang: vi.fn() }),
}));

vi.mock("@/lib/i18n", () => ({
  useTranslation: () => ({
    lang: "id",
    t: (key: string) => key,
  }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

// Mock anchor click so we can observe the CSV download without actually triggering it.
let lastDownloadedFilename: string | null = null;
let lastDownloadedCSV: string | null = null;
let lastCapturedBlob: Blob | null = null;
const originalCreateElement = document.createElement.bind(document);

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("BulkRegisterPage", () => {
  beforeEach(() => {
    presignMutateAsync.mockReset();
    enqueueMutateAsync.mockReset();
    putFile.mockReset();
    lastDownloadedFilename = null;
    lastDownloadedCSV = null;
    lastCapturedBlob = null;
    jobStatusState.data = null;
    pollTick = 0;

    // Spy on anchor click to capture CSV download payload.
    document.createElement = ((tag: string) => {
      const el = originalCreateElement(tag);
      if (tag === "a") {
        (el as HTMLAnchorElement).click = vi.fn(function (this: HTMLAnchorElement) {
          lastDownloadedFilename = (this as HTMLAnchorElement).download;
          // Read the blob that was passed to createObjectURL (JSDOM's anchor.href
          // coerces a Blob to "[object Blob]", so we capture it at createObjectURL time).
          if (lastCapturedBlob) {
            lastCapturedBlob.text().then((t) => {
              lastDownloadedCSV = t;
            });
          }
        });
      }
      return el;
    }) as typeof document.createElement;

    // URL.createObjectURL captures the blob so the click handler can read its text.
    if (!(URL.createObjectURL as any).__mocked) {
      URL.createObjectURL = vi.fn().mockImplementation((blob: Blob) => {
        lastCapturedBlob = blob;
        return "blob:mock" as unknown as string;
      }) as typeof URL.createObjectURL;
      (URL.createObjectURL as any).__mocked = true;
    }
  });

  it("renders the page header and a Download Template button", async () => {
    render(<BulkRegisterPage />, { wrapper: wrapperFactory() });

    // The header title is rendered as h1 by AdminPageHeader.
    expect(
      screen.getByRole("heading", { name: /bulk_register_title/i, level: 1 }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /bulk_register_download_template/i }),
    ).toBeInTheDocument();
  });

  it("clicking Download Template produces a CSV with the exact header and one example row, firing no network request", async () => {
    // No mocks set up for presign/enqueue/put: any fetch they fire would be unhandled.
    render(<BulkRegisterPage />, { wrapper: wrapperFactory() });

    const downloadBtn = screen.getByRole("button", { name: /bulk_register_download_template/i });
    fireEvent.click(downloadBtn);

    await waitFor(() => expect(lastDownloadedFilename).not.toBeNull());
    await waitFor(() => expect(lastDownloadedCSV).not.toBeNull());

    // Filename ends in .csv.
    expect(lastDownloadedFilename).toMatch(/\.csv$/);
    // Body contains the locked header order plus the example row.
    expect(lastDownloadedCSV!).toMatch(
      /name,school,jenjang,provinsi,kota,kecamatan,kode_pos,email/,
    );
    // Example row is present (one illustrative row only).
    const lines = (lastDownloadedCSV ?? "").split(/\r?\n/).filter(Boolean);
    expect(lines.length).toBe(2);
    expect(lines[0]).toBe("name,school,jenjang,provinsi,kota,kecamatan,kode_pos,email");
    expect(lines[1]).toMatch(/Budi Santoso/);

    // No HTTP call fired -- presign/put/enqueue should be untouched.
    expect(presignMutateAsync).not.toHaveBeenCalled();
    expect(putFile).not.toHaveBeenCalled();
    expect(enqueueMutateAsync).not.toHaveBeenCalled();
  });

  it("uploading a valid CSV runs presign -> PUT -> enqueue in order", async () => {
    presignMutateAsync.mockResolvedValueOnce({
      url: "http://minio.local/k?sig=xyz",
      method: "PUT",
      key: "student-bulk/school-1/uuid.csv",
    });
    enqueueMutateAsync.mockResolvedValueOnce({ job_id: "job-1" });
    putFile.mockResolvedValueOnce(undefined);

    render(<BulkRegisterPage />, { wrapper: wrapperFactory() });

    const fileInput = screen.getByLabelText(/choose_file|file/i) as HTMLInputElement;
    const file = new File(["name,school,jenjang,provinsi,kota,kecamatan,kode_pos,email\nBudi,SMAN 1,sma,JB,Bandung,Coblong,40132,budi@example.com"], "students.csv", { type: "text/csv" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    const submitBtn = screen.getByRole("button", { name: /upload|import|submit|start/i });
    fireEvent.click(submitBtn);

    await waitFor(() => expect(presignMutateAsync).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(putFile).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(enqueueMutateAsync).toHaveBeenCalledTimes(1));

    // Order: presign -> put -> enqueue.
    expect(presignMutateAsync.mock.invocationCallOrder[0]).toBeLessThan(
      putFile.mock.invocationCallOrder[0],
    );
    expect(putFile.mock.invocationCallOrder[0]).toBeLessThan(
      enqueueMutateAsync.mock.invocationCallOrder[0],
    );

    // Enqueue called with the file_key returned by presign.
    expect(enqueueMutateAsync).toHaveBeenCalledWith({
      fileKey: "student-bulk/school-1/uuid.csv",
    });
  });

  it("shows terminal error when the job fails", async () => {
    presignMutateAsync.mockResolvedValueOnce({
      url: "http://minio.local/k?sig=xyz",
      method: "PUT",
      key: "k1",
    });
    enqueueMutateAsync.mockResolvedValueOnce({ job_id: "job-bad" });
    putFile.mockResolvedValueOnce(undefined);

    const { rerender } = render(<BulkRegisterPage />, { wrapper: wrapperFactory() });

    const fileInput = screen.getByLabelText(/choose_file|file/i) as HTMLInputElement;
    const file = new File(["x"], "x.csv", { type: "text/csv" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    const submitBtn = screen.getByRole("button", { name: /upload|import|submit|start/i });
    fireEvent.click(submitBtn);

    await waitFor(() => expect(enqueueMutateAsync).toHaveBeenCalled());

    // Simulate the job reaching a terminal-failed state with an error message.
    jobStatusState.data = {
      id: "job-bad",
      type: "student_bulk",
      status: "failed",
      progress: 0,
      result_url: null,
      error: "unresolvable school name 'SMAN 999'",
      created_at: "2026-07-18T00:00:00Z",
      updated_at: "2026-07-18T00:01:00Z",
    };
    pollTick++;
    rerender(<BulkRegisterPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/unresolvable school name/i),
      ).toBeInTheDocument();
    });
  });

  it("shows a download link to the result URL on terminal success", async () => {
    presignMutateAsync.mockResolvedValueOnce({
      url: "http://minio.local/k?sig=xyz",
      method: "PUT",
      key: "k2",
    });
    enqueueMutateAsync.mockResolvedValueOnce({ job_id: "job-ok" });
    putFile.mockResolvedValueOnce(undefined);

    const { rerender } = render(<BulkRegisterPage />, { wrapper: wrapperFactory() });

    const fileInput = screen.getByLabelText(/choose_file|file/i) as HTMLInputElement;
    const file = new File(["x"], "x.csv", { type: "text/csv" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    const submitBtn = screen.getByRole("button", { name: /upload|import|submit|start/i });
    fireEvent.click(submitBtn);

    await waitFor(() => expect(enqueueMutateAsync).toHaveBeenCalled());

    jobStatusState.data = {
      id: "job-ok",
      type: "student_bulk",
      status: "succeeded",
      progress: 100,
      result_url: "http://minio.local/result.csv?sig=abc",
      error: null,
      created_at: "2026-07-18T00:00:00Z",
      updated_at: "2026-07-18T00:02:00Z",
    };
    pollTick++;
    rerender(<BulkRegisterPage />);

    await waitFor(() => {
      const link = screen.getByRole("link");
      expect((link as HTMLAnchorElement).href).toBe(
        "http://minio.local/result.csv?sig=abc",
      );
    });
  });

  it("does not import any direct student-CSV service path (only HTTP hooks)", async () => {
    // Sanity: read the bundled source and grep for the service functions.
    // This is enforced by the architecture: the screen must only use the existing
    // presign/enqueue/poll HTTP surface, never call ParseStudentBulkCSV /
    // ProcessStudentBulkRows directly.
    const fs = await import("fs");
    const path = await import("path");
    const src = fs.readFileSync(
      path.join(__dirname, "page.tsx"),
      "utf8",
    );
    expect(src).not.toMatch(/ParseStudentBulkCSV/);
    expect(src).not.toMatch(/ProcessStudentBulkRows/);
  });
});
