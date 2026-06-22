"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  ClipboardList,
  Eye,
  MoreHorizontal,
  Users,
  Clock,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";
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
import { cn } from "@/lib/utils";

type TestStatus = "draft" | "published" | "live" | "finished";
type TestType = "tryout" | "quiz" | "competition";

interface ExamTest {
  id: string;
  title: string;
  type: TestType;
  status: TestStatus;
  questions: number;
  durationMin: number;
  participants: number;
  scheduledAt?: string;
}

const INITIAL_TESTS: ExamTest[] = [
  {
    id: "T-1001",
    title: "Tryout SNBT Nasional 2026 #12",
    type: "tryout",
    status: "published",
    questions: 120,
    durationMin: 90,
    participants: 843,
    scheduledAt: "2026-07-12T08:00",
  },
  {
    id: "T-1002",
    title: "Quiz Pemantapan Harian — Matematika",
    type: "quiz",
    status: "live",
    questions: 20,
    durationMin: 15,
    participants: 124,
  },
  {
    id: "T-1003",
    title: "Kompetisi UTBK Antar-Sekolah",
    type: "competition",
    status: "draft",
    questions: 120,
    durationMin: 90,
    participants: 0,
    scheduledAt: "2026-08-05T09:00",
  },
  {
    id: "T-1004",
    title: "Tryout Literasi & Numerasi #11",
    type: "tryout",
    status: "finished",
    questions: 80,
    durationMin: 60,
    participants: 620,
    scheduledAt: "2026-06-10T07:30",
  },
  {
    id: "T-1005",
    title: "Mini Quiz Bahasa Inggris",
    type: "quiz",
    status: "draft",
    questions: 10,
    durationMin: 10,
    participants: 0,
  },
];

const TYPES: TestType[] = ["tryout", "quiz", "competition"];

const STATUS_TONE: Record<TestStatus, string> = {
  draft: "bg-ink-100 text-ink-600 border-line",
  published: "bg-success-bg text-success border-success",
  live: "bg-info-bg text-info border-info",
  finished: "bg-warn-bg text-warn border-warn",
};

const TYPE_LABEL: Record<TestType, string> = {
  tryout: "Tryout",
  quiz: "Quiz",
  competition: "Kompetisi",
};

function fmtDate(iso?: string) {
  if (!iso) return "-";
  const d = new Date(iso);
  return d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function ExamTryoutsPage() {
  const { t } = useTranslation();
  const [type, setType] = useState<TestType | "all">("all");
  const [status, setStatus] = useState<TestStatus | "all">("all");
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);

  const filtered = useMemo(() => {
    return INITIAL_TESTS.filter((test) => {
      const matchesType = type === "all" || test.type === type;
      const matchesStatus = status === "all" || test.status === status;
      const matchesSearch =
        search.trim() === "" ||
        test.title.toLowerCase().includes(search.toLowerCase());
      return matchesType && matchesStatus && matchesSearch;
    });
  }, [type, status, search]);

  const stats = useMemo(() => {
    const live = INITIAL_TESTS.filter((t) => t.status === "live").length;
    const published = INITIAL_TESTS.filter((t) => t.status === "published").length;
    const totalParticipants = INITIAL_TESTS.reduce(
      (sum, t) => sum + t.participants,
      0
    );
    return { live, published, totalParticipants };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={ClipboardList}
        title="Daftar Ujian"
        description="Susun, publikasikan, dan pantau sesi ujian."
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-3">
        <StatCard label="Sesi aktif" value={stats.live.toString()} icon={Clock} />
        <StatCard label="Terpublikasi" value={stats.published.toString()} icon={ClipboardList} />
        <StatCard label="Total peserta" value={stats.totalParticipants.toLocaleString("id-ID")} icon={Users} />
      </div>

      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip active={type === "all"} onClick={() => setType("all")}>
            {t("tab_all")}
          </FilterChip>
          {TYPES.map((tType) => (
            <FilterChip
              key={tType}
              active={type === tType}
              onClick={() => setType(tType)}
            >
              {TYPE_LABEL[tType]}
            </FilterChip>
          ))}
        </div>
        <div className="ml-auto flex items-center gap-2">
          <Select value={status} onValueChange={(v) => setStatus(v as TestStatus | "all")}>
            <SelectTrigger className="h-9 w-[150px] text-xs">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua status</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="published">Terpublikasi</SelectItem>
              <SelectItem value="live">Sedang berjalan</SelectItem>
              <SelectItem value="finished">Selesai</SelectItem>
            </SelectContent>
          </Select>
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari ujian…"
            className="h-9 w-[200px] text-xs"
          />
        </div>
      </div>

      <div className="md-card-outlined overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">ID</th>
                <th className="px-4 py-3">Nama ujian</th>
                <th className="px-4 py-3">Tipe</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Soal</th>
                <th className="px-4 py-3">Durasi</th>
                <th className="px-4 py-3">Peserta</th>
                <th className="px-4 py-3">Jadwal</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {filtered.map((test) => (
                <tr key={test.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3 font-mono text-xs text-ink-500">
                    {test.id}
                  </td>
                  <td className="px-4 py-3">
                    <div className="max-w-[260px] truncate font-medium text-ink-900">
                      {test.title}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center rounded-md bg-surface-2 px-2 py-1 text-xs font-medium text-ink-600">
                      {TYPE_LABEL[test.type]}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold capitalize",
                        STATUS_TONE[test.status]
                      )}
                    >
                      {test.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs">{test.questions}</td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {test.durationMin} mnt
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Users className="size-3" />
                      {test.participants.toLocaleString("id-ID")}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {fmtDate(test.scheduledAt)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon-xs">
                          <MoreHorizontal className="size-4 text-ink-500" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem>
                          <Eye className="mr-2 size-4" />
                          Lihat detail
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setCreateOpen(true)}>
                          <ClipboardList className="mr-2 size-4" />
                          Edit
                        </DropdownMenuItem>
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
            <DialogTitle className="font-serif">{t("create")} ujian</DialogTitle>
            <DialogDescription>
              Builder ujian akan tersedia di iterasi berikutnya.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Judul ujian</Label>
              <Input placeholder="mis. Tryout SNBT #13" />
            </div>
            <div>
              <Label>Tipe</Label>
              <Select>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tipe" />
                </SelectTrigger>
                <SelectContent>
                  {TYPES.map((tt) => (
                    <SelectItem key={tt} value={tt}>
                      {TYPE_LABEL[tt]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
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
