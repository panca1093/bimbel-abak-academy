"use client";

import { School } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function SchoolClassesPage() {
  const { t } = useTranslation();
  return <UnderMaintenance icon={School} title={t("school_classes_title")} />;
}
