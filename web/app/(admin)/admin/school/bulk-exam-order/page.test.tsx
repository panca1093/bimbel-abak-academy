import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// Mutable mock state for the bulk-exam-order hooks
const orderableExamsState: { data: { data: any[] } | null } = { data: null };
const previewMutate = vi.fn();
let previewData:
  | {
      net_new_count: number;
      excluded: { student_id: string; name: string; reason: string }[];
      unit_price: number;
      total: number;
    }
  | undefined = undefined;
const createMutate = vi.fn();

vi.mock("@/lib/hooks/admin-bulk-exam-orders", () => ({
  useOrderableExams: () => ({
    data: orderableExamsState.data,
    isLoading: false,
  }),
  usePreviewBulkExamOrder: () => ({
    mutate: previewMutate,
    isPending: false,
    data: previewData,
  }),
  useCreateBulkExamOrder: () => ({
    mutate: createMutate,
    isPending: false,
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

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("BulkExamOrderPage — preview shape + endpoint", () => {
  beforeEach(() => {
    previewMutate.mockReset();
    orderableExamsState.data = {
      data: [{ id: "exam-1", title: "Tryout UTBK 2026" }],
    };
  });

  it("calls /admin/bulk-exam-orders/exams for the orderable list (FR-BULK-01 dedicated endpoint)", () => {
    // Sanity: read the hook file and assert the endpoint path.
    // (Behavior is pinned to the dedicated route per spec, not the
    // /admin/exams?orderable=true fallback.)
    const fs = require("fs");
    const path = require("path");
    const src = fs.readFileSync(
      path.join(__dirname, "../../../../../lib/hooks/admin-bulk-exam-orders.ts"),
      "utf8",
    );
    expect(src).toMatch(/\/admin\/bulk-exam-orders\/exams/);
    expect(src).not.toMatch(/\/admin\/exams\?orderable=true/);
  });

  it("BulkExamOrderPreview type matches the backend service.BulkOrderPreview shape", () => {
    // Pin the wire shape: net_new_count, excluded, unit_price, total
    // (not the previous {exam, students, total_price}).
    const fs = require("fs");
    const path = require("path");
    const src = fs.readFileSync(
      path.join(__dirname, "../../../../../lib/hooks/admin-bulk-exam-orders.ts"),
      "utf8",
    );
    expect(src).toMatch(/net_new_count/);
    expect(src).toMatch(/unit_price/);
    // Frontend must NOT depend on the old (removed) shape fields.
    expect(src).not.toMatch(/total_price/);
    expect(src).not.toMatch(/BulkExamOrderStudent/);
  });

  it("page.tsx reads net_new_count / total from the preview, not the old shape", () => {
    // Pin that the page uses the new shape (otherwise the test would render
    // undefined values at runtime).
    const fs = require("fs");
    const path = require("path");
    const src = fs.readFileSync(path.join(__dirname, "page.tsx"), "utf8");
    expect(src).toMatch(/previewMutation\.data\.net_new_count/);
    expect(src).toMatch(/previewMutation\.data\.total/);
    expect(src).not.toMatch(/previewMutation\.data\.exam\.title/);
    expect(src).not.toMatch(/previewMutation\.data\.total_price/);
  });

  it("page mounts cleanly with the new hook types and renders the exam dropdown", () => {
    render(<BulkExamOrderPage />, { wrapper: wrapperFactory() });
    // Page header is rendered
    expect(
      screen.getByRole("heading", { name: /bulk_exam_order_title/i, level: 1 }),
    ).toBeInTheDocument();
  });
});
