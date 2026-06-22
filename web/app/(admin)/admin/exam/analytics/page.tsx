"use client";

import { useMemo, useState } from "react";
import {
  BarChart,
  BarChart3,
  TrendingUp,
  Users,
  Calendar,
  Clock,
  Target,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

type Period = "7d" | "30d" | "90d";

interface TestMetric {
  id: string;
  title: string;
  participants: number;
  avgScore: number;
  completion: number;
  topScore: number;
}

const MOCK_METRICS: Record<Period, TestMetric[]> = {
  "7d": [
    { id: "T-1001", title: "Tryout SNBT #12", participants: 843, avgScore: 72.4, completion: 94, topScore: 96.5 },
    { id: "T-1002", title: "Quiz Pemantapan Harian", participants: 124, avgScore: 68.1, completion: 88, topScore: 92.0 },
  ],
  "30d": [
    { id: "T-1001", title: "Tryout SNBT #12", participants: 2104, avgScore: 71.2, completion: 91, topScore: 97.0 },
    { id: "T-1002", title: "Quiz Pemantapan Harian", participants: 892, avgScore: 69.4, completion: 86, topScore: 93.5 },
    { id: "T-1004", title: "Tryout Literasi #11", participants: 620, avgScore: 75.6, completion: 89, topScore: 95.0 },
  ],
  "90d": [
    { id: "T-1001", title: "Tryout SNBT #12", participants: 5410, avgScore: 70.8, completion: 90, topScore: 98.0 },
    { id: "T-1002", title: "Quiz Pemantapan Harian", participants: 2410, avgScore: 68.9, completion: 85, topScore: 94.0 },
    { id: "T-1004", title: "Tryout Literasi #11", participants: 1830, avgScore: 74.2, completion: 88, topScore: 96.0 },
    { id: "T-1003", title: "Kompetisi UTBK Antar-Sekolah", participants: 0, avgScore: 0, completion: 0, topScore: 0 },
  ],
};

const OVERVIEW: Record<Period, { exams: number; participants: number; avgScore: number; completion: number }> = {
  "7d": { exams: 2, participants: 967, avgScore: 71.3, completion: 92 },
  "30d": { exams: 3, participants: 3616, avgScore: 71.2, completion: 89 },
  "90d": { exams: 4, participants: 9650, avgScore: 70.5, completion: 88 },
};

const SUBJECT_DISTRIBUTION = [
  { subject: "Matematika", correct: 68 },
  { subject: "B. Indonesia", correct: 76 },
  { subject: "English", correct: 72 },
  { subject: "Penalaran", correct: 65 },
  { subject: "Literasi", correct: 74 },
];

export default function ExamAnalyticsPage() {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<Period>("30d");

  const metrics = useMemo(() => MOCK_METRICS[period], [period]);
  const overview = OVERVIEW[period];

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={BarChart}
        title="Analitik Ujian"
        description="Statistik skor dan performa siswa."
        actions={
          <Tabs value={period} onValueChange={(v) => setPeriod(v as Period)}>
            <TabsList className="h-9">
              <TabsTrigger value="7d" className="text-xs">7 hari</TabsTrigger>
              <TabsTrigger value="30d" className="text-xs">30 hari</TabsTrigger>
              <TabsTrigger value="90d" className="text-xs">90 hari</TabsTrigger>
            </TabsList>
          </Tabs>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Ujian" value={overview.exams.toString()} icon={Calendar} />
        <StatCard label="Peserta" value={overview.participants.toLocaleString("id-ID")} icon={Users} />
        <StatCard label="Rata-rata skor" value={`${overview.avgScore}%`} icon={Target} />
        <StatCard label="Penyelesaian" value={`${overview.completion}%`} icon={TrendingUp} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="md-card-outlined p-5">
          <div className="mb-4 flex items-center gap-2">
            <BarChart3 className="size-5 text-brand-600" />
            <h2 className="font-semibold text-ink-900">Performa per ujian</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
                <tr>
                  <th className="px-3 py-2">Ujian</th>
                  <th className="px-3 py-2">Peserta</th>
                  <th className="px-3 py-2">Skor rata-rata</th>
                  <th className="px-3 py-2">Selesai</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {metrics.map((m) => (
                  <tr key={m.id} className="hover:bg-surface-2">
                    <td className="px-3 py-2">
                      <div className="font-medium text-ink-900">{m.title}</div>
                      <div className="font-mono text-[11px] text-ink-500">{m.id}</div>
                    </td>
                    <td className="px-3 py-2">{m.participants.toLocaleString("id-ID")}</td>
                    <td className="px-3 py-2">
                      <ScorePill score={m.avgScore} />
                    </td>
                    <td className="px-3 py-2">{m.completion}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="md-card-outlined p-5">
          <div className="mb-4 flex items-center gap-2">
            <Clock className="size-5 text-brand-600" />
            <h2 className="font-semibold text-ink-900">Distribusi materi</h2>
          </div>
          <div className="space-y-4">
            {SUBJECT_DISTRIBUTION.map((s) => (
              <div key={s.subject} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-ink-700">{s.subject}</span>
                  <span className="text-ink-500">{s.correct}% benar</span>
                </div>
                <div className="h-2.5 w-full overflow-hidden rounded-full bg-line">
                  <div
                    className="h-full rounded-full bg-brand-600"
                    style={{ width: `${s.correct}%` }}
                  />
                </div>
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
    score >= 75 ? "bg-success-bg text-success" : score >= 60 ? "bg-warn-bg text-warn" : "bg-danger-bg text-danger";
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-semibold", tone)}>
      {score}%
    </span>
  );
}
