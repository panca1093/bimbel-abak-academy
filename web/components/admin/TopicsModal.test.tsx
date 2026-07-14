import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TopicsModal } from "./TopicsModal";
import type { ExamTopic } from "@/lib/types";

const mockCreateAsync = vi.fn();
const mockDeleteAsync = vi.fn();

let topicsState = {
  data: null as { data: ExamTopic[]; next_cursor?: string } | null,
  isLoading: false,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

let createState = { mutateAsync: mockCreateAsync, isPending: false };
let deleteState = { mutateAsync: mockDeleteAsync, isPending: false };

vi.mock("@/lib/hooks/admin-topics", () => ({
  useTopics: () => topicsState,
  useCreateTopic: () => createState,
  useDeleteTopic: () => deleteState,
}));

vi.mock("@/lib/i18n", () => ({
  useTranslation: () => ({ t: (k: string) => k, lang: "id" }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { toast } from "sonner";

const sampleTopics: ExamTopic[] = [
  { id: "topic-1", name: "Algebra", subject: "Math", question_count: 3 },
  { id: "topic-2", name: "Geometry", subject: "Math", question_count: 0 },
];

function renderWithClient(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe("TopicsModal", () => {
  beforeEach(() => {
    topicsState = {
      data: { data: sampleTopics },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    createState = { mutateAsync: mockCreateAsync, isPending: false };
    deleteState = { mutateAsync: mockDeleteAsync, isPending: false };
    mockCreateAsync.mockReset();
    mockDeleteAsync.mockReset();
    (toast.success as ReturnType<typeof vi.fn>).mockReset();
    (toast.error as ReturnType<typeof vi.fn>).mockReset();
  });

  it("lists topics with name, subject and question count", () => {
    renderWithClient(<TopicsModal open={true} onOpenChange={vi.fn()} />);

    expect(screen.getByText("Algebra")).toBeInTheDocument();
    expect(screen.getByText("Geometry")).toBeInTheDocument();
    expect(screen.getByText(/Math · 3 questions/i)).toBeInTheDocument();
    expect(screen.getByText(/Math · 0 questions/i)).toBeInTheDocument();
  });

  it("creates a topic from the form", async () => {
    mockCreateAsync.mockResolvedValueOnce({ id: "topic-3", name: "Calculus", subject: "Math" });

    renderWithClient(<TopicsModal open={true} onOpenChange={vi.fn()} />);

    fireEvent.input(screen.getByLabelText(/topic_name/i), { target: { value: "Calculus" } });
    fireEvent.input(screen.getByLabelText(/subject/i), { target: { value: "Math" } });
    fireEvent.click(screen.getByRole("button", { name: /add_topic/i }));

    await waitFor(() => {
      expect(mockCreateAsync).toHaveBeenCalledWith({ name: "Calculus", subject: "Math" });
      expect(toast.success).toHaveBeenCalledWith("changes_saved");
    });
  });

  it("disables delete for topics with questions", () => {
    renderWithClient(<TopicsModal open={true} onOpenChange={vi.fn()} />);

    const rows = screen.getAllByText(/Algebra|Geometry/).map((el) => el.closest("div")!);
    // Find delete buttons in the list rows
    const deleteButtons = screen.getAllByRole("button", { name: /action_delete/i });
    expect(deleteButtons[0]).toBeDisabled();
    expect(deleteButtons[1]).not.toBeDisabled();
  });

  it("deletes an unreferenced topic and shows success", async () => {
    mockDeleteAsync.mockResolvedValueOnce(undefined);

    renderWithClient(<TopicsModal open={true} onOpenChange={vi.fn()} />);

    const deleteButtons = screen.getAllByRole("button", { name: /action_delete/i });
    fireEvent.click(deleteButtons[1]);

    await waitFor(() => {
      expect(mockDeleteAsync).toHaveBeenCalledWith("topic-2");
      expect(toast.success).toHaveBeenCalledWith("changes_saved");
    });
  });

  it("does not call delete for referenced topics because the button is disabled", async () => {
    renderWithClient(<TopicsModal open={true} onOpenChange={vi.fn()} />);

    const deleteButtons = screen.getAllByRole("button", { name: /action_delete/i });
    fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(mockDeleteAsync).not.toHaveBeenCalled();
    });
  });
});
