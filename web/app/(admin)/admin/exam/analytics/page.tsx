"use client";

import { BarChart } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";

export default function ExamAnalyticsPage() {
  return (
    <UnderMaintenance
      icon={BarChart}
      title="Analitik Ujian"
      estimatedTimeline="Estimasi rilis: Q4 2026"
    />
  );
}
