"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Check, Search } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useBankQuestions } from "@/lib/hooks/admin-bank-questions";
import { useTopics } from "@/lib/hooks/admin-topics";
import { useTranslation } from "@/lib/i18n";
import { stripHtmlToPlainText } from "@/lib/rich-text";
import type { QuestionFormat, QuestionWithOptions } from "@/lib/types";

const FORMAT_LABELS: Record<QuestionFormat, string> = {
  mcq: "fmt_mcq",
  multi_answer: "fmt_multi_answer",
  short: "fmt_short",
  fill_blank: "fmt_fill_blank",
  essay: "fmt_essay",
};

const DIFFICULTY_LABEL: Record<string, string> = {
  easy: "diff_easy",
  medium: "diff_medium",
  hard: "diff_hard",
};

const ALL_FORMATS: QuestionFormat[] = ["mcq", "multi_answer", "short", "fill_blank", "essay"];

interface QuestionPickerModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  testId: string;
  attached: QuestionWithOptions[];
  onAttach: (questionIds: string[]) => Promise<void>;
}

function FilterChip({
  active,
  disabled,
  onClick,
  children,
}: {
  active: boolean;
  disabled?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-pressed={active}
      className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
        active
          ? "border-primary bg-primary text-primary-foreground"
          : "border-input bg-background text-ink-900 hover:bg-muted/40"
      } disabled:opacity-50`}
    >
      {children}
    </button>
  );
}

function SelectionBox({ selected }: { selected: boolean }) {
  return (
    <span
      aria-hidden="true"
      className={`flex size-5 flex-shrink-0 items-center justify-center rounded-md border ${
        selected
          ? "border-primary bg-primary text-primary-foreground"
          : "border-input bg-surface"
      }`}
    >
      {selected && <Check className="size-3" />}
    </span>
  );
}

export function QuestionPickerModal({
  open,
  onOpenChange,
  attached,
  onAttach,
}: QuestionPickerModalProps) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [format, setFormat] = useState<QuestionFormat | "all">("all");
  const [topicId, setTopicId] = useState<string>("all");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [isAttaching, setIsAttaching] = useState(false);

  const filters = useMemo(
    () => ({
      search: search.trim() || undefined,
      format: format === "all" ? undefined : format,
      topic_id: topicId === "all" ? undefined : topicId,
      limit: 50,
    }),
    [search, format, topicId]
  );

  const bank = useBankQuestions(filters);
  const topics = useTopics();

  const attachedIds = useMemo(
    () => new Set(attached.map((a) => a.question.id)),
    [attached]
  );

  useEffect(() => {
    if (!open) {
      setSearch("");
      setFormat("all");
      setTopicId("all");
      setSelected(new Set());
    }
  }, [open]);

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  async function handleAttach() {
    if (selected.size === 0) return;
    setIsAttaching(true);
    try {
      await onAttach(Array.from(selected));
      toast.success(t("tests_save_success"));
      onOpenChange(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error_generic"));
    } finally {
      setIsAttaching(false);
    }
  }

  const rows = bank.data?.data ?? [];
  const topicOptions = topics.data?.data ?? [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[720px]">
        <DialogHeader>
          <DialogTitle>{t("tests_picker_title")}</DialogTitle>
          <DialogDescription>{t("tests_picker_search")}</DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-1">
          <div className="flex h-9 items-center gap-2 rounded-lg border border-input bg-surface-2 px-3 focus-within:ring-2 focus-within:ring-ring/40">
            <Search className="size-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t("tests_picker_search")}
              className="h-8 flex-1 border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
            />
          </div>

          <div className="flex flex-wrap items-center gap-1.5">
            <FilterChip active={format === "all"} onClick={() => setFormat("all")}>
              {t("tab_all")}
            </FilterChip>
            {ALL_FORMATS.map((f) => (
              <FilterChip
                key={f}
                active={format === f}
                onClick={() => setFormat(f)}
              >
                {t(FORMAT_LABELS[f] as Parameters<typeof t>[0])}
              </FilterChip>
            ))}
            <select
              data-slot="select"
              value={topicId}
              onChange={(e) => setTopicId(e.target.value)}
              disabled={topics.isLoading}
              className="ml-auto h-7 rounded-full border border-input bg-background px-3 text-xs"
            >
              <option value="all">{t("tests_picker_topic_all")}</option>
              {topicOptions.map((topic) => (
                <option key={topic.id} value={topic.id}>
                  {topic.name}
                </option>
              ))}
            </select>
          </div>

          <div className="max-h-[320px] overflow-y-auto overflow-x-hidden rounded-md border border-input">
            {bank.isLoading && (
              <div className="p-4 text-center text-muted-foreground">{t("sys_loading")}</div>
            )}
            {!bank.isLoading && rows.length === 0 && (
              <div className="p-4 text-center text-muted-foreground">{t("tests_picker_empty")}</div>
            )}
            {rows.map((row) => {
              const isAttached = attachedIds.has(row.question.id);
              const isSelected = selected.has(row.question.id);
              return (
                <button
                  type="button"
                  key={row.question.id}
                  disabled={isAttached}
                  onClick={() => toggle(row.question.id)}
                  className={`flex w-full items-center gap-3 overflow-hidden border-b px-3 py-2 text-left transition-colors last:border-b-0 ${
                    isAttached
                      ? "cursor-default opacity-50"
                      : isSelected
                        ? "bg-green-50 hover:bg-green-50"
                        : "hover:bg-muted/30"
                  }`}
                >
                  <SelectionBox selected={isSelected} />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm text-ink-900">
                      {stripHtmlToPlainText(row.question.body)}
                    </p>
                    <div className="mt-1 flex flex-wrap items-center gap-1.5">
                      <span className="max-w-[140px] truncate rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                        {row.question.id}
                      </span>
                      <Badge variant="outline">
                        {t(FORMAT_LABELS[row.question.format] as Parameters<typeof t>[0])}
                      </Badge>
                      {row.question.topic && (
                        <span className="max-w-[160px] truncate rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                          {row.question.topic}
                        </span>
                      )}
                      {row.question.difficulty &&
                        DIFFICULTY_LABEL[row.question.difficulty] && (
                          <Badge variant="secondary">
                            {t(DIFFICULTY_LABEL[row.question.difficulty] as Parameters<typeof t>[0])}
                          </Badge>
                        )}
                    </div>
                  </div>
                  {isAttached && (
                    <span className="flex-shrink-0 text-[11px] text-muted-foreground">
                      {t("tests_picker_attached_badge")}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          <p className="text-sm text-muted-foreground">
            {t("tests_picker_selected").replace("{n}", String(selected.size))}
          </p>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isAttaching}>
            {t("cancel")}
          </Button>
          <Button type="button" onClick={handleAttach} disabled={selected.size === 0 || isAttaching}>
            {isAttaching
              ? t("saving")
              : t("tests_picker_confirm").replace("{n}", String(selected.size))}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
