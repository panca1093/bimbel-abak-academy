"use client";

import { BarChart } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function ExamAnalyticsPage() {
  const { t } = useTranslation();
  return (
    <UnderMaintenance
      icon={BarChart}
      title={t("exam_analytics_title")}
      estimatedTimeline={t("maint_eta_q4_2026")}
    />
  );
}
