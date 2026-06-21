"use client";

import { Construction } from "lucide-react";
import { useTranslation } from "@/lib/i18n";

export default function CompetitionPage() {
  const { t } = useTranslation();

  return (
    <div className="mx-auto flex min-h-[60vh] max-w-6xl items-center justify-center px-4">
      <div className="flex flex-col items-center gap-4 text-center">
        <div className="flex size-16 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
          <Construction className="size-8" />
        </div>
        <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">
          {t("comp_title")}
        </h1>
        <p className="max-w-sm text-sm text-ink-500">
          {t("comp_maintenance")}
        </p>
      </div>
    </div>
  );
}
