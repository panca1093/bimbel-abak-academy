"use client";

import Link from "next/link";
import { AlertCircle, Trophy } from "lucide-react";

import { useRegistrations } from "@/lib/hooks/competition";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { RegistrationListItem } from "@/lib/types";

type DerivedStatus = "upcoming" | "ongoing" | "completed";

const STATUS_BADGE_CLASSES: Record<DerivedStatus, string> = {
  upcoming: "bg-brand-50 text-brand-700 border-brand-100",
  ongoing: "bg-warning-bg text-warning border-warning/30",
  completed: "bg-surface-2 text-ink-500 border-line",
};

function deriveStatus(scheduledAt: string | null): DerivedStatus {
  if (!scheduledAt) return "upcoming";
  const now = Date.now();
  const t = new Date(scheduledAt).getTime();
  if (t > now + 60 * 60 * 1000) return "upcoming";
  if (t < now - 60 * 60 * 1000) return "completed";
  return "ongoing";
}

function formatScheduled(scheduledAt: string | null): string {
  if (!scheduledAt) return "—";
  const d = new Date(scheduledAt);
  if (Number.isNaN(d.getTime())) return "—";
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

export default function CompetitionPage() {
  const { t } = useTranslation();
  const { data, isLoading, isError, error, refetch } = useRegistrations();

  return (
    <div>
      <header className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
          {t("comp_title")}
        </h1>
        <p className="mt-2 text-sm text-ink-500">
          {t("competition_list_subtitle")}
        </p>
      </header>

      {isError && (
        <Card className="mb-8 border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              {t("competition_error")}
              {error instanceof Error && error.message ? ` ${error.message}` : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              {t("retry")}
            </Button>
          </div>
        </Card>
      )}

      {isLoading ? (
        <CompetitionSkeleton />
      ) : data && data.length > 0 ? (
        <div className="space-y-3">
          {data.map((reg) => (
            <RegistrationRow key={reg.id} reg={reg} />
          ))}
        </div>
      ) : (
        <EmptyCompetition />
      )}
    </div>
  );
}

function RegistrationRow({ reg }: { reg: RegistrationListItem }) {
  const { t, lang } = useTranslation();
  const status = deriveStatus(reg.scheduled_at);

  return (
    <Card className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="font-serif text-lg font-semibold text-ink-900">
            {reg.exam_title}
          </h2>
          <span
            className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${STATUS_BADGE_CLASSES[status]}`}
          >
            {t(
              status === "upcoming"
                ? "competition_status_upcoming"
                : status === "ongoing"
                ? "competition_status_ongoing"
                : "competition_status_completed"
            )}
          </span>
        </div>
        <p className="mt-1 text-sm text-ink-500">
          {t("competition_detail_scheduled_at")}: {formatScheduled(reg.scheduled_at)}
          {lang === "id" ? " WIB" : ""}
        </p>
      </div>
      <Button asChild variant="outline" size="sm" className="shrink-0">
        <Link href={`/competition/${reg.id}`}>
          {t("competition_view_detail")}
        </Link>
      </Button>
    </Card>
  );
}

function EmptyCompetition() {
  const { t } = useTranslation();
  return (
    <Card className="flex flex-col items-center justify-center gap-3 border-dashed border-line bg-surface-2 px-6 py-10 text-center">
      <div className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600">
        <Trophy className="size-6" />
      </div>
      <p className="font-semibold text-ink-900">{t("competition_empty")}</p>
    </Card>
  );
}

function CompetitionSkeleton() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-4 w-72" />
      </div>
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-lg" />
        ))}
      </div>
    </div>
  );
}