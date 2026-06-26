"use client";

import { Users } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function SchoolStudentsPage() {
  const { t } = useTranslation();
  return <UnderMaintenance icon={Users} title={t("school_students_title")} />;
}
