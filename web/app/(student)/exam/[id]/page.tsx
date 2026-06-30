"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { AlertCircle, ArrowLeft, Download, Eye, EyeOff } from "lucide-react";
import { toast } from "sonner";

import { downloadCard, useRegistration } from "@/lib/hooks/exam";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

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

export default function ExamDetailPage() {
  const { t, lang } = useTranslation();
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const id = params?.id ?? "";

  const { data: reg, isLoading, isError, error, refetch } = useRegistration(id);

  const [showToken, setShowToken] = useState(false);
  const [downloading, setDownloading] = useState(false);

  const handleDownload = async () => {
    if (!id) return;
    setDownloading(true);
    try {
      await downloadCard(id);
    } catch (err) {
      const message = err instanceof Error ? err.message : t("competition_error");
      toast.error(message);
    } finally {
      setDownloading(false);
    }
  };

  if (isLoading) {
    return <DetailSkeleton />;
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8 md:px-6">
        <Card className="border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              {t("competition_error")}
              {error instanceof Error && error.message
                ? ` ${error.message}`
                : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              {t("retry")}
            </Button>
          </div>
        </Card>
        <Button
          variant="ghost"
          size="sm"
          className="mt-4"
          onClick={() => router.push("/exam")}
        >
          {t("competition_detail_back")}
        </Button>
      </div>
    );
  }

  if (!reg) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8 md:px-6">
        <p className="text-sm text-ink-500">
          {t("competition_detail_not_found")}
        </p>
        <Button
          variant="ghost"
          size="sm"
          className="mt-3"
          onClick={() => router.push("/exam")}
        >
          {t("competition_detail_back")}
        </Button>
      </div>
    );
  }

  const exam = reg.exam;
  const checkInWindow = exam.check_in_window_minutes;

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 md:px-6 md:py-8">
      <Link
        href="/exam"
        className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-ink-600 hover:text-ink-900"
      >
        <ArrowLeft className="size-4" />
        {t("competition_detail_back")}
      </Link>

      <header className="mb-6">
        <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">
          {exam.title}
        </h1>
      </header>

      <Card className="space-y-4 p-5">
        <h2 className="font-serif text-lg font-semibold text-ink-900">
          {t("competition_detail_exam_info")}
        </h2>
        <dl className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-500">
              {t("competition_detail_scheduled_at")}
            </dt>
            <dd className="mt-1 font-medium text-ink-900">
              {formatScheduled(exam.scheduled_at)}
              {lang === "id" ? " WIB" : ""}
            </dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-500">
              {t("competition_detail_token_label")}
            </dt>
            <dd className="mt-1 flex items-center gap-2 font-mono text-base font-semibold text-ink-900">
              <span className="select-all">
                {showToken ? reg.token : "••••••••"}
              </span>
              <Button
                type="button"
                size="xs"
                variant="ghost"
                onClick={() => setShowToken((v) => !v)}
                aria-label={
                  showToken
                    ? t("competition_detail_hide_token")
                    : t("competition_detail_show_token")
                }
              >
                {showToken ? (
                  <EyeOff className="size-3.5" />
                ) : (
                  <Eye className="size-3.5" />
                )}
                <span className="ml-1 text-xs">
                  {showToken
                    ? t("competition_detail_hide_token")
                    : t("competition_detail_show_token")}
                </span>
              </Button>
            </dd>
          </div>
        </dl>
        <div>
          <Button
            type="button"
            size="sm"
            onClick={handleDownload}
            disabled={downloading}
          >
            <Download className="size-4" />
            {t("competition_detail_download_card")}
          </Button>
        </div>
      </Card>

      {exam.requires_checkin && (
        <Card className="mt-4 border-warning/30 bg-warning-bg p-5">
          <h3 className="font-serif text-base font-semibold text-ink-900">
            {t("competition_detail_checkin_heading")}
          </h3>
          <p className="mt-2 text-sm text-ink-700">
            {t("competition_detail_checkin_body").replace(
              "{N}",
              String(checkInWindow ?? 15)
            )}
          </p>
        </Card>
      )}
    </div>
  );
}

function DetailSkeleton() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-6 md:px-6 md:py-8">
      <Skeleton className="mb-4 h-4 w-32" />
      <Skeleton className="mb-6 h-8 w-2/3" />
      <Skeleton className="h-64 w-full rounded-lg" />
    </div>
  );
}
