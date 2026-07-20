import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import ExamPage from "./page";
import type { RegistrationListItem } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: toastSuccess, error: toastError },
}));

const downloadCardMock = vi.fn();
const checkInMutate = vi.fn();
const startSessionMutate = vi.fn();

let registrationsState = {
  data: null as RegistrationListItem[] | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/exam", () => ({
  useRegistrations: () => registrationsState,
  downloadCard: (...args: unknown[]) => downloadCardMock(...args),
  useCheckIn: () => ({ mutate: checkInMutate, isPending: false }),
  useStartSession: () => ({ mutateAsync: startSessionMutate, isPending: false }),
}));

const freeUpcoming: RegistrationListItem = {
  id: "reg-1",
  student_id: "s-1",
  exam_id: "e-1",
  token: "ABC12345",
  card_pdf_url: null,
  checked_in_at: null,
  attempts_used: 0,
  status: "registered",
  created_at: "2026-06-01T00:00:00Z",
  exam_title: "Try Out UTBK Gratis #12",
  scheduled_at: "2026-07-15T09:00:00Z",
  is_free: true,
  requires_checkin: false,
  check_in_window_minutes: null,
  duration_minutes: 90,
};

const paidNoSchedule: RegistrationListItem = {
  id: "reg-2",
  student_id: "s-1",
  exam_id: "e-2",
  token: "XYZ98765",
  card_pdf_url: null,
  checked_in_at: null,
  attempts_used: 0,
  status: "registered",
  created_at: "2026-06-02T00:00:00Z",
  exam_title: "Ujian Akhir Matematika",
  scheduled_at: null,
  is_free: false,
  requires_checkin: true,
  check_in_window_minutes: 15,
  duration_minutes: 60,
};

const sample: RegistrationListItem[] = [freeUpcoming, paidNoSchedule];

describe("ExamPage", () => {
  beforeEach(() => {
    uiStore = { lang: "id" };
    registrationsState = {
      data: sample,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("renders the competition list with exam titles grouped by free/paid", async () => {
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText("Try Out UTBK Gratis #12")).toBeInTheDocument();
      expect(screen.getByText("Ujian Akhir Matematika")).toBeInTheDocument();
    });
    expect(screen.getByText("Paket gratis")).toBeInTheDocument();
    expect(screen.getByText("Paket saya")).toBeInTheDocument();
  });

  it("translates copy when language is en", () => {
    uiStore = { lang: "en" };
    render(<ExamPage />);

    expect(
      screen.getByRole("heading", { name: /Competition & Tryout/i })
    ).toBeInTheDocument();
  });

  it("shows a start button for a free exam with no check-in requirement", async () => {
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText("Try Out UTBK Gratis #12")).toBeInTheDocument();
    });
    expect(screen.getAllByRole("button", { name: /Mulai/i }).length).toBeGreaterThan(0);
  });

  it("shows the check-in token input when no schedule is set but check-in is required", async () => {
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText("Ujian Akhir Matematika")).toBeInTheDocument();
    });
    expect(screen.getByPlaceholderText("Token dari kartu ujian")).toBeInTheDocument();
  });

  it("shows an empty state when there are no registrations", async () => {
    registrationsState = {
      data: [],
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText("Anda belum terdaftar pada ujian apapun")).toBeInTheDocument();
    });
  });

  it("shows an error state with a retry action", async () => {
    const refetch = vi.fn();
    registrationsState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("network down"),
      refetch,
    };
    render(<ExamPage />);

    await waitFor(() => {
      expect(screen.getByText(/Gagal memuat data ujian/)).toBeInTheDocument();
    });
    screen.getByRole("button", { name: "Coba lagi" }).click();
    expect(refetch).toHaveBeenCalled();
  });
});
