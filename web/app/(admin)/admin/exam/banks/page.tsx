"use client";

import { FileQuestion } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function ExamBanksPage() {
  const { t } = useTranslation();
  return (
    <UnderMaintenance
      icon={FileQuestion}
      title={t("exam_banks_title")}
      estimatedTimeline={t("maint_eta_q4_2026")}
    />
  );
}
