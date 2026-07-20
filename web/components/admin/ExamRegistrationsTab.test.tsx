import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { formatRupiah } from "@/lib/format";
import { ExamRegistrationsTab } from "./ExamRegistrationsTab";

// ── Mock state ──────────────────────────────────────────────────────────

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

const grantMutateSpy = vi.fn();
let grantShouldError = false;
let grantErrorMessage = "";
let grantMockResult: any = null;

let authUser: { role: string; school_id?: string } = {
  role: "admin_school",
  school_id: "school-1",
};

vi.mock("@/lib/hooks/admin-bulk-exam-orders", () => ({
  usePreviewBulkExamOrder: () => {
    const { useState } = require("react");
    const [data, setData]: [any, (v: any) => void] = useState(undefined);
    return {
      data,
      isPending: false,
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
  useCreateBulkExamOrder: () => ({
    isPending: false,
    reset: () => {},
    mutate: (input: any, opts?: any) => {
      createMutateSpy(input, opts);
      if (createShouldError) {
        opts?.onError?.(new Error(createErrorMessage));
      } else {
        opts?.onSuccess?.({ id: "order-1" });
      }
    },
  }),
}));

vi.mock("@/lib/hooks/admin-exam-grants", () => ({
  useGrantExamAccess: () => ({
    isPending: false,
    isError: false,
    reset: vi.fn(),
    mutate: (input: any, opts?: any) => {
      grantMutateSpy(input, opts);
      if (grantShouldError) {
        opts?.onError?.(new Error(grantErrorMessage));
      } else {
        opts?.onSuccess?.(grantMockResult);
      }
    },
  }),
}));

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
  useAuthStore: (sel: any) => sel({ user: authUser }),
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

import { toast } from "sonner";

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("ExamRegistrationsTab — admin_school order flow (no exam picker, exam is fixed by tab context)", () => {
  beforeEach(() => {
    authUser = { role: "admin_school", school_id: "school-1" };
    previewMutateSpy.mockClear();
    createMutateSpy.mockClear();
    previewShouldError = false;
    previewMockResult = null;
    createShouldError = false;
    createErrorMessage = "";
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("does not render a grant button for admin_school", () => {
    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });
    fireEvent.click(screen.getByTestId("participant-add"));
    expect(screen.queryByText("exam_grant_grant")).not.toBeInTheDocument();
  });

  it("previews and creates an order end-to-end, showing the real total", async () => {
    previewMockResult = {
      net_new_count: 2,
      excluded: [],
      unit_price: 75000,
      total: 150000,
    };

    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });

    fireEvent.click(screen.getByTestId("participant-add"));

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

    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });

    fireEvent.click(screen.getByTestId("participant-add"));

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

describe("ExamRegistrationsTab — super_admin grant flow (cross-school, no order/payment)", () => {
  beforeEach(() => {
    authUser = { role: "super_admin" };
    grantMutateSpy.mockClear();
    grantShouldError = false;
    grantErrorMessage = "";
    grantMockResult = null;
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("does not render the order/preview button for super_admin", () => {
    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });
    fireEvent.click(screen.getByTestId("participant-add"));
    expect(screen.queryByText("bulk_exam_order_preview")).not.toBeInTheDocument();
  });

  it("shows granted student names and usernames after a successful grant", async () => {
    grantMockResult = {
      granted_count: 2,
      granted_students: [
        { id: "s1", name: "Andi Saputra", username: "andi123" },
        { id: "s2", name: "Budi Santoso", username: "budi456" },
      ],
    };

    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });

    fireEvent.click(screen.getByTestId("participant-add"));

    const grantButton = await screen.findByText("exam_grant_grant");
    fireEvent.click(grantButton);

    expect(grantMutateSpy).toHaveBeenCalledWith(
      { exam_id: "exam-1", student_ids: ["student-1", "student-2"] },
      expect.any(Object),
    );

    await waitFor(() => {
      expect(screen.getByText("Andi Saputra")).toBeInTheDocument();
    });
    expect(screen.getByText("@andi123")).toBeInTheDocument();
    expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    expect(screen.getByText("@budi456")).toBeInTheDocument();
  });

  it("shows an error toast when the backend rejects duplicate student_ids", async () => {
    grantShouldError = true;
    grantErrorMessage = "duplicate student_id in participant selector: s1";

    render(<ExamRegistrationsTab examId="exam-1" examName="Tryout UTBK 2026" />, {
      wrapper: wrapperFactory(),
    });

    fireEvent.click(screen.getByTestId("participant-add"));

    const grantButton = await screen.findByText("exam_grant_grant");
    fireEvent.click(grantButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "duplicate student_id in participant selector: s1",
      );
    });
    expect(screen.queryByText("Andi Saputra")).not.toBeInTheDocument();
  });
});
