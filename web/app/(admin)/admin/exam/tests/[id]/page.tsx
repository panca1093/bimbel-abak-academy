"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { ChevronDown, ChevronUp, ClipboardList, Plus, Trash2 } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { QuestionEditor } from "@/components/admin/QuestionEditor";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
  useTestDetail,
  useTestQuestions,
  useDeleteQuestion,
} from "@/lib/hooks/admin-tests";
import { useTranslation } from "@/lib/i18n";
import type { QuestionWithOptions, QuestionFormat } from "@/lib/types";

const FORMAT_BADGE: Record<QuestionFormat, "tests_format_mcq" | "tests_format_multi_answer" | "tests_format_short" | "tests_format_fill_blank" | "tests_format_essay"> = {
  mcq: "tests_format_mcq",
  multi_answer: "tests_format_multi_answer",
  short: "tests_format_short",
  fill_blank: "tests_format_fill_blank",
  essay: "tests_format_essay",
};

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

function QuestionRow({
  question,
  testId,
  expanded,
  onToggle,
  onSaved,
}: {
  question: QuestionWithOptions;
  testId: string;
  expanded: boolean;
  onToggle: () => void;
  onSaved: () => void;
}) {
  const { t } = useTranslation();
  const remove = useDeleteQuestion(testId);

  async function handleDelete() {
    if (!confirm(t("tests_confirm_delete_question"))) return;
    try {
      await remove.mutateAsync(question.question.id);
      toast.success(t("tests_delete_success"));
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  return (
    <div data-question-row className="rounded-lg border bg-card">
      <div className="flex items-center justify-between gap-2 p-3">
        <button
          type="button"
          onClick={onToggle}
          className="flex flex-1 items-center gap-3 text-left min-w-0"
          aria-expanded={expanded}
        >
          <span className="text-xs text-muted-foreground">#{question.question.sort_order}</span>
          <Badge variant="outline">{t(FORMAT_BADGE[question.question.format])}</Badge>
          <span className="flex-1 truncate text-sm">{question.question.body}</span>
          {expanded ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
        </button>
        <Button
          type="button"
          size="icon-xs"
          variant="ghost"
          onClick={handleDelete}
          disabled={remove.isPending}
          aria-label={t("action_delete")}
        >
          <Trash2 className="size-3" />
        </Button>
      </div>
      {expanded && (
        <div className="border-t p-3">
          <QuestionEditor
            testId={testId}
            question={question}
            onCancel={onToggle}
            onSaved={onToggle}
          />
        </div>
      )}
    </div>
  );
}

export default function TestDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";
  const { t } = useTranslation();
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  const { data: detail, isLoading: detailLoading, isError: detailError, error: detailErr } =
    useTestDetail(id);
  const { data: questionsResp, isLoading: questionsLoading, isError: questionsError, error: questionsErr } =
    useTestQuestions(id);

  const isLoading = detailLoading || questionsLoading;
  const isError = detailError || questionsError;
  const error = detailErr ?? questionsErr;

  const questions = questionsResp?.data ?? detail?.questions ?? [];

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={ClipboardList}
        title={t("tests_detail_page_title")}
        description={
          detail
            ? `${detail.test.subject} · ${detail.test.topic} · ${detail.test.duration_minutes} min`
            : undefined
        }
      />

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-14 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          {errorMessage(error, t("error_generic"))}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="space-y-3">
          {creating && (
            <QuestionEditor
              testId={id}
              onCancel={() => setCreating(false)}
              onSaved={() => setCreating(false)}
            />
          )}

          {!creating && (
            <Button type="button" onClick={() => setCreating(true)}>
              <Plus className="mr-1 size-4" />
              {t("tests_add_question")}
            </Button>
          )}

          {questions.length === 0 ? (
            <div className="rounded-lg border p-8 text-center text-muted-foreground">
              {t("tests_questions_empty")}
            </div>
          ) : (
            <div className="space-y-2">
              {questions.map((q) => (
                <QuestionRow
                  key={q.question.id}
                  question={q}
                  testId={id}
                  expanded={expandedId === q.question.id}
                  onToggle={() =>
                    setExpandedId((prev) => (prev === q.question.id ? null : q.question.id))
                  }
                  onSaved={() => setExpandedId(null)}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
