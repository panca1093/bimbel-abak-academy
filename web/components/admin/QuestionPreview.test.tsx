import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { QuestionPreview } from "./QuestionPreview";
import type { BankQuestionListItem } from "@/lib/types";

vi.mock("@/lib/i18n", () => ({
  useTranslation: () => ({ t: (k: string) => k, lang: "id" }),
}));

const sampleItem: BankQuestionListItem = {
  question: {
    id: "q1",
    format: "mcq",
    body: "What is 2+2?",
    difficulty: "easy",
    point_correct: 1,
    point_wrong: 0,
    sort_order: 1,
    topic_id: "topic-1",
    topic: "Arithmetic",
  },
  options: [
    { question_id: "q1", key: "a", text: "3", is_correct: false, sort_order: 1 },
    { question_id: "q1", key: "b", text: "4", is_correct: true, sort_order: 2 },
  ],
  attached_count: 2,
};

function renderWithClient(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe("QuestionPreview", () => {
  const onOpenChange = vi.fn();
  const onEdit = vi.fn();

  beforeEach(() => {
    onOpenChange.mockReset();
    onEdit.mockReset();
  });

  it("renders read-only question details and options", () => {
    renderWithClient(
      <QuestionPreview
        item={sampleItem}
        open={true}
        onOpenChange={onOpenChange}
        onEdit={onEdit}
      />
    );

    expect(screen.getByText("What is 2+2?")).toBeInTheDocument();
    expect(screen.getByText("Arithmetic")).toBeInTheDocument();
    expect(screen.getByText("4")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("calls onEdit when Edit is clicked", () => {
    renderWithClient(
      <QuestionPreview
        item={sampleItem}
        open={true}
        onOpenChange={onOpenChange}
        onEdit={onEdit}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: /action_edit/i }));
    expect(onEdit).toHaveBeenCalled();
  });

  it("returns null when item is not provided", () => {
    const { container } = renderWithClient(
      <QuestionPreview
        item={null}
        open={true}
        onOpenChange={onOpenChange}
        onEdit={onEdit}
      />
    );
    expect(container.firstChild).toBeNull();
  });
});
