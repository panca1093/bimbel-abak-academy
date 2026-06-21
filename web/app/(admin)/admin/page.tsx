"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Shield, DollarSign, Users, Eye, Store, Clipboard, BarChart3, UserPlus } from "lucide-react";
import { useAuthStore } from "@/stores/auth";
import { useMe } from "@/lib/hooks/auth";
import { adminHomeForRole } from "@/lib/auth-redirect";
import { StatCard } from "@/components/admin/StatCard";
import type { UserRole } from "@/lib/nav-config";

const AUDIT = [
  { initial: "R", user: "Rina Wijayanti", action: "Mempublikasikan paket ujian", target: "UTBK 2024 Tryout 1", time: "10m lalu" },
  { initial: "H", user: "Hendra Gunawan", action: "Mengubah stok produk", target: "Modul Fisika Kelas 12", time: "1j lalu" },
  { initial: "S", user: "Sri Wahyuni", action: "Mendaftarkan siswa", target: "Andi P. (SMAN 1)", time: "2j lalu" },
];

const QUICK_ACTIONS = [
  { icon: Clipboard, label: "Buat Soal Baru" },
  { icon: Store, label: "Tambah Produk" },
  { icon: UserPlus, label: "Daftarkan Siswa" },
  { icon: BarChart3, label: "Laporan Penjualan" },
];

export default function AdminIndexPage() {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const storeRole = user?.role as UserRole | undefined;
  const me = useMe({ enabled: !storeRole });
  const role = storeRole ?? (me.data?.role as UserRole | undefined);
  const name = user?.name ?? me.data?.name ?? "Super Admin";

  useEffect(() => {
    if (!role) return;
    if (role !== "super_admin") router.replace(adminHomeForRole(role));
  }, [role, router]);

  if (role !== "super_admin") return null;

  return (
    <div className="fade-in">
      {/* Hero band */}
      <div
        className="mb-8 rounded-[20px] px-8 py-7"
        style={{
          background: "linear-gradient(135deg, #1A5CFF 0%, #0A3DBF 55%, #005B8E 100%)",
          color: "#FFFFFF",
          boxShadow: "0 4px 24px rgba(26,92,255,0.28)",
        }}
      >
        <div className="flex items-center gap-6">
          <div
            className="flex size-[72px] shrink-0 items-center justify-center rounded-[24px]"
            style={{
              backgroundColor: "rgba(255,255,255,0.18)",
              backdropFilter: "blur(8px)",
            }}
          >
            <Shield size={36} color="#FFFFFF" />
          </div>
          <div>
            <div
              className="text-label"
              style={{ letterSpacing: "0.08em", textTransform: "uppercase", opacity: 0.75 }}
            >
              Super Admin · Abak Academy
            </div>
            <h1 className="text-headline" style={{ color: "#FFFFFF" }}>{name}</h1>
            <p className="text-body" style={{ marginTop: "4px", opacity: 0.85 }}>
              Akses penuh ke semua domain. Pantau seluruh platform dari satu tempat.
            </p>
          </div>
        </div>
      </div>

      {/* Stat grid */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-4 mb-8">
        <StatCard
          label="Pendapatan Bulan Ini"
          value="Rp 48,75 Jt"
          trend="+18% vs bulan lalu"
          accent="primary"
          icon={DollarSign}
        />
        <StatCard
          label="Total Siswa"
          value="1.303"
          trend="+42 minggu ini"
          accent="secondary"
          icon={Users}
        />
        <StatCard
          label="Sesi Ujian Aktif"
          value="4"
          trend="Sedang berlangsung"
          accent="error"
          icon={Eye}
        />
        <StatCard
          label="Jumlah Sekolah"
          value="4"
          trend="Bimbel mitra aktif"
          accent="tertiary"
          icon={Store}
        />
      </div>

      {/* Content grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Log Aktivitas */}
        <div className="lg:col-span-2 md-card-outlined">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-title-large">Log Aktivitas</h3>
            <button className="md-btn-tonal" type="button">
              Lihat Semua
            </button>
          </div>
          <div className="space-y-4">
            {AUDIT.map((item, i) => (
              <div key={i} className="flex items-center gap-3">
                <div
                  className="flex size-8 items-center justify-center rounded-full"
                  style={{
                    backgroundColor: "var(--md-sys-color-primary-container)",
                    color: "var(--md-sys-color-primary)",
                  }}
                >
                  <span className="text-label">{item.initial}</span>
                </div>
                <div>
                  <div className="text-body" style={{ fontWeight: 500 }}>
                    {item.user}
                  </div>
                  <div className="text-label color-on-surface-variant">
                    {item.action} · {item.target} · {item.time}
                  </div>
                </div>
              </div>
            ))}
            <div className="flex items-center gap-3">
              <div
                className="flex size-8 items-center justify-center rounded-full"
                style={{
                  backgroundColor: "var(--md-sys-color-primary-container)",
                  color: "var(--md-sys-color-primary)",
                }}
              >
                <span className="text-label">{name.charAt(0).toUpperCase()}</span>
              </div>
              <div>
                <div className="text-body" style={{ fontWeight: 500 }}>
                  {name}
                </div>
                <div className="text-label color-on-surface-variant">
                  Login ke sistem · Portal Super Admin · 3j lalu
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Akses Cepat */}
        <div className="md-card-outlined">
          <h3 className="text-title-large mb-6">Akses Cepat</h3>
          <div className="grid gap-3">
            {QUICK_ACTIONS.map((action, i) => (
              <button
                key={i}
                type="button"
                className="flex items-center gap-3 p-3 rounded-[12px] border-none w-full text-left"
                style={{
                  backgroundColor: "var(--md-sys-color-surface-container-high)",
                  cursor: "pointer",
                }}
              >
                <div
                  className="flex size-10 items-center justify-center rounded-[10px]"
                  style={{
                    backgroundColor: "var(--md-sys-color-primary-container)",
                    color: "var(--md-sys-color-primary)",
                  }}
                >
                  <action.icon size={20} />
                </div>
                <span className="text-body" style={{ fontWeight: 500 }}>
                  {action.label}
                </span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
