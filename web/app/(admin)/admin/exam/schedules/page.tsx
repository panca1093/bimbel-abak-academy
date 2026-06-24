"use client";

import { Calendar } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";

export default function ExamSchedulesPage() {
  return (
    <UnderMaintenance
      icon={Calendar}
      title="Jadwal Ujian"
      estimatedTimeline="Estimasi rilis: Q4 2026"
    />
  );
}
