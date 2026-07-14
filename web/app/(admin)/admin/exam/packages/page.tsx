"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Package } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { ExamModal } from "@/components/admin/ExamModal";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useExams } from "@/lib/hooks/admin-exams";
import { useTranslation } from "@/lib/i18n";
import { formatRupiah } from "@/lib/format";
import type { ExamListItem } from "@/lib/types";

function statusBadgeClass(status?: string): string {
  switch (status) {
    case "published":
      return "bg-green-100 text-green-800 border-green-200";
    case "draft":
      return "bg-line-2 text-ink-700 border-line";
    case "hidden":
      return "bg-amber-100 text-amber-800 border-amber-200";
    case "archived":
      return "bg-red-100 text-red-800 border-red-200";
    default:
      return "bg-line-2 text-ink-700 border-line";
  }
}

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
                <th className="px-4 py-3 text-left font-medium">
                  {t("exam_packages_col_status")}
                </th>
                <th className="px-4 py-3 text-right font-medium">
                  {t("exam_packages_col_price")}
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
                  <td className="px-4 py-3">
                    <Badge className={statusBadgeClass(exam.product_status)}>
                      {exam.product_status ?? "draft"}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-right">
                    {formatRupiah(exam.product_price)}
                  </td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
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