"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useParams } from "next/navigation";
import {
  AlertCircle,
  CheckCircle2,
  Maximize2,
  ChevronLeft,
  ChevronRight,
  Flag,
  BookOpen,
} from "lucide-react";

import {
  useReconnectSession,
  useSaveAnswers,
  useSubmitSession,
  useLogViolation,
} from "@/lib/hooks/exam";
import { useTranslation, DICT } from "@/lib/i18n";
type I18nKey = keyof typeof DICT.id;
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import type {
  SessionQuestion,
  SubmitResult,
} from "@/lib/types";

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export default function SessionPage() {
  const { t } = useTranslation();
  const params = useParams<{ id: string }>();
  const sessionId = params?.id ?? "";

  const {
    data: session,
    isLoading,
    isError,
    error,
    refetch,
  } = useReconnectSession(sessionId);
  const saveAnswers = useSaveAnswers(sessionId);
  const submitSession = useSubmitSession(sessionId);
  const logViolation = useLogViolation(sessionId);

  const [submitResult, setSubmitResult] = useState<SubmitResult | null>(null);
  const [fullscreenGranted, setFullscreenGranted] = useState(false);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [flagged, setFlagged] = useState<Record<string, boolean>>({});
  const [currentQIndex, setCurrentQIndex] = useState(0);
  const [remaining, setRemaining] = useState<number>(0);
  const [showConfirm, setShowConfirm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const autoSubmittedRef = useRef(false);
  const answersRef = useRef(answers);
  answersRef.current = answers;

  const allQuestions = session
    ? session.tests.flatMap((t) => t.questions)
    : [];

  // Initialize from session data (reconnect)
  useEffect(() => {
    if (!session) return;
    const initAnswers: Record<string, string> = {};
    const initFlags: Record<string, boolean> = {};
    for (const a of session.answers) {
      if (a.answer != null) initAnswers[a.question_id] = a.answer;
    }
    setAnswers(initAnswers);
    setRemaining(session.remaining_seconds);
    autoSubmittedRef.current = false;
    if (session.status === "submitted") {
      setSubmitResult({ submitted: true });
    }
  }, [session]);

  // Timer countdown
  useEffect(() => {
    if (!session || session.status !== "in_progress" || remaining <= 0)
      return;
    const id = setInterval(() => {
      setRemaining((prev) => Math.max(0, prev - 1));
    }, 1000);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session, remaining <= 0]);

  // Auto-submit when timer expires
  useEffect(() => {
    if (
      !session ||
      session.status !== "in_progress" ||
      remaining > 0 ||
      autoSubmittedRef.current
    )
      return;
    autoSubmittedRef.current = true;
    const doSubmit = async () => {
      const cur = answersRef.current;
      const arr = Object.entries(cur).map(([qid, ans]) => ({
        question_id: qid,
        answer: ans,
      }));
      if (arr.length > 0) {
        try {
          await saveAnswers.mutateAsync(arr);
        } catch {
          /* best-effort */
        }
      }
      submitSession.mutate(undefined, {
        onSuccess: (result) => setSubmitResult(result),
      });
    };
    doSubmit();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [remaining <= 0]);

  // Periodic save every 30s
  useEffect(() => {
    if (!sessionId || session?.status !== "in_progress") return;
    const id = setInterval(() => {
      const cur = answersRef.current;
      const arr = Object.entries(cur).map(([qid, ans]) => ({
        question_id: qid,
        answer: ans,
      }));
      if (arr.length > 0) {
        saveAnswers.mutate(arr);
      }
    }, 30000);
    return () => clearInterval(id);
  }, [sessionId, session?.status, saveAnswers]);

  // Violation logging
  useEffect(() => {
    if (!sessionId || session?.status !== "in_progress") return;
    const onFullscreen = () => {
      if (!document.fullscreenElement)
        logViolation.mutate("fullscreen_exit");
    };
    const onVisibility = () => {
      if (document.hidden) logViolation.mutate("tab_switch");
    };
    const onCopy = () => logViolation.mutate("copy_attempt");
    document.addEventListener("fullscreenchange", onFullscreen);
    document.addEventListener("visibilitychange", onVisibility);
    document.addEventListener("copy", onCopy);
    return () => {
      document.removeEventListener("fullscreenchange", onFullscreen);
      document.removeEventListener("visibilitychange", onVisibility);
      document.removeEventListener("copy", onCopy);
    };
  }, [sessionId, session?.status, logViolation]);

  // Request fullscreen
  const enterFullscreen = useCallback(async () => {
    try {
      if (document.documentElement.requestFullscreen) {
        await document.documentElement.requestFullscreen();
      }
    } catch {
      /* non-critical */
    }
    setFullscreenGranted(true);
  }, []);

  const setAnswer = useCallback((questionId: string, value: string) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
  }, []);

  const toggleFlag = useCallback((questionId: string) => {
    setFlagged((prev) => ({ ...prev, [questionId]: !prev[questionId] }));
  }, []);

  const handleSubmit = useCallback(async () => {
    if (submitting) return;
    setSubmitting(true);
    const cur = answersRef.current;
    const arr = Object.entries(cur).map(([qid, ans]) => ({
      question_id: qid,
      answer: ans,
    }));
    if (arr.length > 0) {
      try {
        await saveAnswers.mutateAsync(arr);
      } catch {
        /* best-effort */
      }
    }
    submitSession.mutate(undefined, {
      onSuccess: (result) => {
        setSubmitResult(result);
        setShowConfirm(false);
        setSubmitting(false);
        if (document.fullscreenElement) {
          document.exitFullscreen().catch(() => {});
        }
      },
      onError: () => {
        setSubmitting(false);
        setShowConfirm(false);
      },
    });
  }, [submitting, saveAnswers, submitSession]);

  // ── Error state (check before !session to handle query error) ────────

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

  // ── Loading state ─────────────────────────────────────────────────────

  if (isLoading || !session) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8">
        <p className="mb-4 text-sm text-ink-500">{t("sys_loading")}</p>
        <Skeleton className="mb-6 h-8 w-2/3" />
        <Skeleton className="mb-4 h-64 w-full rounded-lg" />
        <Skeleton className="h-10 w-32" />
      </div>
    );
  }

  // ── Submitted state ───────────────────────────────────────────────────

  if (submitResult) {
    return (
      <div className="mx-auto max-w-lg px-4 py-16 text-center">
        <CheckCircle2 className="mx-auto mb-4 size-12 text-success" />
        <h1 className="mb-2 text-2xl font-bold text-ink-900">
          {t("submitted")}
        </h1>
        {submitResult.score != null && (
          <p className="text-lg text-ink-700">
            {t("score")}: {submitResult.score}
          </p>
        )}
      </div>
    );
  }

  // ── Fullscreen gate ───────────────────────────────────────────────────

  if (!fullscreenGranted) {
    return (
      <div className="mx-auto flex max-w-md flex-col items-center justify-center px-4 py-24 text-center">
        <Maximize2 className="mb-4 size-12 text-brand-600" />
        <h1 className="mb-4 text-xl font-bold text-ink-900">
          {t("fullscreen_required")}
        </h1>
        <Button onClick={enterFullscreen} data-testid="enter-fullscreen">
          <Maximize2 className="size-4" />
          {t("start_exam")}
        </Button>
      </div>
    );
  }

  // ── Active exam ───────────────────────────────────────────────────────

  const currentQ = allQuestions[currentQIndex];
  const answeredCount = Object.keys(answers).length;
  const isFlagged = currentQ ? flagged[currentQ.id] ?? false : false;
  const timerExpired = remaining <= 0;

  return (
    <div className="mx-auto max-w-4xl px-4 py-4">
      {/* Header */}
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-ink-600">
          <BookOpen className="size-4" />
          <span>
            {t("session_question")} {Math.min(currentQIndex + 1, allQuestions.length)} {t("of")}{" "}
            {allQuestions.length}
          </span>
        </div>
        <div
          className={`rounded-md px-3 py-1 text-lg font-mono font-bold ${
            timerExpired
              ? "bg-danger-bg text-danger"
              : "bg-surface-2 text-ink-900"
          }`}
        >
          {formatTime(remaining)}
        </div>
      </div>

      {/* Question navigator grid */}
      <div className="mb-4 flex flex-wrap gap-1.5">
        {allQuestions.map((q, i) => {
          const hasAnswer = answers[q.id] != null;
          const isFlagQ = flagged[q.id] ?? false;
          const isCurrent = i === currentQIndex;

          let cellClass = "flex size-8 items-center justify-center rounded-md text-xs font-medium transition-colors";
          if (isCurrent) {
            cellClass += " bg-brand-600 text-white";
          } else if (hasAnswer && isFlagQ) {
            cellClass += " border border-warning/30 bg-warning-bg text-warning";
          } else if (hasAnswer) {
            cellClass += " bg-brand-50 text-brand-700";
          } else if (isFlagQ) {
            cellClass += " border border-warning/30 text-warning";
          } else {
            cellClass += " bg-surface-2 text-ink-600 hover:bg-surface-3";
          }

          return (
            <button
              key={q.id}
              type="button"
              onClick={() => setCurrentQIndex(i)}
              className={cellClass}
              data-testid={`session-nav-${i}`}
            >
              {i + 1}
            </button>
          );
        })}
      </div>

      {/* Question card */}
      {currentQ && (
        <Card className="mb-4 p-5">
          <div className="mb-2 text-xs uppercase tracking-wide text-ink-500">
            {t(("fmt_" + currentQ.format) as I18nKey)}
          </div>
          <p className="mb-4 whitespace-pre-wrap text-base text-ink-900">
            {currentQ.body}
          </p>

          {renderAnswerInput(
            currentQ,
            answers[currentQ.id] ?? "",
            (val) => setAnswer(currentQ.id, val),
            timerExpired,
          )}

          {/* Flag toggle */}
          <div className="mt-4 flex items-center gap-2">
            <Button
              type="button"
              variant={isFlagged ? "default" : "outline"}
              size="sm"
              onClick={() => toggleFlag(currentQ.id)}
            >
              <Flag className="size-3.5" />
              {isFlagged ? t("unflag") : t("flag")}
            </Button>
          </div>
        </Card>
      )}

      {/* Navigation buttons */}
      <div className="flex items-center justify-between">
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={currentQIndex === 0}
          onClick={() => setCurrentQIndex((i) => Math.max(0, i - 1))}
        >
          <ChevronLeft className="size-4" />
        </Button>

        <Button
          type="button"
          variant="destructive"
          onClick={() => setShowConfirm(true)}
          disabled={timerExpired || submitting}
        >
          {t("submit")}
        </Button>

        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={currentQIndex >= allQuestions.length - 1}
          onClick={() =>
            setCurrentQIndex((i) =>
              Math.min(allQuestions.length - 1, i + 1),
            )
          }
        >
          <ChevronRight className="size-4" />
        </Button>
      </div>

      {/* Submit confirmation dialog */}
      <Dialog open={showConfirm} onOpenChange={setShowConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("submit_confirm")}</DialogTitle>
            <DialogDescription>
              {answeredCount}/{allQuestions.length} {t("session_question").toLowerCase()}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">{t("cancel")}</Button>
            </DialogClose>
            <Button
              variant="destructive"
              onClick={handleSubmit}
              disabled={submitting}
            >
              {submitting ? t("sys_loading") : t("submit")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function renderAnswerInput(
  question: SessionQuestion,
  currentValue: string,
  onChange: (val: string) => void,
  disabled: boolean,
) {
  const { format, options } = question;

  if (format === "mcq") {
    return (
      <div className="space-y-2">
        {options.map((opt) => (
          <label
            key={opt.key}
            className={`flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors ${
              currentValue === opt.key
                ? "border-brand-500 bg-brand-50"
                : "border-line hover:bg-surface-2"
            } ${disabled ? "cursor-not-allowed opacity-60" : ""}`}
          >
            <input
              type="radio"
              name={`q-${question.id}`}
              value={opt.key}
              checked={currentValue === opt.key}
              onChange={() => onChange(opt.key)}
              disabled={disabled}
              className="size-4 accent-brand-600"
            />
            <span className="text-sm text-ink-800">{opt.text}</span>
          </label>
        ))}
      </div>
    );
  }

  if (format === "multi_answer") {
    const selectedKeys = currentValue
      ? currentValue.split(",").filter(Boolean)
      : [];
    const toggle = (key: string) => {
      const next = selectedKeys.includes(key)
        ? selectedKeys.filter((k) => k !== key)
        : [...selectedKeys, key];
      onChange(next.sort().join(","));
    };
    return (
      <div className="space-y-2">
        {options.map((opt) => (
          <label
            key={opt.key}
            className={`flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors ${
              selectedKeys.includes(opt.key)
                ? "border-brand-500 bg-brand-50"
                : "border-line hover:bg-surface-2"
            } ${disabled ? "cursor-not-allowed opacity-60" : ""}`}
          >
            <input
              type="checkbox"
              checked={selectedKeys.includes(opt.key)}
              onChange={() => toggle(opt.key)}
              disabled={disabled}
              className="size-4 accent-brand-600"
            />
            <span className="text-sm text-ink-800">{opt.text}</span>
          </label>
        ))}
      </div>
    );
  }

  if (format === "short" || format === "fill_blank") {
    return (
      <input
        type="text"
        value={currentValue}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="w-full rounded-lg border border-line bg-background px-3 py-2 text-sm text-ink-900 outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500 disabled:opacity-60"
      />
    );
  }

  if (format === "essay") {
    return (
      <textarea
        value={currentValue}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        rows={5}
        className="w-full resize-y rounded-lg border border-line bg-background px-3 py-2 text-sm text-ink-900 outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500 disabled:opacity-60"
      />
    );
  }

  return null;
}
