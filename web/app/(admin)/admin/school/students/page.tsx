"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Search,
  Users,
  GraduationCap,
  MoreHorizontal,
  Mail,
  School,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";

type StudentStatus = "active" | "inactive" | "pending";

interface Student {
  id: string;
  name: string;
  email: string;
  school: string;
  grade: string;
  status: StudentStatus;
  avgScore: number;
  exams: number;
}

const INITIAL_STUDENTS: Student[] = [
  {
    id: "STU-1001",
    name: "Budi Santoso",
    email: "budi.santoso@example.com",
    school: "SMAN 1 Jakarta",
    grade: "Kelas 12",
    status: "active",
    avgScore: 78.5,
    exams: 14,
  },
  {
    id: "STU-1002",
    name: "Citra Wulandari",
    email: "citra.w@example.com",
    school: "SMAN 3 Bandung",
    grade: "Kelas 11",
    status: "active",
    avgScore: 84.2,
    exams: 18,
  },
  {
    id: "STU-1003",
    name: "Andi Pratama",
    email: "andi.p@example.com",
    school: "SMAN 2 Surabaya",
    grade: "Kelas 12",
    status: "pending",
    avgScore: 0,
    exams: 0,
  },
  {
    id: "STU-1004",
    name: "Dewi Lestari",
    email: "dewi.l@example.com",
    school: "SMAN 5 Yogyakarta",
    grade: "Kelas 10",
    status: "inactive",
    avgScore: 62.0,
    exams: 5,
  },
  {
    id: "STU-1005",
    name: "Eko Saputra",
    email: "eko.s@example.com",
    school: "SMAN 1 Jakarta",
    grade: "Kelas 11",
    status: "active",
    avgScore: 71.3,
    exams: 9,
  },
];

const STATUS_TONE: Record<StudentStatus, string> = {
  active: "bg-success-bg text-success border-success",
  inactive: "bg-ink-100 text-ink-600 border-line",
  pending: "bg-warn-bg text-warn border-warn",
};

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

export default function SchoolStudentsPage() {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StudentStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);

  const rows = useMemo(() => {
    return INITIAL_STUDENTS.filter((s) => {
      const matchesStatus = statusFilter === "all" || s.status === statusFilter;
      const q = search.toLowerCase();
      const matchesSearch =
        search.trim() === "" ||
        s.name.toLowerCase().includes(q) ||
        s.email.toLowerCase().includes(q) ||
        s.school.toLowerCase().includes(q);
      return matchesStatus && matchesSearch;
    });
  }, [search, statusFilter]);

  const stats = useMemo(() => {
    const active = INITIAL_STUDENTS.filter((s) => s.status === "active").length;
    const pending = INITIAL_STUDENTS.filter((s) => s.status === "pending").length;
    const schools = new Set(INITIAL_STUDENTS.map((s) => s.school)).size;
    return { total: INITIAL_STUDENTS.length, active, pending, schools };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={Users}
        title="Manajemen Siswa"
        description="Daftar dan kelola data siswa."
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total siswa" value={String(stats.total)} />
        <StatCard label="Aktif" value={String(stats.active)} />
        <StatCard label="Menunggu aktivasi" value={String(stats.pending)} />
        <StatCard label="Mitra sekolah" value={String(stats.schools)} />
      </div>

      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip active={statusFilter === "all"} onClick={() => setStatusFilter("all")}>
            {t("tab_all")}
          </FilterChip>
          <FilterChip active={statusFilter === "active"} onClick={() => setStatusFilter("active")}>
            Aktif
          </FilterChip>
          <FilterChip active={statusFilter === "pending"} onClick={() => setStatusFilter("pending")}>
            Pending
          </FilterChip>
          <FilterChip active={statusFilter === "inactive"} onClick={() => setStatusFilter("inactive")}>
            Nonaktif
          </FilterChip>
        </div>
        <div className="ml-auto flex items-center gap-2">
          <Search className="size-4 text-ink-400" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari siswa atau sekolah…"
            className="h-9 w-[220px] text-xs"
          />
        </div>
      </div>

      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">Siswa</th>
                <th className="px-4 py-3">Sekolah / Kelas</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Skor rata-rata</th>
                <th className="px-4 py-3">Ujian</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((s) => (
                <tr key={s.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Avatar size="sm">
                        <AvatarFallback className="bg-brand-50 text-brand-700 text-xs">
                          {initials(s.name)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-ink-900">{s.name}</div>
                        <div className="flex items-center gap-1 text-[11px] text-ink-500">
                          <Mail className="size-3" />
                          {s.email}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 text-xs text-ink-700">
                      <School className="size-3.5 text-ink-400" />
                      {s.school}
                    </div>
                    <div className="flex items-center gap-1 text-[11px] text-ink-500">
                      <GraduationCap className="size-3" />
                      {s.grade}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn("text-[11px] font-semibold capitalize", STATUS_TONE[s.status])}
                    >
                      {s.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    {s.avgScore > 0 ? (
                      <span
                        className={cn(
                          "rounded-full px-2 py-0.5 text-xs font-semibold",
                          s.avgScore >= 75
                            ? "bg-success-bg text-success"
                            : s.avgScore >= 60
                            ? "bg-warn-bg text-warn"
                            : "bg-danger-bg text-danger"
                        )}
                      >
                        {s.avgScore}%
                      </span>
                    ) : (
                      <span className="text-xs text-ink-400">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">{s.exams}</td>
                  <td className="px-4 py-3 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon-xs">
                          <MoreHorizontal className="size-4 text-ink-500" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => setCreateOpen(true)}>
                          Edit profil
                        </DropdownMenuItem>
                        <DropdownMenuItem>Lihat laporan</DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">{t("create")} siswa</DialogTitle>
            <DialogDescription>
              Form pendaftaran siswa akan tersedia di iterasi berikutnya.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Nama lengkap</Label>
              <Input placeholder="mis. Budi Santoso" />
            </div>
            <div>
              <Label>Email</Label>
              <Input type="email" placeholder="budi@example.com" />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                {t("cancel")}
              </Button>
              <Button onClick={() => setCreateOpen(false)}>{t("create")}</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-lg border px-3 py-[7px] text-xs font-semibold transition-colors",
        active
          ? "border-brand-600 bg-brand-600 text-white"
          : "border-line bg-surface text-ink-600 hover:text-ink-900"
      )}
    >
      {children}
    </button>
  );
}
