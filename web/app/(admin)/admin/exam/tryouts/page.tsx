"use client";

import { ClipboardList } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";

export default function ExamTryoutsPage() {
  return (
    <UnderMaintenance
      icon={ClipboardList}
      title="Daftar Ujian"
      estimatedTimeline="Estimasi rilis: Q4 2026"
    />
  );
}
