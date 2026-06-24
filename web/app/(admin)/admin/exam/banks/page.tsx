"use client";

import { FileQuestion } from "lucide-react";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";

export default function ExamBanksPage() {
  return (
    <UnderMaintenance
      icon={FileQuestion}
      title="Bank Soal"
      estimatedTimeline="Estimasi rilis: Q4 2026"
    />
  );
}
