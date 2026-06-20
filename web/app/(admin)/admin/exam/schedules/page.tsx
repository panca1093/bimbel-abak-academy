"use client";

import { useMemo, useState } from "react";
import { Plus, Calendar, Clock, Users, MoreHorizontal, Edit, Trash2 } from "lucide-react";
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

type ScheduleStatus = "scheduled" | "open" | "ongoing" | "closed";

interface ExamSchedule {
  id: string;
  testId: string;
  title: string;
  startAt: string;
  endAt: string;
  timezone: string;
  participants: number;
  status: ScheduleStatus;
  token?: string;
}

const INITIAL_SCHEDULES: ExamSchedule[] = [
  {
    id: "SCH-201",
    testId: "T-1001",
    title: "Tryout SNBT Nasional 2026 #12 — Sesi Pagi",
    startAt: "2026-07-12T08:00",
    endAt: "2026-07-12T10:30",
    timezone: "Asia/Jakarta",
    participants: 843,
    status: "scheduled",
    token: "SNBT12-PAGI",
  },
  {
    id: "SCH-202",
    testId: "T-1002",
    title: "Quiz Pemantapan Harian — Matematika",
    startAt: "2026-06-20T19:00",
    endAt: "2026-06-20T19:15",
    timezone: "Asia/Jakarta",
    participants: 124,
    status: "ongoing",
  },
  {
    id: "SCH-203",
    testId: "T-1004",
    title: "Tryout Literasi & Numerasi #11",
    startAt: "2026-06-10T07:30",
    endAt: "2026-06-10T09:30",
    timezone: "Asia/Jakarta",
    participants: 620,
    status: "closed",
  },
  {
    id: "SCH-204",
    testId: "T-1003",
    title: "Kompetisi UTBK Antar-Sekolah 2026",
    startAt: "2026-08-05T09:00",
    endAt: "2026-08-05T11:00",
    timezone: "Asia/Jakarta",
    participants: 0,
    status: "scheduled",
    token: "KOM-UTBK26",
  },
];

const STATUS_TONE: Record<ScheduleStatus, string> = {
  scheduled: "bg-ink-100 text-ink-600 border-line",
  open: "bg-success-bg text-success border-success",
  ongoing: "bg-info-bg text-info border-info",
  closed: "bg-warn-bg text-warn border-warn",
};

function fmtSchedule(iso: string) {
  const d = new Date(iso);
  return d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function ExamSchedulesPage() {
  const { t } = useTranslation();
  const [createOpen, setCreateOpen] = useState(false);
  const [filter, setFilter] = useState<ScheduleStatus | "all">("all");

  const rows = useMemo(() => {
    if (filter === "all") return INITIAL_SCHEDULES;
    return INITIAL_SCHEDULES.filter((s) => s.status === filter);
  }, [filter]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("schedules")}
          </h1>
          <p className="mt-2 text-sm text-ink-500">
            Atur jadwal sesi ujian, token check-in, dan durasi.
          </p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1 size-4" />
          {t("create")}
        </Button>
      </header>

      <div className="mb-4 flex flex-wrap gap-2">
        <FilterChip active={filter === "all"} onClick={() => setFilter("all")}>
          {t("tab_all")}
        </FilterChip>
        <FilterChip active={filter === "scheduled"} onClick={() => setFilter("scheduled")}>
          Terjadwal
        </FilterChip>
        <FilterChip active={filter === "open"} onClick={() => setFilter("open")}>
          Check-in dibuka
        </FilterChip>
        <FilterChip active={filter === "ongoing"} onClick={() => setFilter("ongoing")}>
          Sedang berjalan
        </FilterChip>
        <FilterChip active={filter === "closed"} onClick={() => setFilter("closed")}>
          Selesai
        </FilterChip>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">ID</th>
                <th className="px-4 py-3">Sesi</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Mulai</th>
                <th className="px-4 py-3">Selesai</th>
                <th className="px-4 py-3">Token</th>
                <th className="px-4 py-3">Peserta</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((sch) => (
                <tr key={sch.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3 font-mono text-xs text-ink-500">{sch.id}</td>
                  <td className="px-4 py-3">
                    <div className="max-w-[260px] truncate font-medium text-ink-900">{sch.title}</div>
                    <div className="text-[11px] text-ink-500">{sch.testId}</div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn("text-[11px] font-semibold capitalize", STATUS_TONE[sch.status])}
                    >
                      {sch.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">{fmtSchedule(sch.startAt)}</td>
                  <td className="px-4 py-3 text-xs text-ink-600">{fmtSchedule(sch.endAt)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-ink-600">
                    {sch.token || "—"}
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Users className="size-3" />
                      {sch.participants.toLocaleString("id-ID")}
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
                          <Edit className="mr-2 size-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-danger">
                          <Trash2 className="mr-2 size-4" />
                          Hapus
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
            <DialogTitle className="font-serif">{t("create")} jadwal</DialogTitle>
            <DialogDescription>Scheduler sesi ujian akan tersedia di iterasi berikutnya.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Pilih ujian</Label>
              <Input placeholder="Cari ujian…" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>Mulai</Label>
                <Input type="datetime-local" />
              </div>
              <div>
                <Label>Selesai</Label>
                <Input type="datetime-local" />
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
