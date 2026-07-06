import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { ExamModal } from "./ExamModal";
import type { ExamListItem } from "@/lib/types";

const mockCreateExam = vi.fn();
const mockUpdateExam = vi.fn();

vi.mock("@/lib/hooks/admin-exams", () => ({
  useCreateExam: () => ({ mutateAsync: mockCreateExam, isPending: false }),
  useUpdateExam: () => ({ mutateAsync: mockUpdateExam, isPending: false }),
  fetchCertificatePreview: vi.fn(),
}));

const sampleExam: ExamListItem = {
  id: "exam-1",
  title: "UTS Matematika",
  product_price: 0,
  product_status: "draft",
  certificate_template: "modern",
  timer_mode: "overall",
  duration_minutes: 90,
};

describe("ExamModal", () => {
  beforeEach(() => {
    mockCreateExam.mockReset();
    mockUpdateExam.mockReset();
  });

  it("pre-fills certificate template from exam data on edit", async () => {
    render(
      <ExamModal open={true} onClose={vi.fn()} exam={sampleExam} onSaved={vi.fn()} />,
    );

    await waitFor(() => {
      const classic = screen.getByRole("radio", { name: "Klasik" });
      const modern = screen.getByRole("radio", { name: "Modern" });
      const elegant = screen.getByRole("radio", { name: "Elegan" });
      expect(classic).not.toBeChecked();
      expect(modern).toBeChecked();
      expect(elegant).not.toBeChecked();
    });
  });

  it("submitted payload includes certificate_template", async () => {
    mockUpdateExam.mockResolvedValue({ id: "exam-1", title: "UTS Matematika" });

    render(
      <ExamModal open={true} onClose={vi.fn()} exam={sampleExam} onSaved={vi.fn()} />,
    );

    await waitFor(() => {
      expect(screen.getByRole("radio", { name: "Modern" })).toBeChecked();
    });

    const titleInput = screen.getByLabelText(/judul/i);
    fireEvent.input(titleInput, { target: { value: "UTS Matematika Updated" } });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockUpdateExam).toHaveBeenCalledWith(
        expect.objectContaining({ certificate_template: "modern", title: "UTS Matematika Updated" }),
      );
    });
  });

  it("preview button disabled on create, enabled on edit", async () => {
    // Create mode (no exam)
    const { unmount } = render(
      <ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />,
    );

    expect(screen.getByRole("button", { name: "Pratinjau Sertifikat" })).toBeDisabled();
    unmount();

    // Edit mode (with exam)
    render(
      <ExamModal open={true} onClose={vi.fn()} exam={sampleExam} onSaved={vi.fn()} />,
    );

    // Wait for effect to populate fields
    await waitFor(() => {
      expect(screen.getByRole("radio", { name: "Modern" })).toBeChecked();
    });

    expect(screen.getByRole("button", { name: "Pratinjau Sertifikat" })).toBeEnabled();
  });
});
