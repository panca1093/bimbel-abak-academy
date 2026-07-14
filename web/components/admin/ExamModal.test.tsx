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
  certificate_template: "modern",
  timer_mode: "overall",
  duration_minutes: 90,
  result_config: "score_only",
  result_release_at: "2026-07-02T10:00:00Z",
  check_in_window_minutes: 15,
  grace_window_minutes: 5,
  max_attempts: 2,
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

  it("renders a mode selector with standard, utbk, ielts options on create", () => {
    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    expect(screen.getByLabelText("Standar")).toBeInTheDocument();
    expect(screen.getByLabelText("UTBK")).toBeInTheDocument();
    expect(screen.getByLabelText("IELTS")).toBeInTheDocument();
  });

  it("defaults to standard mode on create", () => {
    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    expect(screen.getByLabelText("Standar")).toBeChecked();
    expect(screen.getByLabelText("UTBK")).not.toBeChecked();
    expect(screen.getByLabelText("IELTS")).not.toBeChecked();
  });

  it("includes mode in create payload", async () => {
    mockCreateExam.mockResolvedValue({ id: "exam-1" });

    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    fireEvent.input(screen.getByLabelText(/judul/i), {
      target: { value: "UTBK Tryout" },
    });

    // Use per_test timer so overall duration isn't required
    fireEvent.click(screen.getByLabelText("Per Tes"));

    // Select UTBK mode
    fireEvent.click(screen.getByLabelText("UTBK"));

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockCreateExam).toHaveBeenCalledWith(
        expect.objectContaining({ mode: "utbk", title: "UTBK Tryout" }),
      );
    });
  });

  it("includes mode in update payload", async () => {
    mockUpdateExam.mockResolvedValue({ id: "exam-1", title: "Updated" });

    render(
      <ExamModal
        open={true}
        onClose={vi.fn()}
        exam={{ ...sampleExam, mode: "utbk" }}
        onSaved={vi.fn()}
      />,
    );

    await waitFor(() => {
      expect(screen.getByLabelText("UTBK")).toBeChecked();
    });

    fireEvent.input(screen.getByLabelText(/judul/i), {
      target: { value: "Updated Title" },
    });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockUpdateExam).toHaveBeenCalledWith(
        expect.objectContaining({ mode: "utbk", title: "Updated Title" }),
      );
    });
  });

  it("shows a hint when utbk or ielts mode is selected", () => {
    const { rerender } = render(
      <ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />,
    );

    // Default: no hint
    expect(
      screen.queryByText(
        "Setiap tes terlampir akan menjadi sesi dengan timer tersendiri.",
      ),
    ).not.toBeInTheDocument();

    // Select UTBK → hint appears
    fireEvent.click(screen.getByLabelText("UTBK"));
    expect(
      screen.getByText(
        "Setiap tes terlampir akan menjadi sesi dengan timer tersendiri.",
      ),
    ).toBeInTheDocument();

    // Switch back to Standard → hint disappears
    fireEvent.click(screen.getByLabelText("Standar"));
    expect(
      screen.queryByText(
        "Setiap tes terlampir akan menjadi sesi dengan timer tersendiri.",
      ),
    ).not.toBeInTheDocument();
  });

  it("pre-fills mode from exam data on edit", async () => {
    const examWithMode: ExamListItem = {
      ...sampleExam,
      mode: "ielts",
    };

    render(
      <ExamModal
        open={true}
        onClose={vi.fn()}
        exam={examWithMode}
        onSaved={vi.fn()}
      />,
    );

    await waitFor(() => {
      expect(screen.getByLabelText("IELTS")).toBeChecked();
      expect(screen.getByLabelText("Standar")).not.toBeChecked();
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

  it("pre-fills extended package fields from exam data on edit", async () => {
    render(
      <ExamModal open={true} onClose={vi.fn()} exam={sampleExam} onSaved={vi.fn()} />,
    );

    await waitFor(() => {
      expect(
        (screen.getByLabelText(/konfigurasi hasil/i) as HTMLSelectElement).value,
      ).toBe("score_only");
    });

    expect(
      (screen.getByLabelText(/jendela check-in/i) as HTMLInputElement).value,
    ).toBe("15");
    expect(
      (screen.getByLabelText(/jendela toleransi/i) as HTMLInputElement).value,
    ).toBe("5");
    expect(
      (screen.getByLabelText(/maks\. percobaan/i) as HTMLInputElement).value,
    ).toBe("2");
  });

  it("submitted update payload includes extended package fields", async () => {
    mockUpdateExam.mockResolvedValue({ id: "exam-1", title: "UTS Matematika" });

    render(
      <ExamModal open={true} onClose={vi.fn()} exam={sampleExam} onSaved={vi.fn()} />,
    );

    await waitFor(() => {
      expect(screen.getByLabelText(/judul/i)).toBeInTheDocument();
    });

    fireEvent.input(screen.getByLabelText(/judul/i), {
      target: { value: "UTS Matematika Updated" },
    });
    fireEvent.change(screen.getByLabelText(/konfigurasi hasil/i), {
      target: { value: "score_pembahasan" },
    });
    fireEvent.input(screen.getByLabelText(/jendela check-in/i), {
      target: { value: "30" },
    });
    fireEvent.input(screen.getByLabelText(/jendela toleransi/i), {
      target: { value: "10" },
    });
    fireEvent.input(screen.getByLabelText(/maks\. percobaan/i), {
      target: { value: "3" },
    });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockUpdateExam).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "UTS Matematika Updated",
          result_config: "score_pembahasan",
          check_in_window_minutes: 30,
          grace_window_minutes: 10,
          max_attempts: 3,
        }),
      );
    });
  });

  it("number window and attempts inputs reject negatives via min attribute", () => {
    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    expect(screen.getByLabelText(/jendela check-in/i)).toHaveAttribute("min", "0");
    expect(screen.getByLabelText(/jendela toleransi/i)).toHaveAttribute("min", "0");
    expect(screen.getByLabelText(/maks\. percobaan/i)).toHaveAttribute("min", "0");
  });

  it("omits empty window and attempts from create payload", async () => {
    mockCreateExam.mockResolvedValue({ id: "exam-1" });

    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    fireEvent.input(screen.getByLabelText(/judul/i), {
      target: { value: "Paket Minimal" },
    });
    fireEvent.click(screen.getByLabelText("Per Tes"));

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockCreateExam).toHaveBeenCalledWith(
        expect.objectContaining({
          check_in_window_minutes: null,
          grace_window_minutes: null,
          max_attempts: null,
        }),
      );
    });
  });

  it("result_config is constrained to allowed options", () => {
    render(<ExamModal open={true} onClose={vi.fn()} onSaved={vi.fn()} />);

    const select = screen.getByLabelText(/konfigurasi hasil/i) as HTMLSelectElement;
    const options = Array.from(select.options).map((o) => o.value);
    expect(options).toContain("hidden");
    expect(options).toContain("score_only");
    expect(options).toContain("score_pembahasan");
  });
});
