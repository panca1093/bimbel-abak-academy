"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import {
  ArrowDown,
  ArrowUp,
  ClipboardList,
  Library,
  Plus,
  Unlink,
} from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { QuestionEditor } from "@/components/admin/QuestionEditor";
import { QuestionPickerModal } from "@/components/admin/QuestionPickerModal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
  useTestDetail,
  useTestQuestions,
  useUpdateTest,
  useDetachQuestion,
  useReorderTestQuestions,
  useAttachQuestions,
} from "@/lib/hooks/admin-tests";
import { useTranslation } from "@/lib/i18n";
import { stripHtmlToPlainText } from "@/lib/rich-text";
import type { QuestionWithOptions, QuestionFormat } from "@/lib/types";

const FORMAT_BADGE: Record<
  QuestionFormat,
  | "tests_format_mcq"
  | "tests_format_multi_answer"
  | "tests_format_short"
  | "tests_format_fill_blank"
  | "tests_format_essay"
> = {
  mcq: "tests_format_mcq",
  multi_answer: "tests_format_multi_answer",
  short: "tests_format_short",
  fill_blank: "tests_format_fill_blank",
  essay: "tests_format_essay",
};

const SECTION_TYPES: Array<{ value: string; labelKey: string }> = [
  { value: "", labelKey: "tests_field_section_type_none" },
  { value: "listening", labelKey: "tests_field_section_type_listening" },
  { value: "reading", labelKey: "tests_field_section_type_reading" },
  { value: "writing", labelKey: "tests_field_section_type_writing" },
];

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

function useSyncedTestForm(test: { test: { title: string; subject: string; topic: string; duration_minutes: number; audio_url?: string | null; audio_play_limit?: number | null; section_type?: string | null } } | undefined) {
  const [title, setTitle] = useState("");
  const [subject, setSubject] = useState("");
  const [topic, setTopic] = useState("");
  const [duration, setDuration] = useState("");
  const [audioUrl, setAudioUrl] = useState("");
  const [audioPlayLimit, setAudioPlayLimit] = useState("");
  const [sectionType, setSectionType] = useState<string>("");

  useEffect(() => {
    if (!test) return;
    const t = test.test;
    setTitle(t.title ?? "");
    setSubject(t.subject ?? "");
    setTopic(t.topic ?? "");
    setDuration(t.duration_minutes != null ? String(t.duration_minutes) : "");
    setAudioUrl(t.audio_url ?? "");
    setAudioPlayLimit(t.audio_play_limit != null ? String(t.audio_play_limit) : "");
    setSectionType(t.section_type ?? "");
  }, [test]);

  return {
    title,
    setTitle,
    subject,
    setSubject,
    topic,
    setTopic,
    duration,
    setDuration,
    audioUrl,
    setAudioUrl,
    audioPlayLimit,
    setAudioPlayLimit,
    sectionType,
    setSectionType,
  };
}

function QuestionRow({
  question,
  index,
  total,
  onReorder,
  onDetach,
}: {
  question: QuestionWithOptions;
  index: number;
  total: number;
  onReorder: (questionId: string, direction: "up" | "down") => void;
  onDetach: (questionId: string) => void;
}) {
  const { t } = useTranslation();
  const isFirst = index === 0;
  const isLast = index === total - 1;

  return (
    <div data-question-row className="flex items-center gap-3 rounded-lg border bg-card p-3">
      <span className="w-6 text-xs text-muted-foreground">#{index + 1}</span>
      <Badge variant="outline">{t(FORMAT_BADGE[question.question.format])}</Badge>
      <span className="flex-1 truncate text-sm">{stripHtmlToPlainText(question.question.body)}</span>
      <div className="flex items-center gap-1">
        <Button
          type="button"
          size="icon-xs"
          variant="ghost"
          onClick={() => onReorder(question.question.id, "up")}
          disabled={isFirst}
          aria-label={t("tests_reorder_up")}
        >
          <ArrowUp className="size-3" />
        </Button>
        <Button
          type="button"
          size="icon-xs"
          variant="ghost"
          onClick={() => onReorder(question.question.id, "down")}
          disabled={isLast}
          aria-label={t("tests_reorder_down")}
        >
          <ArrowDown className="size-3" />
        </Button>
        <Button
          type="button"
          size="icon-xs"
          variant="ghost"
          onClick={() => onDetach(question.question.id)}
          aria-label={t("tests_detach_question")}
        >
          <Unlink className="size-3" />
        </Button>
      </div>
    </div>
  );
}

