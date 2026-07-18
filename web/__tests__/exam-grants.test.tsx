import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DICT } from "@/lib/i18n";
import { NAV_CONFIG } from "@/lib/nav-config";

// Get actual translations for assertions
const t = (key: string) => DICT.id[key as keyof typeof DICT.id] || DICT.en[key as keyof typeof DICT.en];

// ── Mock state ──────────────────────────────────────────────────────────

const orderableExamsState: { data: { data: any[] } | null } = {
  data: { data: [{ id: "exam-1", name: "Tryout UTBK 2026" }] },
};

const grantMutateSpy = vi.fn();
let grantShouldError = false;
let grantErrorMessage = "";
let grantMockResult: any = null;

vi.mock("@/lib/hooks/admin-bulk-exam-orders", () => ({
  useOrderableExams: () => ({
    data: orderableExamsState.data,
    isLoading: false,
  }),
}));

// mutate calls through to the real component's onSuccess/onError callbacks,
// same as the real POST /admin/exam-grants mutation would — so the assertions
// below exercise the page's actual handleGrant + setGrantResult code path,
// not a hand-rolled callback the test invents itself.
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
  useSearchStudentsAcrossSchools: () => ({
    data: { data: [] },
    isLoading: false,
  }),
}));

vi.mock("@/components/admin/ParticipantPicker", () => ({
  ParticipantPicker: ({ onChange }: { selected: string[]; onChange: (ids: string[]) => void }) => (
    <button
      data-testid="participant-picker-stub"
      onClick={() => onChange(["s1", "s2"])}
    >
      Pick Students
    </button>
  ),
}));

vi.mock("@/lib/i18n", async (importOriginal) => {
  const actual = (await importOriginal()) as any;
  return {
    ...actual,
    useTranslation: () => ({
      lang: "id",
      t: (key: string) => key,
    }),
  };
});

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

import ExamGrantPage from "@/app/(admin)/admin/exam-grants/page";
import { toast } from "sonner";

describe("FR-FE-13/14: exam grant screen requirements", () => {
  it("should have exam-grant navigation item in super_admin nav", () => {
    const saNav = NAV_CONFIG["super_admin"];
    const examGroup = saNav.find((g) => g.titleKey === "nav_group_exam");

    expect(examGroup).toBeDefined();
    const grantItem = examGroup!.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeDefined();
    expect(grantItem!.labelKey).toBe("nav_exam_grant");
  });

  it("should have correct translations for exam grant screen", () => {
    expect(t("exam_grant_title")).toBe("Beri Akses Ujian");
    expect(t("exam_grant_subtitle")).toBeDefined();
    expect(t("exam_grant_select_exam")).toBe("Pilih ujian");
    expect(t("exam_grant_grant")).toBe("Beri Akses");
    expect(t("exam_grant_success_title")).toBe("Akses Diberikan");
    expect(t("bulk_exam_order_pick_participants")).toBeDefined();
  });

  it("should not include exam-grant in admin_exam nav subset", () => {
    const examNav = NAV_CONFIG["admin_exam"];
    const examGroup = examNav[0];
    const grantItem = examGroup.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeUndefined();
  });

  it("should not include exam-grant in admin_school nav subset", () => {
    const schoolNav = NAV_CONFIG["admin_school"];
    const examGroup = schoolNav[0];
    const grantItem = examGroup.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeUndefined();
  });

  it("should include bulk-exam-order in admin_school nav", () => {
    const schoolNav = NAV_CONFIG["admin_school"];
    const examGroup = schoolNav[0];
    const bulkItem = examGroup.items.find((i) => i.href === "/admin/school/bulk-exam-order");
    expect(bulkItem).toBeDefined();
  });

  it("should have nav_exam_grant key in both id and en dictionaries", () => {
    const idDict = DICT.id as Record<string, string>;
    const enDict = DICT.en as Record<string, string>;

    expect(idDict["nav_exam_grant"]).toBeDefined();
    expect(enDict["nav_exam_grant"]).toBeDefined();
    expect(idDict["nav_exam_grant"]).toBe("Beri Akses Ujian");
    expect(enDict["nav_exam_grant"]).toBe("Grant Exam Access");
  });
});

describe("ExamGrantPage — real grant flow (drives the actual component, not mocked callbacks)", () => {
  beforeEach(() => {
    grantMutateSpy.mockClear();
    grantShouldError = false;
    grantErrorMessage = "";
    grantMockResult = null;
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  function openExamSelect() {
    fireEvent.click(screen.getByRole("combobox"));
  }

  it("renders exam options using the real Product name field (not the removed title field)", async () => {
    render(<ExamGrantPage />, { wrapper: wrapperFactory() });
    openExamSelect();
    await waitFor(() => {
      expect(screen.getByText("Tryout UTBK 2026")).toBeInTheDocument();
    });
  });

  it("shows granted student names and usernames after a successful grant (regression for the granted_students crash)", async () => {
    grantMockResult = {
      granted_count: 2,
      granted_students: [
        { id: "s1", name: "Andi Saputra", username: "andi123" },
        { id: "s2", name: "Budi Santoso", username: "budi456" },
      ],
    };

    render(<ExamGrantPage />, { wrapper: wrapperFactory() });

    openExamSelect();
    const examOption = await screen.findByText("Tryout UTBK 2026");
    fireEvent.click(examOption);

    const pickButton = await screen.findByTestId("participant-picker-stub");
    fireEvent.click(pickButton);

    const grantButton = await screen.findByText("exam_grant_grant");
    fireEvent.click(grantButton);

    expect(grantMutateSpy).toHaveBeenCalledWith(
      { exam_id: "exam-1", student_ids: ["s1", "s2"] },
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

    render(<ExamGrantPage />, { wrapper: wrapperFactory() });

    openExamSelect();
    const examOption = await screen.findByText("Tryout UTBK 2026");
    fireEvent.click(examOption);

    const pickButton = await screen.findByTestId("participant-picker-stub");
    fireEvent.click(pickButton);

    const grantButton = await screen.findByText("exam_grant_grant");
    fireEvent.click(grantButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "duplicate student_id in participant selector: s1",
      );
    });
    // Must not have fallen through to the success state.
    expect(screen.queryByText("Andi Saputra")).not.toBeInTheDocument();
  });
});
