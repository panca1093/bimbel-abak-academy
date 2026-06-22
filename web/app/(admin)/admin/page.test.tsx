import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import AdminIndexPage from "./page";

const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
}));

let authStore: {
  token: string | null;
  user: { role?: string; name?: string } | null;
} = {
  token: null,
  user: null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

let meState: {
  data: { role?: string; name?: string } | null;
  isError: boolean;
  isLoading: boolean;
} = { data: null, isError: false, isLoading: false };

vi.mock("@/lib/hooks/auth", async () => {
  const actual = await vi.importActual<typeof import("@/lib/hooks/auth")>(
    "@/lib/hooks/auth"
  );
  return {
    ...actual,
    useMe: ({ enabled }: { enabled?: boolean }) =>
      enabled
        ? meState
        : { data: null, isError: false, isLoading: false },
  };
});

describe("AdminIndexPage", () => {
  beforeEach(() => {
    replace.mockClear();
    authStore = { token: null, user: null };
    meState = { data: null, isError: false, isLoading: false };
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("renders dashboard for super_admin role", () => {
    authStore = { token: "t", user: { role: "super_admin", name: "Super Admin" } };
    const { getByText, container } = render(<AdminIndexPage />);
    expect(getByText("Super Admin · Abak Academy")).toBeInTheDocument();
    const headings = container.querySelectorAll("h1");
    expect(headings.length).toBeGreaterThan(0);
  });

  it("renders the hero band description for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(
      getByText(
        "Akses penuh ke semua domain. Pantau seluruh platform dari satu tempat."
      )
    ).toBeInTheDocument();
  });

  it("renders stat cards for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Pendapatan Bulan Ini")).toBeInTheDocument();
    expect(getByText("Total Siswa")).toBeInTheDocument();
    expect(getByText("Sesi Ujian Aktif")).toBeInTheDocument();
    expect(getByText("Jumlah Sekolah")).toBeInTheDocument();
  });

  it("renders audit log section for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Log Aktivitas")).toBeInTheDocument();
    expect(getByText("Lihat Semua")).toBeInTheDocument();
    expect(getByText("Rina Wijayanti")).toBeInTheDocument();
    expect(getByText("Hendra Gunawan")).toBeInTheDocument();
    expect(getByText("Sri Wahyuni")).toBeInTheDocument();
  });

  it("renders quick actions section for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Akses Cepat")).toBeInTheDocument();
    expect(getByText("Buat Soal Baru")).toBeInTheDocument();
    expect(getByText("Tambah Produk")).toBeInTheDocument();
    expect(getByText("Daftarkan Siswa")).toBeInTheDocument();
    expect(getByText("Laporan Penjualan")).toBeInTheDocument();
  });

  it("renders the Shield icon (hero band) for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { container } = render(<AdminIndexPage />);
    // Shield icon renders an SVG
    const svgs = container.querySelectorAll("svg");
    expect(svgs.length).toBeGreaterThan(0);
  });

  it("returns null for admin_store role", () => {
    authStore = { token: "t", user: { role: "admin_store" } };
    const { container } = render(<AdminIndexPage />);
    expect(container.innerHTML).toBe("");
  });

  it("returns null for admin_exam role", async () => {
    authStore = { token: "t", user: { role: "admin_exam" } };
    const { container } = render(<AdminIndexPage />);
    await waitFor(() => expect(replace).toHaveBeenCalled());
    expect(container.innerHTML).toBe("");
  });

  it("redirects non-super_admin roles via router.replace", async () => {
    authStore = { token: "t", user: { role: "admin_store" } };
    render(<AdminIndexPage />);
    await waitFor(() => expect(replace).toHaveBeenCalledWith("/admin/store"));
  });

  it("uses name from store when available", () => {
    authStore = {
      token: "t",
      user: { role: "super_admin", name: "Budi Santoso" },
    };
    const { container } = render(<AdminIndexPage />);
    expect(container.textContent).toContain("Budi Santoso");
  });

  it("falls back to 'Super Admin' when name is missing", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { container } = render(<AdminIndexPage />);
    expect(container.textContent).toContain("Super Admin");
  });
});
