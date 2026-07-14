"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Package } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { ExamModal } from "@/components/admin/ExamModal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useExams } from "@/lib/hooks/admin-exams";
import { useTranslation } from "@/lib/i18n";
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
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<ExamListItem | null>(null);

  const { data, isLoading, isError, error } = useExams();
  const items = data?.data ?? [];

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={Package}
        title={t("exam_packages_page_title")}
        description={t("exam_packages_page_description")}
        actions={
          <Button onClick={() => setShowCreate(true)}>
            {t("exam_packages_create")}
          </Button>
        }
      />

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
        <div className="overflow-x-auto md-card-outlined">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">
                  {t("exam_packages_col_title")}
                </th>
                <th className="px-4 py-3 text-left font-medium">
                  {t("exam_packages_col_scheduled")}
                </th>
              </tr>
            </thead>
            <tbody>
              {items.map((exam) => (
                <tr
                  key={exam.id}
                  className="cursor-pointer border-t transition-colors hover:bg-muted/40"
                  onClick={() => router.push(`/admin/exam/packages/${exam.id}`)}
                >
                  <td className="px-4 py-3 font-medium">{exam.title}</td>
                  <td className="px-4 py-3">{formatScheduled(exam.scheduled_at)}</td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr>
                  <td colSpan={2} className="px-4 py-8 text-center text-muted-foreground">
                    {t("exam_packages_empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
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