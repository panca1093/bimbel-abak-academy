"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Users,
  BookOpen,
  MoreHorizontal,
  School,
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";

interface SchoolClass {
  id: string;
  name: string;
  program: string;
  grade: string;
  students: number;
  capacity: number;
  school: string;
  exams: number;
}

const INITIAL_CLASSES: SchoolClass[] = [
  {
    id: "CLS-501",
    name: "UTBK Pagi A",
    program: "SNBT",
    grade: "Kelas 12",
    students: 28,
    capacity: 32,
    school: "SMAN 1 Jakarta",
    exams: 6,
  },
  {
    id: "CLS-502",
    name: "UTBK Pagi B",
    program: "SNBT",
    grade: "Kelas 12",
    students: 30,
    capacity: 30,
    school: "SMAN 1 Jakarta",
    exams: 6,
  },
  {
    id: "CLS-503",
    name: "English Bootcamp",
    program: "IELTS",
    grade: "Kelas 11",
    students: 14,
    capacity: 20,
    school: "SMAN 3 Bandung",
    exams: 3,
  },
  {
    id: "CLS-504",
    name: "Literasi Intensif",
    program: "SNBT",
    grade: "Kelas 10",
    students: 18,
    capacity: 24,
    school: "SMAN 2 Surabaya",
    exams: 4,
  },
];

export default function SchoolClassesPage() {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [program, setProgram] = useState<"all" | string>("all");
  const [createOpen, setCreateOpen] = useState(false);

  const rows = useMemo(() => {
    return INITIAL_CLASSES.filter((c) => {
      const matchesProgram = program === "all" || c.program === program;
      const q = search.toLowerCase();
      const matchesSearch =
        search.trim() === "" ||
        c.name.toLowerCase().includes(q) ||
        c.school.toLowerCase().includes(q);
      return matchesProgram && matchesSearch;
    });
  }, [search, program]);

  const programs = useMemo(
    () => [...new Set(INITIAL_CLASSES.map((c) => c.program))],
    []
  );

  const stats = useMemo(() => {
    const totalStudents = INITIAL_CLASSES.reduce((sum, c) => sum + c.students, 0);
    const totalCapacity = INITIAL_CLASSES.reduce((sum, c) => sum + c.capacity, 0);
    return {
      classes: INITIAL_CLASSES.length,
      students: totalStudents,
      capacity: totalCapacity,
      utilization: totalCapacity > 0 ? Math.round((totalStudents / totalCapacity) * 100) : 0,
    };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={School}
        title="Kelas"
        description="Kelola kelas dan pengelompokan siswa."
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total kelas" value={String(stats.classes)} />
        <StatCard label="Siswa aktif" value={String(stats.students)} />
        <StatCard label="Kapasitas" value={String(stats.capacity)} />
        <StatCard label="Utilisasi" value={`${stats.utilization}%`} />
      </div>

      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip active={program === "all"} onClick={() => setProgram("all")}>
            {t("tab_all")}
          </FilterChip>
          {programs.map((p) => (
            <FilterChip
              key={p}
              active={program === p}
              onClick={() => setProgram(p)}
            >
              {p}
            </FilterChip>
          ))}
        </div>
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Cari kelas atau sekolah…"
          className="h-9 w-[240px] text-xs sm:ml-auto"
        />
      </div>

      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">Kelas</th>
                <th className="px-4 py-3">Program</th>
                <th className="px-4 py-3">Sekolah</th>
                <th className="px-4 py-3">Kapasitas</th>
                <th className="px-4 py-3">Ujian</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((c) => {
                const pct = Math.round((c.students / c.capacity) * 100);
                return (
                  <tr key={c.id} className="group hover:bg-surface-2">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2 font-medium text-ink-900">
                        <BookOpen className="size-4 text-brand-600" />
                        {c.name}
                      </div>
                      <div className="text-[11px] text-ink-500">{c.grade}</div>
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant="outline" className="text-[11px] font-semibold">
                        {c.program}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 text-xs text-ink-700">
                        <School className="size-3.5 text-ink-400" />
                        {c.school}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Progress value={pct} className="w-20" />
                        <span className="text-xs text-ink-600">
                          {c.students}/{c.capacity}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-ink-600">
                      <span className="inline-flex items-center gap-1">
                        <Users className="size-3" />
                        {c.exams} ujian
                      </span>
                    </td>
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
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">{t("create")} kelas</DialogTitle>
            <DialogDescription>Form pembuatan kelas akan tersedia di iterasi berikutnya.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Nama kelas</Label>
              <Input placeholder="mis. UTBK Pagi C" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>Program</Label>
                <Select>
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih program" />
                  </SelectTrigger>
                  <SelectContent>
                    {programs.map((p) => (
                      <SelectItem key={p} value={p}>{p}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Kapasitas</Label>
                <Input type="number" placeholder="32" />
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
