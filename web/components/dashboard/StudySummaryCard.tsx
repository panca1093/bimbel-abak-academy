"use client";

import { Card } from "@/components/ui/card";
import { Donut } from "@/components/dashboard/Donut";
import { useTranslation } from "@/lib/i18n";
import type { DashboardStudySummary } from "@/lib/types";

export function StudySummaryCard({ study }: { study: DashboardStudySummary }) {
  const { t } = useTranslation();
  const total = study.total_lectures;
  const visited = study.visited_lectures;
  const pct = total > 0 ? visited / total : 0;

  return (
    <Card className="flex flex-col border-line px-5 py-5">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          {t("dash_study_hours")}
        </h3>
      </div>

      {total === 0 ? (
        <div className="flex flex-1 items-center justify-center py-8 text-center">
          <p className="text-sm text-ink-500">{t("dash_study_empty")}</p>
        </div>
      ) : (
        <div className="mt-2 flex items-center gap-4">
          <Donut
            size={120}
            thickness={16}
            value={pct}
            centerLabel={t("dash_donut_lectures")}
            centerSub={t("dash_hours")}
          />
          <div className="flex flex-col gap-4">
            <div className="flex gap-3">
              <span
                className="w-1 shrink-0 rounded"
                style={{ background: "var(--color-brand-300)" }}
              />
              <div>
                <div className="font-serif text-lg font-bold text-ink-900">
                  {visited}
                  <span className="text-sm font-normal text-ink-400">/{total}</span>
                </div>
                <div className="text-xs text-ink-500">{t("dash_visited_lectures")}</div>
              </div>
            </div>
            <div className="flex gap-3">
              <span
                className="w-1 shrink-0 rounded"
                style={{ background: "var(--color-brand-600)" }}
              />
              <div>
                <div className="font-serif text-lg font-bold text-ink-900">
                  {study.completed_courses}
                  <span className="text-sm font-normal text-ink-400">
                    /{study.enrolled_courses_count}
                  </span>
                </div>
                <div className="text-xs text-ink-500">{t("dash_completed_courses_label")}</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </Card>
  );
}
