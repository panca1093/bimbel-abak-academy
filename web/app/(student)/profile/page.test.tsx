import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ProfilePage from "./page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
}));

let authStore = {
  user: null as { name?: string } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

let profileState = {
  data: null as {
    name?: string;
    email?: string;
    phone?: string;
    nis?: string;
    grade?: string;
    target_exam?: string;
    alamat_domisili?: string;
  } | null,
  isLoading: false,
  isError: false,
  error: null,
  refetch: vi.fn(),
};

const mutateMock = vi.fn();

vi.mock("@/lib/hooks/students", () => ({
  useDashboard: () => ({ data: null, isLoading: false, isError: false, refetch: vi.fn() }),
  useProfile: () => profileState,
  useUpdateProfile: () => ({ mutate: mutateMock, isPending: false }),
  useChangePassword: () => ({ mutate: vi.fn(), isPending: false }),
}));


vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

describe("ProfilePage", () => {
  beforeEach(() => {
    authStore = { user: { name: "Budi Santoso" } };
    profileState = {
      data: {
        name: "Budi Santoso",
        email: "budi@example.com",
        phone: "08123456789",
        nis: "1928374650",
        grade: "12 SMA",
        target_exam: "SNBT",
        alamat_domisili: "Jl. Merdeka No. 1",
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    mutateMock.mockClear();
  });

  it("renders locked read-only profile fields returned by the API", async () => {
    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByLabelText(/nama/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/nomor telepon|phone/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/nis/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/kelas|grade/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/target ujian|target exam/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/alamat domisili|address/i)).toBeInTheDocument();
    });
  });

  it("locks extended fields so they cannot be edited", async () => {
    render(<ProfilePage />);

    await waitFor(() => {
      const lockedLabels = [
        screen.getByLabelText(/nomor telepon|phone/i),
        screen.getByLabelText(/nis/i),
        screen.getByLabelText(/kelas|grade/i),
        screen.getByLabelText(/target ujian|target exam/i),
        screen.getByLabelText(/alamat domisili|address/i),
      ];
      for (const input of lockedLabels) {
        expect(input).toBeDisabled();
      }
    });
  });
});
