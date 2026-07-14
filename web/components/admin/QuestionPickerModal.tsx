"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Search } from "lucide-react";
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

const ALL_FORMATS: QuestionFormat[] = ["mcq", "multi_answer", "short", "fill_blank", "essay"];

interface QuestionPickerModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  testId: string;
  attached: QuestionWithOptions[];
  onAttach: (questionIds: string[]) => Promise<void>;
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
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("tests_picker_title")}</DialogTitle>
          <DialogDescription>{t("tests_picker_search")}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="flex flex-col gap-3 sm:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t("tests_picker_search")}
                className="pl-9"
              />
            </div>
            <select
              data-slot="select"
              value={format}
              onChange={(e) => setFormat(e.target.value as QuestionFormat | "all")}
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            >
              <option value="all">{t("tests_picker_format_all")}</option>
              {ALL_FORMATS.map((f) => (
                <option key={f} value={f}>
                  {t(FORMAT_LABELS[f] as Parameters<typeof t>[0])}
                </option>
              ))}
            </select>
            <select
              data-slot="select"
              value={topicId}
              onChange={(e) => setTopicId(e.target.value)}
              disabled={topics.isLoading}
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            >
              <option value="all">{t("tests_picker_topic_all")}</option>
              {topicOptions.map((topic) => (
                <option key={topic.id} value={topic.id}>
                  {topic.name}
                </option>
              ))}
            </select>
          </div>

          <div className="max-h-[360px] overflow-y-auto rounded-lg border">
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
                <label
                  key={row.question.id}
                  className={`flex items-center gap-3 border-b p-3 last:border-b-0 ${
                    isAttached ? "bg-muted/40" : "hover:bg-muted/20"
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    disabled={isAttached}
                    onChange={() => toggle(row.question.id)}
                    className="size-4"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 text-sm">
                      <Badge variant="outline">{t(FORMAT_LABELS[row.question.format] as Parameters<typeof t>[0])}</Badge>
                      <span className="text-muted-foreground">{row.question.topic || "—"}</span>
                    </div>
                    <p className="mt-1 truncate text-sm">{stripHtmlToPlainText(row.question.body)}</p>
                  </div>
                  {isAttached && (
                    <Badge variant="secondary">{t("tests_picker_attached_badge")}</Badge>
                  )}
                </label>
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
