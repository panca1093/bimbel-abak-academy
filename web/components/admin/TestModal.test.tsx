import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { TestModal } from "./TestModal";
import type { Test } from "@/lib/types";

const mockOnSubmit = vi.fn();
const mockOnOpenChange = vi.fn();

describe("TestModal", () => {
  beforeEach(() => {
    mockOnSubmit.mockReset();
    mockOnOpenChange.mockReset();
  });

  it("renders create modal with empty fields", () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    expect(screen.getByRole("dialog", { name: /tes baru/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/judul/i)).toHaveValue("");
    expect(screen.getByLabelText(/mata pelajaran/i)).toHaveValue("");
    expect(screen.getByLabelText(/topik/i)).toHaveValue("");
  });

  it("renders edit modal prefilled with test data", () => {
    const test: Test = {
      id: "t1",
      title: "Tryout UTBK Saintek",
      subject: "Matematika",
      topic: "Aljabar",
      duration_minutes: 90,
      audio_url: "https://cdn/audio.mp3",
      audio_play_limit: 2,
    };

    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        test={test}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    expect(screen.getByRole("dialog", { name: /sunting tes/i })).toBeInTheDocument();
    expect(screen.getByDisplayValue("Tryout UTBK Saintek")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Matematika")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Aljabar")).toBeInTheDocument();
    expect(screen.getByDisplayValue("90")).toBeInTheDocument();
    expect(screen.getByDisplayValue("https://cdn/audio.mp3")).toBeInTheDocument();
    expect(screen.getByDisplayValue("2")).toBeInTheDocument();
  });

  it("submits create input with all required fields", async () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/judul/i), { target: { value: "Tryout 1" } });
    fireEvent.input(screen.getByLabelText(/mata pelajaran/i), { target: { value: "Matematika" } });
    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Aljabar" } });
    fireEvent.input(screen.getByLabelText(/durasi/i), { target: { value: "60" } });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "Tryout 1",
          subject: "Matematika",
          topic: "Aljabar",
          duration_minutes: 60,
        })
      );
    });
  });

  it("submits edit input only with allowed fields (no title mutation in payload if untouched)", async () => {
    const test: Test = {
      id: "t1",
      title: "Tryout Lama",
      subject: "Matematika",
      topic: "Aljabar",
      duration_minutes: 60,
    };

    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        test={test}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Geometri" } });
    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({ topic: "Geometri" })
      );
    });
  });

  it("disables save when required fields are empty", () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    const saveButton = screen.getByRole("button", { name: /^simpan$/i });
    expect(saveButton).toBeDisabled();
  });

  it("disables save when duration is zero", () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/judul/i), { target: { value: "Tryout 1" } });
    fireEvent.input(screen.getByLabelText(/mata pelajaran/i), { target: { value: "Matematika" } });
    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Aljabar" } });
    fireEvent.input(screen.getByLabelText(/durasi/i), { target: { value: "0" } });

    const saveButton = screen.getByRole("button", { name: /^simpan$/i });
    expect(saveButton).toBeDisabled();
  });

  it("omits optional audio fields when blank", async () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/judul/i), { target: { value: "Tryout 1" } });
    fireEvent.input(screen.getByLabelText(/mata pelajaran/i), { target: { value: "Matematika" } });
    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Aljabar" } });
    fireEvent.input(screen.getByLabelText(/durasi/i), { target: { value: "60" } });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      const payload = mockOnSubmit.mock.calls[0][0];
      expect(payload).not.toHaveProperty("audio_url");
      expect(payload).not.toHaveProperty("audio_play_limit");
    });
  });

  it("includes audio_url and audio_play_limit when provided", async () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/judul/i), { target: { value: "Listening Test" } });
    fireEvent.input(screen.getByLabelText(/mata pelajaran/i), { target: { value: "Bahasa Inggris" } });
    fireEvent.input(screen.getByLabelText(/topik/i), { target: { value: "Listening" } });
    fireEvent.input(screen.getByLabelText(/durasi/i), { target: { value: "30" } });
    fireEvent.input(screen.getByLabelText(/url audio/i), { target: { value: "https://cdn/audio.mp3" } });
    fireEvent.input(screen.getByLabelText(/batas pemutaran audio/i), { target: { value: "3" } });

    fireEvent.click(screen.getByRole("button", { name: /^simpan$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          audio_url: "https://cdn/audio.mp3",
          audio_play_limit: 3,
        })
      );
    });
  });

  it("calls onOpenChange(false) when cancel clicked", () => {
    render(
      <TestModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: /batal/i }));
    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });
});