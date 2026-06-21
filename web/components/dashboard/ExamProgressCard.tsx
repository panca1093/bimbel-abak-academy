"use client";

import { BarChart3 } from "lucide-react";
import { Card } from "@/components/ui/card";
import { GroupedBars } from "@/components/dashboard/GroupedBars";
import { useTranslation } from "@/lib/i18n";
import type { ExamProgressEntry } from "@/lib/types";

export function ExamProgressCard({
  examProgress,
}: {
  examProgress: ExamProgressEntry[];
}) {
  const { t } = useTranslation();

  return (
    <Card className="flex flex-col border-line px-5 py-5">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          {t("dash_exam_progress")}
        </h3>
        <BarChart3 className="size-4 text-brand-600" />
      </div>

      {examProgress.length === 0 ? (
        <div className="flex flex-1 items-center justify-center py-8 text-center">
          <p className="text-sm text-ink-500">{t("dash_exam_empty")}</p>
        </div>
      ) : (
        <div className="mt-2">
          <GroupedBars
            height={160}
            groups={examProgress.map((e) => [e.completed, e.in_progress])}
            labels={examProgress.map((e) => e.label)}
            series={[
              { label: t("dash_completed_label"), color: "var(--color-brand-600)" },
              { label: t("dash_in_progress_label"), color: "var(--color-brand-200)" },
            ]}
          />
          {/* Legend */}
          <div className="mt-3 flex items-center justify-center gap-4">
            <div className="flex items-center gap-1.5">
              <span
                className="inline-block size-2.5 rounded-sm"
                style={{ background: "var(--color-brand-600)" }}
              />
              <span className="text-[11px] text-ink-500">{t("dash_completed_label")}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span
                className="inline-block size-2.5 rounded-sm"
                style={{ background: "var(--color-brand-200)" }}
              />
              <span className="text-[11px] text-ink-500">{t("dash_in_progress_label")}</span>
            </div>
          </div>
        </div>
      )}
    </Card>
  );
}
