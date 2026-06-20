import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor, fireEvent, within } from "@testing-library/react";
import ExamBanksPage from "./page";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("ExamBanksPage", () => {
  it("renders the question bank title and action buttons", async () => {
    render(<ExamBanksPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Bank Soal/i })
      ).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: /Kelola topik/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Buat$/i })).toBeInTheDocument();
  });

  it("lists all mock questions by default", async () => {
    render(<ExamBanksPage />);

    await waitFor(() => {
      expect(screen.getByText("AQ-1041")).toBeInTheDocument();
      expect(screen.getByText("AQ-1204")).toBeInTheDocument();
    });
  });

  it("filters rows by format chips", async () => {
    render(<ExamBanksPage />);

    await waitFor(() => expect(screen.getByText("AQ-1041")).toBeInTheDocument());

    const essayChip = screen.getByRole("button", { name: /^Esai$/i });
    fireEvent.click(essayChip);

    expect(screen.queryByText("AQ-1041")).not.toBeInTheDocument();
    expect(screen.getByText("AQ-1204")).toBeInTheDocument();

    const allChip = screen.getByRole("button", { name: /^Semua$/i });
    fireEvent.click(allChip);

    expect(screen.getByText("AQ-1041")).toBeInTheDocument();
    expect(screen.getByText("AQ-1204")).toBeInTheDocument();
  });

  it("opens the manage topics dialog", async () => {
    render(<ExamBanksPage />);

    await waitFor(() =>
      expect(screen.getByRole("button", { name: /Kelola topik/i })).toBeInTheDocument()
    );

    fireEvent.click(screen.getByRole("button", { name: /Kelola topik/i }));

    const dialog = screen.getByRole("dialog", { name: /Kelola topik/i });
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText("Aljabar")).toBeInTheDocument();
  });

  it("opens the create question dialog", async () => {
    render(<ExamBanksPage />);

    await waitFor(() =>
      expect(screen.getByRole("button", { name: /^Buat$/i })).toBeInTheDocument()
    );

    fireEvent.click(screen.getByRole("button", { name: /^Buat$/i }));

    expect(screen.getByRole("dialog", { name: /Buat soal/i })).toBeInTheDocument();
  });
});
