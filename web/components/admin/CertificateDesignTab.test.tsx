import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { toast } from "sonner";
import { CertificateDesignTab } from "./CertificateDesignTab";
import type { CertificateDesign, ExamDetail } from "@/lib/types";

const mockFetchCertificatePreview = vi.fn();
const mockUpdateDesignMutateAsync = vi.fn();
const mockPresignMutateAsync = vi.fn();

let certificateDesignState: {
  data: CertificateDesign | undefined;
  isLoading: boolean;
  isError: boolean;
} = { data: undefined, isLoading: true, isError: false };

const sampleLayout: CertificateDesign["layout"] = {
  page: { width_mm: 297, height_mm: 210 },
  background: { kind: "builtin", ref: "classic" },
  fields: [],
};

vi.mock("@/lib/hooks/admin-exams", () => ({
  useCertificateDesign: () => certificateDesignState,
  useUpdateCertificateDesign: () => ({
    mutateAsync: mockUpdateDesignMutateAsync,
    isPending: false,
  }),
  fetchCertificatePreview: (...args: unknown[]) => mockFetchCertificatePreview(...args),
}));

vi.mock("@/lib/hooks/students", () => ({
  usePresignUpload: () => ({ mutateAsync: mockPresignMutateAsync, isPending: false }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const sampleExam: ExamDetail = {
  id: "exam-1",
  title: "UTS Matematika",
  certificate_template: "custom",
  certificate_background_key: "certificates/exam-1/old-bg.png",
  tests: [],
};

describe("CertificateDesignTab", () => {
  beforeEach(() => {
    mockFetchCertificatePreview
      .mockReset()
      .mockResolvedValue(new Blob(["pdf"], { type: "application/pdf" }));
    mockUpdateDesignMutateAsync.mockReset();
    mockPresignMutateAsync.mockReset();
    certificateDesignState = {
      data: { template: "classic", background_url: null, layout: sampleLayout },
      isLoading: false,
      isError: false,
    };
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    URL.createObjectURL = vi.fn().mockReturnValue("blob:test-url");
    URL.revokeObjectURL = vi.fn();
  });

  it("loads the initial live preview for the saved template", async () => {
    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalledWith("exam-1", "classic");
    });
  });

  it("refreshes the preview when the template is switched", async () => {
    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalledWith("exam-1", "classic");
    });

    fireEvent.click(screen.getByLabelText("Modern"));

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalledWith("exam-1", "modern");
    });
  });

  it("uploads a background then saves with the returned object key", async () => {
    mockPresignMutateAsync.mockResolvedValue({
      url: "https://upload.example.com/put-here",
      method: "PUT",
      key: "certificates/exam-1/new-bg.png",
    });
    const fetchSpy = vi.fn().mockResolvedValue({ ok: true });
    vi.stubGlobal("fetch", fetchSpy);
    mockUpdateDesignMutateAsync.mockResolvedValue({
      template: "custom",
      background_url: "https://signed.example.com/new-bg.png",
      layout: sampleLayout,
    });

    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalled();
    });

    const fileInput = screen.getByTestId("certificate-background-upload-input");
    const file = new File(["bg"], "bg.png", { type: "image/png" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(mockPresignMutateAsync).toHaveBeenCalledWith({
        filename: "bg.png",
        content_type: "image/png",
      });
    });

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        "https://upload.example.com/put-here",
        expect.objectContaining({ method: "PUT", body: file }),
      );
    });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockUpdateDesignMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          template: "classic",
          background_key: "certificates/exam-1/new-bg.png",
          layout: sampleLayout,
        }),
      );
    });

    vi.unstubAllGlobals();
  });

  it("pre-fills the background key from the exam even when the background isn't touched", async () => {
    mockUpdateDesignMutateAsync.mockResolvedValue({
      template: "classic",
      background_url: null,
      layout: sampleLayout,
    });

    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockUpdateDesignMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          background_key: "certificates/exam-1/old-bg.png",
        }),
      );
    });
  });

  it("surfaces a toast when saving fails instead of failing silently", async () => {
    mockUpdateDesignMutateAsync.mockRejectedValue(new Error("boom"));

    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
  });

  it("surfaces a toast when the upload PUT fails instead of failing silently", async () => {
    mockPresignMutateAsync.mockResolvedValue({
      url: "https://upload.example.com/put-here",
      method: "PUT",
      key: "certificates/exam-1/new-bg.png",
    });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 500 }));

    render(<CertificateDesignTab examId="exam-1" exam={sampleExam} />);

    await waitFor(() => {
      expect(mockFetchCertificatePreview).toHaveBeenCalled();
    });

    const fileInput = screen.getByTestId("certificate-background-upload-input");
    const file = new File(["bg"], "bg.png", { type: "image/png" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });

    vi.unstubAllGlobals();
  });
});
