import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";

import ExamDetailPage from "./page";
import type { RegistrationDetail } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "reg-1" }),
  useRouter: () => ({ push: vi.fn() }),
}));

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

const downloadCardMock = vi.fn();

let registrationState = {
  data: null as RegistrationDetail | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/exam", () => ({
  useRegistration: () => registrationState,
  downloadCard: (...args: unknown[]) => downloadCardMock(...args),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const sampleWithCheckin: RegistrationDetail = {
  id: "reg-1",
  student_id: "s-1",
  exam_id: "e-1",
  token: "AB12CD34",
  card_pdf_url: null,
  checked_in_at: null,
  attempts_used: 0,
  status: "registered",
  created_at: "2026-06-01T00:00:00Z",
  exam: {
    id: "e-1",
    title: "Try Out UTBK Nasional",
    scheduled_at: "2026-07-15T09:00:00Z",
    requires_checkin: true,
    check_in_window_minutes: 30,
    timer_mode: "overall",
    duration_minutes: 120,
    result_config: "{}",
  },
};

const sampleNoCheckin: RegistrationDetail = {
  ...sampleWithCheckin,
  exam: { ...sampleWithCheckin.exam, requires_checkin: false },
};

describe("ExamDetailPage", () => {
  beforeEach(() => {
    uiStore = { lang: "id" };
    downloadCardMock.mockReset();
    registrationState = {
      data: sampleWithCheckin,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("shows exam title and token masked by default (FR12)", async () => {
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Try Out UTBK Nasional/i })
      ).toBeInTheDocument();
    });

    // Token should be masked — not the raw value
    expect(screen.queryByText("AB12CD34")).not.toBeInTheDocument();
    expect(screen.getByText("••••••••")).toBeInTheDocument();
  });

  it("token toggle reveals and then re-hides the token (FR12)", async () => {
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(screen.getByText("••••••••")).toBeInTheDocument();
    });

    const showBtn = screen.getByRole("button", { name: /tampilkan token/i });
    fireEvent.click(showBtn);

    expect(screen.getByText("AB12CD34")).toBeInTheDocument();
    expect(screen.queryByText("••••••••")).not.toBeInTheDocument();

    const hideBtn = screen.getByRole("button", { name: /sembunyikan token/i });
    fireEvent.click(hideBtn);

    expect(screen.queryByText("AB12CD34")).not.toBeInTheDocument();
    expect(screen.getByText("••••••••")).toBeInTheDocument();
  });

  it("download button calls downloadCard with the registration id (FR12)", async () => {
    downloadCardMock.mockResolvedValue(undefined);
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /unduh kartu peserta/i })
      ).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /unduh kartu peserta/i }));

    await waitFor(() => {
      expect(downloadCardMock).toHaveBeenCalledWith("reg-1");
    });
  });

  it("renders check-in section when requires_checkin=true (FR12)", async () => {
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(screen.getByText(/petunjuk check-in/i)).toBeInTheDocument();
    });
  });

  it("does NOT render check-in section when requires_checkin=false (FR12)", async () => {
    registrationState = {
      ...registrationState,
      data: sampleNoCheckin,
    };
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("heading", { name: /Try Out UTBK Nasional/i })
      ).toBeInTheDocument();
    });

    expect(screen.queryByText(/petunjuk check-in/i)).not.toBeInTheDocument();
  });

  it("shows not-found message when data is null and not loading", async () => {
    registrationState = {
      data: null,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/pendaftaran tidak ditemukan/i)
      ).toBeInTheDocument();
    });
  });

  it("shows error card when isError=true", async () => {
    registrationState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("not found"),
      refetch: vi.fn(),
    };
    render(<ExamDetailPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/gagal memuat data ujian/i)
      ).toBeInTheDocument();
    });
  });
});
