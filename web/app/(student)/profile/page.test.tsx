import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
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
  useSchools: () => ({ data: null, isLoading: false }),
  usePresignUpload: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdatePhoto: () => ({ mutateAsync: vi.fn(), isPending: false }),
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

  it("shows a read-only profile with an edit trigger and no save button by default", async () => {
    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByLabelText(/nama/i, { selector: "input" })).toBeInTheDocument();
    });

    expect(screen.getByLabelText(/nama/i, { selector: "input" })).toBeDisabled();
    expect(screen.getByLabelText(/telepon|phone/i, { selector: "input" })).toBeDisabled();
    expect(screen.getByLabelText(/alamat|address/i, { selector: "input" })).toBeDisabled();
    expect(screen.getByRole("button", { name: /ubah profil|edit profile/i })).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /simpan perubahan|save changes/i })
    ).not.toBeInTheDocument();
  });

  it("entering edit mode unlocks fields and does not save on its own", async () => {
    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /ubah profil|edit profile/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /ubah profil|edit profile/i }));

    // entering edit mode must never trigger a save (regression: edit click used to PATCH)
    expect(mutateMock).not.toHaveBeenCalled();
    expect(screen.getByLabelText(/nama/i, { selector: "input" })).toBeEnabled();
    expect(screen.getByLabelText(/telepon|phone/i, { selector: "input" })).toBeEnabled();
    expect(
      screen.getByRole("button", { name: /simpan perubahan|save changes/i })
    ).toBeInTheDocument();
    // email stays locked even in edit mode
    expect(screen.getByLabelText(/email/i, { selector: "input" })).toBeDisabled();
  });
});
