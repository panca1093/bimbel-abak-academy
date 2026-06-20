"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Building,
  Users,
  MapPin,
  MoreHorizontal,
  Edit,
  Trash2,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

type SchoolStatus = "active" | "inactive" | "pending";

interface SchoolOrg {
  id: string;
  name: string;
  city: string;
  province: string;
  status: SchoolStatus;
  students: number;
  classes: number;
  exams: number;
}

const INITIAL_SCHOOLS: SchoolOrg[] = [
  {
    id: "SCH-501",
    name: "SMAN 1 Jakarta",
    city: "Jakarta",
    province: "DKI Jakarta",
    status: "active",
    students: 142,
    classes: 6,
    exams: 24,
  },
  {
    id: "SCH-502",
    name: "SMAN 3 Bandung",
    city: "Bandung",
    province: "Jawa Barat",
    status: "active",
    students: 98,
    classes: 4,
    exams: 18,
  },
  {
    id: "SCH-503",
    name: "SMAN 2 Surabaya",
    city: "Surabaya",
    province: "Jawa Timur",
    status: "active",
    students: 76,
    classes: 3,
    exams: 15,
  },
  {
    id: "SCH-504",
    name: "SMAN 5 Yogyakarta",
    city: "Yogyakarta",
    province: "DI Yogyakarta",
    status: "pending",
    students: 0,
    classes: 0,
    exams: 0,
  },
];

const STATUS_TONE: Record<SchoolStatus, string> = {
  active: "bg-success-bg text-success border-success",
  inactive: "bg-ink-100 text-ink-600 border-line",
  pending: "bg-warn-bg text-warn border-warn",
};

export default function SystemSchoolsPage() {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<SchoolStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);

  const rows = useMemo(() => {
    return INITIAL_SCHOOLS.filter((s) => {
      const q = search.toLowerCase();
      const matchesSearch =
        search.trim() === "" ||
        s.name.toLowerCase().includes(q) ||
        s.city.toLowerCase().includes(q) ||
        s.province.toLowerCase().includes(q);
      const matchesStatus = statusFilter === "all" || s.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [search, statusFilter]);

  const stats = useMemo(() => {
    const students = INITIAL_SCHOOLS.reduce((sum, s) => sum + s.students, 0);
    const active = INITIAL_SCHOOLS.filter((s) => s.status === "active").length;
    return {
      total: INITIAL_SCHOOLS.length,
      students,
      active,
    };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("schools")}
          </h1>
          <p className="mt-2 text-sm text-ink-500">
            Daftar mitra sekolah, status kerja sama, dan statistik penggunaan.
          </p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1 size-4" />
          {t("create")}
        </Button>
      </header>

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card className="p-5">
          <div className="text-xs text-ink-500">Total sekolah</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.total}</div>
        </Card>
        <Card className="p-5">
          <div className="text-xs text-ink-500">Aktif</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.active}</div>
        </Card>
        <Card className="p-5">
          <div className="text-xs text-ink-500">Total siswa</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">
            {stats.students.toLocaleString("id-ID")}
          </div>
        </Card>
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
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Cari sekolah atau kota…"
          className="h-9 w-[260px] text-xs sm:ml-auto"
        />
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">Sekolah</th>
                <th className="px-4 py-3">Lokasi</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Siswa</th>
                <th className="px-4 py-3">Kelas</th>
                <th className="px-4 py-3">Ujian</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((s) => (
                <tr key={s.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 font-medium text-ink-900">
                      <Building className="size-4 text-brand-600" />
                      {s.name}
                    </div>
                    <div className="font-mono text-[11px] text-ink-500">{s.id}</div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 text-xs text-ink-700">
                      <MapPin className="size-3.5 text-ink-400" />
                      {s.city}, {s.province}
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
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Users className="size-3" />
                      {s.students.toLocaleString("id-ID")}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">{s.classes}</td>
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
                          <Edit className="mr-2 size-4" />Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-danger">
                          <Trash2 className="mr-2 size-4" />Hapus
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">{t("create")} sekolah</DialogTitle>
            <DialogDescription>Form tambah mitra sekolah akan tersedia di iterasi berikutnya.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Nama sekolah</Label>
              <Input placeholder="mis. SMAN 1 Jakarta" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>Kota</Label>
                <Input placeholder="Kota" />
              </div>
              <div>
                <Label>Provinsi</Label>
                <Input placeholder="Provinsi" />
              </div>
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
