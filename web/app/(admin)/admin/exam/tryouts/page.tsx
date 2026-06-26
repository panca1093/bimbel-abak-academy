"use client";

import { ClipboardList } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function ExamTryoutsPage() {
  const { t } = useTranslation();
  return (
    <UnderMaintenance
      icon={ClipboardList}
      title={t("exam_tryouts_title")}
      estimatedTimeline={t("maint_eta_q4_2026")}
    />
  );
}
