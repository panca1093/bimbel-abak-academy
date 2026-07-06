"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { AlertCircle, ArrowLeft, Medal, Trophy } from "lucide-react";

import { useSessionLeaderboard } from "@/lib/hooks/exam";
import type { ExamLeaderboardEntry } from "@/lib/types";
import { ApiError } from "@/lib/api";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default function SessionLeaderboardPage() {
  const { t } = useTranslation();
  const params = useParams<{ id: string }>();
  const sessionId = params?.id ?? "";

  const [entries, setEntries] = useState<ExamLeaderboardEntry[]>([]);
  const [cursor, setCursor] = useState<string | undefined>(undefined);

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
  } = useSessionLeaderboard(sessionId, cursor ? { cursor } : undefined);

  // Reset pagination when navigating to a different session
  useEffect(() => {
    setEntries([]);
    setCursor(undefined);
  }, [sessionId]);

  // Accumulate leaderboard pages
  useEffect(() => {
    if (!data) return;
    if (!cursor) {
      setEntries(data.data);
    } else {
      setEntries((prev) => [...prev, ...data.data]);
    }
  }, [data]);

  // ── 403: leaderboard not available ─────────────────────────────────────────

  if (
    isError &&
    error instanceof ApiError &&
    error.status === 403 &&
    error.code === "leaderboard_not_available"
  ) {
    return (
      <div className="mx-auto flex max-w-md flex-col items-center justify-center px-4 py-24 text-center">
        <Medal className="mb-4 size-12 text-ink-400" />
        <h1 className="text-xl font-bold text-ink-900">
          {t("result_leaderboard_not_available")}
        </h1>
      </div>
    );
  }

  // ── Generic error ──────────────────────────────────────────────────────────

  if (isError) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8">
        <Card className="border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              {t("sys_error_load")}
              {error instanceof Error && error.message
                ? ` ${error.message}`
                : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              {t("retry")}
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  // ── Loading (first page only — keep accumulated rows during load-more) ────

  if (entries.length === 0 && (isLoading || !data)) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8">
        <p className="mb-4 text-sm text-ink-500">{t("sys_loading")}</p>
        <Skeleton className="mb-4 h-12 w-full rounded-lg" />
        <Skeleton className="mb-2 h-10 w-full rounded-lg" />
        <Skeleton className="mb-2 h-10 w-full rounded-lg" />
        <Skeleton className="h-10 w-full rounded-lg" />
      </div>
    );
  }

  // ── Empty state ───────────────────────────────────────────────────────────

  if (entries.length === 0) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8">
        <Link
          href={`/exam/sessions/${sessionId}/result`}
          className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-ink-600 hover:text-ink-900"
        >
          <ArrowLeft className="size-4" />
          {t("competition_detail_back")}
        </Link>
        <Card className="p-5 text-center text-sm text-ink-500">
          {t("admin_exam_leaderboard_empty")}
        </Card>
      </div>
    );
  }

  // ── Leaderboard ──────────────────────────────────────────────────────────

  return (
    <div className="mx-auto max-w-3xl px-4 py-8">
      <Link
        href={`/exam/sessions/${sessionId}/result`}
        className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-ink-600 hover:text-ink-900"
      >
        <ArrowLeft className="size-4" />
        {t("competition_detail_back")}
      </Link>

      {/* Header row */}
      <div className="mb-1 grid grid-cols-[3rem_1fr_5rem] gap-2 px-4 text-xs font-semibold uppercase tracking-wide text-ink-500">
        <span>{t("admin_exam_leaderboard_col_rank")}</span>
        <span>{t("admin_exam_leaderboard_col_student")}</span>
        <span className="text-right">{t("admin_exam_leaderboard_col_score")}</span>
      </div>

      <div className="space-y-1">
        {entries.map((entry) => {
          const isTop3 = entry.rank <= 3;
          return (
            <Card
              key={entry.session_id}
              className={`flex items-center gap-2 px-4 py-3 ${
                isTop3 ? "border-brand-200 bg-brand-50/50" : ""
              }`}
            >
              <span
                className={`flex size-8 shrink-0 items-center justify-center rounded-full text-sm font-bold ${
                  entry.rank === 1
                    ? "bg-yellow-100 text-yellow-700"
                    : entry.rank === 2
                      ? "bg-gray-100 text-gray-600"
                      : entry.rank === 3
                        ? "bg-orange-100 text-orange-700"
                        : "text-ink-600"
                }`}
              >
                {entry.rank === 1 ? (
                  <Trophy className="size-4" />
                ) : (
                  `#${entry.rank}`
                )}
              </span>
              <span className="flex-1 truncate text-sm font-medium text-ink-800">
                {entry.student_name}
              </span>
              <span className="w-12 text-right text-sm font-semibold text-ink-900">
                {entry.score}
              </span>
            </Card>
          );
        })}
      </div>

      {/* Load more */}
      {data?.next_cursor && (
        <div className="mt-4 flex justify-center">
          <Button
            type="button"
            variant="outline"
            onClick={() => setCursor(data.next_cursor)}
            disabled={isFetching}
          >
            {isFetching ? t("sys_loading") : t("sys_load_more")}
          </Button>
        </div>
      )}
    </div>
  );
}
