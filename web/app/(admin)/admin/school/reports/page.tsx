"use client";

import { useMemo, useState } from "react";
import {
  FileText,
  Download,
  TrendingUp,
  School,
  Users,
  Award,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";

interface SchoolReport {
  id: string;
  school: string;
  students: number;
  exams: number;
  avgScore: number;
  topScore: number;
  growth: number;
}

const INITIAL_REPORTS: SchoolReport[] = [
  {
    id: "RPT-701",
    school: "SMAN 1 Jakarta",
    students: 142,
    exams: 24,
    avgScore: 76.5,
    topScore: 94.2,
    growth: 8.4,
  },
  {
    id: "RPT-702",
    school: "SMAN 3 Bandung",
    students: 98,
    exams: 18,
    avgScore: 72.1,
    topScore: 91.0,
    growth: 3.2,
  },
  {
    id: "RPT-703",
    school: "SMAN 2 Surabaya",
    students: 76,
    exams: 15,
    avgScore: 69.8,
    topScore: 88.5,
    growth: -1.5,
  },
  {
    id: "RPT-704",
    school: "SMAN 5 Yogyakarta",
    students: 54,
    exams: 12,
    avgScore: 74.3,
    topScore: 89.0,
    growth: 5.7,
  },
];

const SUBJECTS = [
  { name: "Matematika", avg: 68 },
  { name: "B. Indonesia", avg: 78 },
  { name: "English", avg: 74 },
  { name: "Penalaran", avg: 66 },
];

export default function SchoolReportsPage() {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<"month" | "quarter" | "year">("month");

  const totals = useMemo(() => {
    const students = INITIAL_REPORTS.reduce((sum, r) => sum + r.students, 0);
    const exams = INITIAL_REPORTS.reduce((sum, r) => sum + r.exams, 0);
    const avgScore =
      INITIAL_REPORTS.reduce((sum, r) => sum + r.avgScore, 0) /
      INITIAL_REPORTS.length;
    return { students, exams, avgScore };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={FileText}
        title="Laporan Sekolah"
        description="Laporan performa per sekolah mitra."
        actions={
          <Tabs value={period} onValueChange={(v) => setPeriod(v as typeof period)}>
            <TabsList className="h-9">
              <TabsTrigger value="month" className="text-xs">
                Bulan
              </TabsTrigger>
              <TabsTrigger value="quarter" className="text-xs">
                Kuartal
              </TabsTrigger>
              <TabsTrigger value="year" className="text-xs">
                Tahun
              </TabsTrigger>
            </TabsList>
          </Tabs>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total siswa" value={totals.students.toLocaleString("id-ID")} />
        <StatCard label="Total ujian" value={String(totals.exams)} />
        <StatCard label="Rata-rata skor" value={`${totals.avgScore.toFixed(1)}%`} />
        <StatCard label="Mitra sekolah" value={String(INITIAL_REPORTS.length)} />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="md-card-outlined lg:col-span-2">
          <div className="flex items-center justify-between px-5 py-4">
            <div className="flex items-center gap-2">
              <School className="size-5 text-brand-600" />
              <h2 className="font-semibold text-ink-900">Performa per sekolah</h2>
            </div>
            <button className="md-btn-outlined">
              <Download className="mr-1 size-4" />
              Export
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
                <tr>
                  <th className="px-4 py-3">Sekolah</th>
                  <th className="px-4 py-3">Siswa</th>
                  <th className="px-4 py-3">Ujian</th>
                  <th className="px-4 py-3">Skor rata-rata</th>
                  <th className="px-4 py-3">Skor tertinggi</th>
                  <th className="px-4 py-3">Pertumbuhan</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {INITIAL_REPORTS.map((r) => (
                  <tr key={r.id} className="hover:bg-surface-2">
                    <td className="px-4 py-3">
                      <div className="font-medium text-ink-900">{r.school}</div>
                      <div className="font-mono text-[11px] text-ink-500">{r.id}</div>
                    </td>
                    <td className="px-4 py-3 text-xs text-ink-600">
                      <span className="inline-flex items-center gap-1">
                        <Users className="size-3" />
                        {r.students}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-ink-600">{r.exams}</td>
                    <td className="px-4 py-3">
                      <ScorePill score={r.avgScore} />
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center gap-1 rounded-full bg-gold-bg px-2 py-0.5 text-xs font-semibold text-gold">
                        <Award className="size-3" />
                        {r.topScore}%
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={cn(
                          "text-xs font-semibold",
                          r.growth >= 0 ? "text-success" : "text-danger"
                        )}
                      >
                        {r.growth >= 0 ? "+" : ""}
                        {r.growth}%
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="md-card-outlined">
          <div className="mb-4 flex items-center gap-2">
            <FileText className="size-5 text-brand-600" />
            <h2 className="font-semibold text-ink-900">Penguasaan materi</h2>
          </div>
          <div className="space-y-4">
            {SUBJECTS.map((s) => (
              <div key={s.name} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-ink-700">{s.name}</span>
                  <span className="text-ink-500">{s.avg}%</span>
                </div>
                <Progress value={s.avg} className="h-2" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function ScorePill({ score }: { score: number }) {
  const tone =
    score >= 75
      ? "bg-success-bg text-success"
      : score >= 60
      ? "bg-warn-bg text-warn"
      : "bg-danger-bg text-danger";
  return (
    <span
      className={cn(
        "rounded-full px-2 py-0.5 text-xs font-semibold",
        tone
      )}
    >
      {score.toFixed(1)}%
    </span>
  );
}
