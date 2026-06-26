"use client";

import { Calendar } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function ExamSchedulesPage() {
  const { t } = useTranslation();
  return (
    <UnderMaintenance
      icon={Calendar}
      title={t("exam_schedules_title")}
      estimatedTimeline={t("maint_eta_q4_2026")}
    />
  );
}
