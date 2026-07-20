"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronRight } from "lucide-react";
import { ExamModal } from "@/components/admin/ExamModal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { useExams } from "@/lib/hooks/admin-exams";
import { useTranslation } from "@/lib/i18n";
import { useAuthStore } from "@/stores/auth";
import type { ExamListItem } from "@/lib/types";

// Selling an exam (price/status/publish) is managed on the attached Product(s)
// via /admin/products — mirrors Course, which shows no status/price columns here.
function formatScheduled(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export default function ExamPackagesPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const role = useAuthStore((s) => s.user?.role);
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<ExamListItem | null>(null);

  const { data, isLoading, isError, error } = useExams();
  const items = data?.data ?? [];

  return (
    <div className="space-y-6 fade-in">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="font-serif text-[27px] font-semibold tracking-tight text-ink-900">
            {t("exam_packages_page_title")}
          </h1>
          <p className="mt-1.5 text-sm text-ink-500">{t("exam_packages_page_description")}</p>
        </div>
        {role !== "admin_school" && (
          <Button className="rounded-full" onClick={() => setShowCreate(true)}>
            {t("exam_packages_create")}
          </Button>
        )}
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          {error instanceof Error ? error.message : t("error_generic")}
        </div>
      )}

      {!isLoading && !isError && (
        items.length === 0 ? (
          <div className="md-card-outlined px-4 py-8 text-center text-muted-foreground">
            {t("exam_packages_empty")}
          </div>
        ) : (
          <div className="md-card-outlined overflow-hidden p-0!">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-line text-left text-xs font-semibold uppercase tracking-wide text-ink-500">
                  <th className="px-4 py-3">{t("exam_packages_page_title")}</th>
                  <th className="px-4 py-3">{t("exam_packages_col_scheduled")}</th>
                  <th className="px-4 py-3">Timer</th>
                  <th className="px-4 py-3">{t("th_status")}</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {items.map((exam) => (
                  <tr
                    key={exam.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => router.push(`/admin/exam/packages/${exam.id}`)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        router.push(`/admin/exam/packages/${exam.id}`);
                      }
                    }}
                    className="cursor-pointer transition-colors hover:bg-surface-2"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-ink-900">{exam.title}</span>
                        {exam.is_free && (
                          <Badge variant="secondary">{t("exam_packages_modal_is_free")}</Badge>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs whitespace-nowrap text-ink-600">
                      {formatScheduled(exam.scheduled_at)}
                      {exam.scheduled_end_at && ` – ${formatScheduled(exam.scheduled_end_at)}`}
                    </td>
                    <td className="px-4 py-3 text-xs">
                      {exam.timer_mode && (
                        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-brand-700">
                          {exam.timer_mode}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={exam.status === "published" ? "default" : "secondary"}>
                        {exam.status ?? "draft"}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <ChevronRight className="ml-auto size-4 text-ink-400" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      <ExamModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onSaved={() => setShowCreate(false)}
      />

      <ExamModal
        open={Boolean(editing)}
        exam={editing}
        onClose={() => setEditing(null)}
        onSaved={() => setEditing(null)}
      />
    </div>
  );
}