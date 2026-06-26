"use client";

import { FileText } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function SchoolReportsPage() {
  const { t } = useTranslation();
  return <UnderMaintenance icon={FileText} title={t("school_reports_title")} />;
}
