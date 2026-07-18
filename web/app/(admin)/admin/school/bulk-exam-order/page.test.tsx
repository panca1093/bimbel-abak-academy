import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { formatRupiah } from "@/lib/format";

// ── Mock state ──────────────────────────────────────────────────────────

const orderableExamsState: { data: { data: any[] } | null } = { data: null };
const previewMutateSpy = vi.fn();
const createMutateSpy = vi.fn();
let previewShouldError = false;
let previewMockResult:
  | {
      net_new_count: number;
      excluded: { student_id: string; name: string; reason: string }[];
      unit_price: number;
      total: number;
    }
  | null = null;
let createShouldError = false;
let createErrorMessage = "";

// usePreviewBulkExamOrder/useCreateBulkExamOrder are implemented as real
// hooks (using React's own useState) so that calling mutate() drives an
// actual re-render of the page, the same way the real react-query mutation
// hooks would — this lets tests assert on rendered DOM after an action,
// not on the mock's own call arguments.
vi.mock("@/lib/hooks/admin-bulk-exam-orders", () => {
  return {
    useOrderableExams: () => ({
      data: orderableExamsState.data,
      isLoading: false,
    }),
    usePreviewBulkExamOrder: () => {
      const { useState } = require("react");
      const [data, setData]: [any, (v: any) => void] = useState(undefined);
      const [isPending, setIsPending] = useState(false);
      return {
        data,
        isPending,
        isError: false,
        reset: () => setData(undefined),
        mutate: (input: any, opts?: any) => {
          previewMutateSpy(input, opts);
          if (previewShouldError) {
            opts?.onError?.(new Error("preview failed"));
          } else {
            setData(previewMockResult);
          }
        },
      };
    },
    useCreateBulkExamOrder: () => {
      const { useState } = require("react");
      const [isPending, setIsPending] = useState(false);
      return {
        isPending,
        reset: () => {},
        mutate: (input: any, opts?: any) => {
          createMutateSpy(input, opts);
          if (createShouldError) {
            opts?.onError?.(new Error(createErrorMessage));
          } else {
            opts?.onSuccess?.({ id: "order-1" });
          }
        },
      };
    },
  };
});

vi.mock("@/components/admin/ParticipantPicker", () => ({
  ParticipantPicker: ({ onChange }: { selected: string[]; onChange: (ids: string[]) => void }) => (
    <button
      data-testid="participant-add"
      onClick={() => onChange(["student-1", "student-2"])}
    >
      pick
    </button>
  ),
}));

vi.mock("@/components/cart/SnapCheckout", () => ({
  SnapCheckout: () => <div data-testid="snap-checkout" />,
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: (sel: any) => sel({ user: { school_id: "school-1" } }),
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

import BulkExamOrderPage from "./page";
import { toast } from "sonner";

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

function openExamSelect() {
  fireEvent.click(screen.getByRole("combobox"));
}

describe("BulkExamOrderPage — real preview + create flow", () => {
  beforeEach(() => {
    previewMutateSpy.mockClear();
    createMutateSpy.mockClear();
    previewShouldError = false;
    previewMockResult = null;
    createShouldError = false;
    createErrorMessage = "";
    orderableExamsState.data = {
      data: [{ id: "exam-1", name: "Tryout UTBK 2026" }],
    };
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("page mounts cleanly with the new hook types and renders the exam dropdown", () => {
    render(<BulkExamOrderPage />, { wrapper: wrapperFactory() });
    expect(
      screen.getByRole("heading", { name: /bulk_exam_order_title/i, level: 1 }),
    ).toBeInTheDocument();
  });

  it("renders exam options using the real Product name field (not the removed title field)", async () => {
    render(<BulkExamOrderPage />, { wrapper: wrapperFactory() });
    openExamSelect();
    await waitFor(() => {
      expect(screen.getByText("Tryout UTBK 2026")).toBeInTheDocument();
    });
  });

  it("previews and creates an order end-to-end, showing the real total", async () => {
    previewMockResult = {
      net_new_count: 2,
      excluded: [],
      unit_price: 75000,
      total: 150000,
    };

    render(<BulkExamOrderPage />, { wrapper: wrapperFactory() });

    openExamSelect();
    const examOption = await screen.findByText("Tryout UTBK 2026");
    fireEvent.click(examOption);

    const pickButton = await screen.findByTestId("participant-add");
    fireEvent.click(pickButton);

    const previewButton = await screen.findByText("bulk_exam_order_preview");
    fireEvent.click(previewButton);

    expect(previewMutateSpy).toHaveBeenCalledWith(
      { exam_id: "exam-1", student_ids: ["student-1", "student-2"] },
      expect.any(Object),
    );

    await waitFor(() => {
      expect(screen.getByText(formatRupiah(150000))).toBeInTheDocument();
    });

    const confirmButton = await screen.findByText("bulk_exam_order_confirm");
    fireEvent.click(confirmButton);

    expect(createMutateSpy).toHaveBeenCalledWith(
      { exam_id: "exam-1", student_ids: ["student-1", "student-2"] },
      expect.any(Object),
    );

    await waitFor(() => {
      expect(screen.getByText("bulk_exam_order_created")).toBeInTheDocument();
    });
    expect(screen.getByText(/Tryout UTBK 2026/)).toBeInTheDocument();
  });

  it("shows an error toast when the backend rejects duplicate participant_ids on create", async () => {
    previewMockResult = {
      net_new_count: 2,
      excluded: [],
      unit_price: 75000,
      total: 150000,
    };
    createShouldError = true;
    createErrorMessage = "duplicate student_id in participant selector: student-1";

    render(<BulkExamOrderPage />, { wrapper: wrapperFactory() });

    openExamSelect();
    const examOption = await screen.findByText("Tryout UTBK 2026");
    fireEvent.click(examOption);

    const pickButton = await screen.findByTestId("participant-add");
    fireEvent.click(pickButton);

    const previewButton = await screen.findByText("bulk_exam_order_preview");
    fireEvent.click(previewButton);

    const confirmButton = await screen.findByText("bulk_exam_order_confirm");
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "duplicate student_id in participant selector: student-1",
      );
    });
    expect(screen.queryByText("bulk_exam_order_created")).not.toBeInTheDocument();
  });
});
