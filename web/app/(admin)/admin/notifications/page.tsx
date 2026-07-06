"use client";

import { Bell } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { useTranslation } from "@/lib/i18n";

export default function NotificationsPage() {
  const { t } = useTranslation();
  return <UnderMaintenance icon={Bell} title={t("notifications")} />;
}
