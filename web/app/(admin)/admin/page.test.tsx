import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import AdminIndexPage from "./page";

const replace = vi.fn();
const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace, push }),
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

// Shared state for admin hooks — each test reassigns before render
let auditState: {
  data: { id: number; actor_name?: string | null; actor_id?: string | null; actor_email?: string | null; target_type: string; target_id: string; action: string; created_at: string }[];
  isLoading: boolean;
  isError: boolean;
  refetch: ReturnType<typeof vi.fn>;
} = {
  data: [],
  isLoading: true,
  isError: false,
  refetch: vi.fn(),
};

let revenueState: {
  data: { total: number } | null;
  isLoading: boolean;
} = {
  data: null,
  isLoading: true,
};

let schoolsState: {
  data: { id: string; name: string; code: string; student_count?: number }[] | null;
  isLoading: boolean;
} = {
  data: null,
  isLoading: true,
};

vi.mock("@/lib/hooks/admin-audit", () => ({
  useAdminAuditLog: () => auditState,
}));

vi.mock("@/lib/hooks/admin-revenue", () => ({
  useAdminRevenue: () => revenueState,
}));

vi.mock("@/lib/hooks/students", () => ({
  useSchools: () => schoolsState,
}));

const sampleAuditEntries = [
  { id: 1, actor_name: "Rina Wijayanti", target_type: "Product", target_id: "P-101", action: "Menambahkan produk baru", created_at: new Date(Date.now() - 120_000).toISOString() },
  { id: 2, actor_name: "Hendra Gunawan", target_type: "Order", target_id: "ORD-204", action: "Mengubah status pesanan", created_at: new Date(Date.now() - 7200_000).toISOString() },
  { id: 3, actor_name: "Sri Wahyuni", target_type: "Course", target_id: "C-55", action: "Memperbarui konten kursus", created_at: new Date(Date.now() - 86400_000).toISOString() },
] as const;

describe("AdminIndexPage", () => {
  beforeEach(() => {
    replace.mockClear();
    push.mockClear();
    authStore = { token: null, user: null };
    meState = { data: null, isError: false, isLoading: false };
    auditState = {
      data: sampleAuditEntries as unknown as typeof auditState.data,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    };
    revenueState = { data: { total: 15_000_000 }, isLoading: false };
    schoolsState = {
      data: [
        { id: "s1", name: "SMA N 1 Jakarta", code: "SMA01", student_count: 40 },
        { id: "s2", name: "SMA N 2 Jakarta", code: "SMA02", student_count: 25 },
      ],
      isLoading: false,
    };
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

  it("renders revenue formatted as Rupiah", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Rp15.000.000")).toBeInTheDocument();
  });

  it("renders school count and a genuine sum of per-school student counts", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    // 2 schools, but 65 students (40 + 25) — the two stats must NOT collapse to the same number.
    expect(getByText("2")).toBeInTheDocument();
    expect(getByText("65")).toBeInTheDocument();
  });

  it("renders 'Belum tersedia' on Sesi Ujian Aktif", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Belum tersedia")).toBeInTheDocument();
  });

  it("shows skeleton when revenue is loading", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    revenueState = { data: null, isLoading: true };
    schoolsState = { data: null, isLoading: true };
    const { container } = render(<AdminIndexPage />);
    // Skeleton renders as divs with animate-pulse
    const skeletons = container.querySelectorAll(".animate-pulse");
    // At least the revenue card skeleton should appear
    // (school skeleton also present — that's fine)
    expect(skeletons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders audit log section for super_admin with real data", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Log Aktivitas")).toBeInTheDocument();
    expect(getByText("Lihat Semua")).toBeInTheDocument();
    // Audit entries come from mock hook data
    expect(getByText("Rina Wijayanti")).toBeInTheDocument();
    expect(getByText("Hendra Gunawan")).toBeInTheDocument();
    expect(getByText("Sri Wahyuni")).toBeInTheDocument();
  });

  it("shows loading skeleton while audit log loads", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    auditState = { data: [], isLoading: true, isError: false, refetch: vi.fn() };
    const { container } = render(<AdminIndexPage />);
    const skeletons = container.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThanOrEqual(2); // stat + audit skeletons
  });

  it("shows empty state when audit log is empty", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    auditState = { data: [], isLoading: false, isError: false, refetch: vi.fn() };
    const { getByText } = render(<AdminIndexPage />);
    expect(getByText("Belum ada aktivitas.")).toBeInTheDocument();
  });

  it("shows error state with retry button when audit log fails", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const refetch = vi.fn();
    auditState = { data: [], isLoading: false, isError: true, refetch };
    const { getByText, getByRole } = render(<AdminIndexPage />);
    expect(getByText("Gagal memuat log aktivitas. Coba lagi nanti.")).toBeInTheDocument();
    fireEvent.click(getByRole("button", { name: /muat ulang/i }));
    expect(refetch).toHaveBeenCalled();
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

  it("quick action buttons navigate to correct routes", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    render(<AdminIndexPage />);
    fireEvent.click(screen.getByText("Buat Soal Baru"));
    expect(push).toHaveBeenCalledWith("/admin/exam/tests");
    fireEvent.click(screen.getByText("Tambah Produk"));
    expect(push).toHaveBeenCalledWith("/admin/products");
    fireEvent.click(screen.getByText("Daftarkan Siswa"));
    expect(push).toHaveBeenCalledWith("/admin/school/students");
    fireEvent.click(screen.getByText("Laporan Penjualan"));
    expect(push).toHaveBeenCalledWith("/admin/revenue");
  });

  it("renders the Shield icon (hero band) for super_admin", () => {
    authStore = { token: "t", user: { role: "super_admin" } };
    const { container } = render(<AdminIndexPage />);
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
