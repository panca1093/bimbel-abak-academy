"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Layers } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
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
      <AdminPageHeader
        icon={Layers}
        title={t("exam_packages_page_title")}
        description={t("exam_packages_page_description")}
        actions={
          role === "admin_school" ? undefined : (
            <Button onClick={() => setShowCreate(true)}>
              {t("exam_packages_create")}
            </Button>
          )
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
        items.length === 0 ? (
          <div className="md-card-outlined px-4 py-8 text-center text-muted-foreground">
            {t("exam_packages_empty")}
          </div>
        ) : (
          <div className="md-card-outlined divide-y">
            {items.map((exam) => (
              <div
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
                className="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/40"
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate font-semibold text-ink-900">{exam.title}</div>
                  <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <span>{formatScheduled(exam.scheduled_at)}</span>
                    {exam.duration_minutes ? (
                      <>
                        <span aria-hidden="true">·</span>
                        <span>{exam.duration_minutes} min</span>
                      </>
                    ) : null}
                    {exam.is_free ? (
                      <Badge variant="secondary" className="ml-1">
                        {t("exam_packages_modal_is_free")}
                      </Badge>
                    ) : null}
                  </div>
                </div>
              </div>
            ))}
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