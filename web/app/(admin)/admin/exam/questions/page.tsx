"use client";

import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Library, Plus, Search, Upload } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { QuestionEditor } from "@/components/admin/QuestionEditor";
import { QuestionPreview } from "@/components/admin/QuestionPreview";
import { TopicsModal } from "@/components/admin/TopicsModal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { useBankQuestions } from "@/lib/hooks/admin-bank-questions";
import { useTopics } from "@/lib/hooks/admin-topics";
import { useTranslation } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { BankQuestionListItem, QuestionFormat } from "@/lib/types";

const ALL_FORMATS: Array<QuestionFormat | "all"> = [
  "all",
  "mcq",
  "multi_answer",
  "short",
  "fill_blank",
  "essay",
];

const FORMAT_LABELS: Record<QuestionFormat | "all", string> = {
  all: "tab_all",
  mcq: "fmt_mcq",
  multi_answer: "fmt_multi_answer",
  short: "fmt_short",
  fill_blank: "fmt_fill_blank",
  essay: "fmt_essay",
};

const DIFFICULTY_LABELS: Record<string, "diff_easy" | "diff_medium" | "diff_hard"> = {
  easy: "diff_easy",
  medium: "diff_medium",
  hard: "diff_hard",
};

export default function QuestionBankPage() {
  const { t } = useTranslation();
  const [format, setFormat] = useState<QuestionFormat | "all">("all");
  const [topicId, setTopicId] = useState<string>("");
  const [search, setSearch] = useState("");

  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewItem, setPreviewItem] = useState<BankQuestionListItem | null>(null);

  const [editorOpen, setEditorOpen] = useState(false);
  const [editorItem, setEditorItem] = useState<BankQuestionListItem | null>(null);

  const [topicsOpen, setTopicsOpen] = useState(false);

  const filters = useMemo(
    () => ({
      format: format === "all" ? undefined : format,
      topic_id: topicId || undefined,
      search: search.trim() || undefined,
      limit: 25,
    }),
    [format, topicId, search]
  );

  const bank = useBankQuestions(filters);
  const topics = useTopics();

  const rows = bank.data?.data ?? [];
  const topicOptions = topics.data?.data ?? [];

  function openCreate() {
    setEditorItem(null);
    setEditorOpen(true);
  }

  function openEdit(item: BankQuestionListItem) {
    setEditorItem(item);
    setEditorOpen(true);
  }

  function openPreview(item: BankQuestionListItem) {
    setPreviewItem(item);
    setPreviewOpen(true);
  }

  function handleRowClick(item: BankQuestionListItem) {
    openPreview(item);
  }

  function handleEditFromPreview() {
    setPreviewOpen(false);
    if (previewItem) {
      setEditorItem(previewItem);
      setEditorOpen(true);
    }
  }

  function handleSaved() {
    setEditorOpen(false);
    setEditorItem(null);
    toast.success(t("tests_save_success"));
  }

  function handleCsvClick() {
    toast.info(t("maint_default_desc"));
  }

  function errorMessage(err: unknown): string {
    if (err instanceof Error) return err.message;
    return t("error_generic");
  }

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={Library}
        title={t("question_bank_title")}
        description={t("question_bank_subtitle")}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" onClick={() => setTopicsOpen(true)}>
              {t("manage_topics")}
            </Button>
            <Button variant="outline" onClick={handleCsvClick}>
              <Upload className="mr-1 size-4" />
              CSV
            </Button>
            <Button onClick={openCreate}>
              <Plus className="mr-1 size-4" />
              {t("create")}
            </Button>
          </div>
        }
      />

      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-wrap gap-2">
          {ALL_FORMATS.map((f) => {
            const active = format === f;
            return (
              <button
                key={f}
                type="button"
                onClick={() => setFormat(f)}
                className={cn(
                  "admin-shell md-chip cursor-pointer border-none",
                  active && "admin-shell md-chip-primary"
                )}
              >
                {t(FORMAT_LABELS[f] as Parameters<typeof t>[0])}
              </button>
            );
          })}
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <select
            data-slot="select"
            value={topicId}
            onChange={(e) => setTopicId(e.target.value)}
            disabled={topics.isLoading}
            className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
          >
            <option value="">{t("all_topics")}</option>
            {topicOptions.map((topic) => (
              <option key={topic.id} value={topic.id}>
                {topic.name}
              </option>
            ))}
          </select>

          <div className="relative w-full sm:w-64">
            <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t("search")}
              className="pl-9"
            />
          </div>
        </div>
      </div>

      {bank.isLoading && !bank.data && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {bank.isError && !bank.data && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          {errorMessage(bank.error)}
        </div>
      )}

      {!bank.isLoading && !bank.isError && (
        <div className="overflow-x-auto md-card-outlined">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">ID</th>
                <th className="px-4 py-3 text-left font-medium">{t("question")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("used_in")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("topic")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("format")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("difficulty")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("points")}</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((item) => {
                const { question } = item;
                const difficultyKey = question.difficulty
                  ? DIFFICULTY_LABELS[question.difficulty]
                  : undefined;
                return (
                  <tr
                    key={question.id}
                    onClick={() => handleRowClick(item)}
                    className="border-t transition-colors hover:bg-muted/40 cursor-pointer"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {question.id.slice(0, 8)}
                    </td>
                    <td className="px-4 py-3 max-w-xs truncate">{question.body}</td>
                    <td className="px-4 py-3">{item.attached_count}</td>
                    <td className="px-4 py-3">{question.topic || "—"}</td>
                    <td className="px-4 py-3">
                      <Badge variant="outline">
                        {t(FORMAT_LABELS[question.format] as Parameters<typeof t>[0])}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      {difficultyKey ? t(difficultyKey) : "—"}
                    </td>
                    <td className="px-4 py-3">
                      {question.point_correct}/{question.point_wrong}
                    </td>
                  </tr>
                );
              })}
              {rows.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">
                    {t("tests_picker_empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <QuestionPreview
        item={previewItem}
        open={previewOpen}
        onOpenChange={setPreviewOpen}
        onEdit={handleEditFromPreview}
      />

      <TopicsModal open={topicsOpen} onOpenChange={setTopicsOpen} />

      {editorOpen && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 sm:p-8">
          <div className="w-full max-w-4xl rounded-lg bg-background p-4 shadow-lg">
            <QuestionEditor
              question={editorItem ?? undefined}
              onCancel={() => {
                setEditorOpen(false);
                setEditorItem(null);
              }}
              onSaved={handleSaved}
            />
          </div>
        </div>
      )}
    </div>
  );
}