export default function TestDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";
  const { t } = useTranslation();
  const [creating, setCreating] = useState(false);
  const [pickerOpen, setPickerOpen] = useState(false);

  const {
    data: detail,
    isLoading: detailLoading,
    isError: detailError,
    error: detailErr,
  } = useTestDetail(id);
  const {
    data: questionsResp,
    isLoading: questionsLoading,
    isError: questionsError,
    error: questionsErr,
  } = useTestQuestions(id);

  const update = useUpdateTest(id);
  const detach = useDetachQuestion(id);
  const reorder = useReorderTestQuestions(id);
  const attach = useAttachQuestions(id);

  const form = useSyncedTestForm(detail);

  const isLoading = detailLoading || questionsLoading;
  const isError = detailError || questionsError;
  const error = detailErr ?? questionsErr;
  const questions = questionsResp?.data ?? detail?.questions ?? [];

  const questionIds = useMemo(() => questions.map((q) => q.question.id), [questions]);

  const listeningRequiresAudio = form.sectionType === "listening" && form.audioUrl.trim() === "";
  const canSaveTest =
    form.title.trim() !== "" &&
    form.subject.trim() !== "" &&
    form.topic.trim() !== "" &&
    form.duration !== "" &&
    Number(form.duration) > 0 &&
    !listeningRequiresAudio &&
    !update.isPending;

  async function handleSaveTest(e: React.FormEvent) {
    e.preventDefault();
    if (!canSaveTest || !id) return;
    const payload: Record<string, unknown> = {
      title: form.title.trim(),
      subject: form.subject.trim(),
      topic: form.topic.trim(),
      duration_minutes: Number(form.duration),
    };
    // Explicit null (not an omitted key) when cleared, so the backend can tell
    // "clear this field" apart from "field not sent" — the form is synced from
    // the existing test on load, so an empty value here means the user cleared it.
    payload.audio_url = form.audioUrl.trim() || null;
    payload.audio_play_limit = form.audioPlayLimit !== "" ? Number(form.audioPlayLimit) : null;
    payload.section_type = form.sectionType || null;
    try {
      await update.mutateAsync(payload);
      toast.success(t("tests_update_success"));
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  async function handleReorder(questionId: string, direction: "up" | "down") {
    const idx = questionIds.indexOf(questionId);
    if (idx < 0) return;
    const targetIdx = direction === "up" ? idx - 1 : idx + 1;
    if (targetIdx < 0 || targetIdx >= questionIds.length) return;
    const next = [...questionIds];
    [next[idx], next[targetIdx]] = [next[targetIdx], next[idx]];
    try {
      await reorder.mutateAsync({ question_ids: next });
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  async function handleDetach(questionId: string) {
    if (!confirm(t("tests_confirm_delete_question"))) return;
    try {
      await detach.mutateAsync(questionId);
      toast.success(t("tests_save_success"));
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  async function handleAttach(selectedIds: string[]) {
    await attach.mutateAsync({ question_ids: selectedIds });
  }

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
        <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[360px_1fr]">
          {/* Left: sticky test details form */}
          <form
            onSubmit={handleSaveTest}
            className="space-y-4 rounded-lg border bg-card p-4 lg:sticky lg:top-6 lg:self-start"
          >
            <h2 className="text-lg font-semibold">{t("tests_field_title")}</h2>

            <div className="grid gap-2">
              <Label htmlFor="test-title">{t("tests_field_title")}</Label>
              <Input
                id="test-title"
                value={form.title}
                onChange={(e) => form.setTitle(e.target.value)}
                placeholder={t("tests_field_title")}
                disabled={update.isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="grid gap-2">
                <Label htmlFor="test-subject">{t("tests_field_subject")}</Label>
                <Input
                  id="test-subject"
                  value={form.subject}
                  onChange={(e) => form.setSubject(e.target.value)}
                  placeholder={t("tests_field_subject")}
                  disabled={update.isPending}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="test-topic">{t("tests_field_topic")}</Label>
                <Input
                  id="test-topic"
                  value={form.topic}
                  onChange={(e) => form.setTopic(e.target.value)}
                  placeholder={t("tests_field_topic")}
                  disabled={update.isPending}
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="test-duration">{t("tests_field_duration")}</Label>
              <Input
                id="test-duration"
                type="number"
                min={1}
                value={form.duration}
                onChange={(e) => form.setDuration(e.target.value)}
                placeholder="60"
                disabled={update.isPending}
              />
            </div>

            <div className="grid gap-2">
              <Label>{t("tests_field_section_type")}</Label>
              <div className="grid grid-cols-2 gap-2">
                {SECTION_TYPES.map(({ value, labelKey }) => (
                  <label
                    key={value || "none"}
                    className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm"
                  >
                    <input
                      type="radio"
                      name="section_type"
                      value={value}
                      checked={form.sectionType === value}
                      onChange={() => form.setSectionType(value)}
                      disabled={update.isPending}
                    />
                    <span>{t(labelKey as Parameters<typeof t>[0])}</span>
                  </label>
                ))}
              </div>
              {form.sectionType === "listening" && (
                <p className="text-xs text-muted-foreground">
                  URL audio wajib diisi untuk sesi Listening.
                </p>
              )}
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="grid gap-2">
                <Label htmlFor="test-audio-url">{t("tests_field_audio_url")}</Label>
                <Input
                  id="test-audio-url"
                  value={form.audioUrl}
                  onChange={(e) => form.setAudioUrl(e.target.value)}
                  placeholder="https://..."
                  disabled={update.isPending}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="test-audio-play-limit">{t("tests_field_audio_play_limit")}</Label>
                <Input
                  id="test-audio-play-limit"
                  type="number"
                  min={0}
                  value={form.audioPlayLimit}
                  onChange={(e) => form.setAudioPlayLimit(e.target.value)}
                  placeholder="0"
                  disabled={update.isPending}
                />
              </div>
            </div>

            <Button type="submit" disabled={!canSaveTest} className="w-full">
              {update.isPending ? t("saving") : t("save")}
            </Button>
          </form>

          {/* Right: questions panel */}
          <div className="space-y-4">
            <div className="flex items-center justify-between gap-2">
              <h2 className="text-lg font-semibold">
                {t("tests_question_count")} ({questions.length})
              </h2>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setPickerOpen(true)}
                >
                  <Library className="mr-1 size-4" />
                  {t("tests_from_bank")}
                </Button>
                <Button
                  type="button"
                  size="sm"
                  onClick={() => setCreating(true)}
                  disabled={creating}
                >
                  <Plus className="mr-1 size-4" />
                  {t("tests_new_question")}
                </Button>
              </div>
            </div>

            {creating && (
              <QuestionEditor
                testId={id}
                onCancel={() => setCreating(false)}
                onSaved={() => setCreating(false)}
              />
            )}

            {questions.length === 0 ? (
              <div className="rounded-lg border p-8 text-center text-muted-foreground">
                {t("tests_questions_empty")}
              </div>
            ) : (
              <div className="space-y-2">
                {questions.map((q, idx) => (
                  <QuestionRow
                    key={q.question.id}
                    question={q}
                    index={idx}
                    total={questions.length}
                    onReorder={handleReorder}
                    onDetach={handleDetach}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      <QuestionPickerModal
        open={pickerOpen}
        onOpenChange={setPickerOpen}
        testId={id}
        attached={questions}
        onAttach={handleAttach}
      />
    </div>
  );
}
