"use client";

import { Calendar } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function ExamPackagesPage() {
  const { t } = useTranslation();
  return (
    <UnderMaintenance
      icon={Calendar}
      title={t("exam_packages_title")}
      estimatedTimeline={t("maint_eta_q4_2026")}
    />
  );
}