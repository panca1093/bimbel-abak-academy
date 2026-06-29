"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSaveQuestion } from "@/lib/hooks/admin-tests";
import { useTranslation } from "@/lib/i18n";
import type {
  AdminQuestionInput,
  AdminQuestionOptionInput,
  QuestionFormat,
  QuestionWithOptions,
} from "@/lib/types";

interface QuestionEditorProps {
  testId: string;
  question?: QuestionWithOptions;
  onCancel: () => void;
  onSaved?: () => void;
}

const DEFAULT_KEYS: Array<"a" | "b" | "c" | "d"> = ["a", "b", "c", "d"];

function nextKey(existing: AdminQuestionOptionInput[]): string {
  const used = new Set(existing.map((o) => o.key.toLowerCase()));
  for (const k of DEFAULT_KEYS) {
    if (!used.has(k)) return k;
  }
  return "x";
}

function OptionEditor({
  format,
  options,
  onChange,
  disabled,
}: {
  format: "mcq" | "multi_answer";
  options: AdminQuestionOptionInput[];
  onChange: (next: AdminQuestionOptionInput[]) => void;
  disabled: boolean;
}) {
  const { t } = useTranslation();
  const isSingle = format === "mcq";

  function update(index: number, patch: Partial<AdminQuestionOptionInput>) {
    onChange(options.map((o, i) => (i === index ? { ...o, ...patch } : o)));
  }

  function setCorrect(index: number, value: boolean) {
    if (isSingle) {
      onChange(options.map((o, i) => ({ ...o, is_correct: i === index })));
      return;
    }
    onChange(options.map((o, i) => (i === index ? { ...o, is_correct: value } : o)));
  }

  function remove(index: number) {
    if (options.length <= 2) return;
    onChange(options.filter((_, i) => i !== index));
  }

  function add() {
    const key = nextKey(options);
    onChange([
      ...options,
      {
        key,
        text: "",
        is_correct: false,
        sort_order: options.length + 1,
      },
    ]);
  }

  return (
    <div className="space-y-2">
      {options.map((opt, index) => (
        <div key={index} className="flex items-center gap-2">
          <div className="w-8 text-sm font-mono uppercase text-muted-foreground">
            {opt.key}
          </div>
          <Input
            aria-label={t("tests_field_option_text")}
            value={opt.text}
            onChange={(e) => update(index, { text: e.target.value })}
            placeholder={t("tests_field_option_text")}
            disabled={disabled}
            className="flex-1"
          />
          <Input
            aria-label={`${t("tests_field_image_url")} ${opt.key}`}
            value={opt.image_url ?? ""}
            onChange={(e) => update(index, { image_url: e.target.value || undefined })}
            placeholder={t("tests_field_image_url")}
            disabled={disabled}
            className="w-40"
          />
          <label className="flex items-center gap-1 text-sm">
            {isSingle ? (
              <input
                type="radio"
                name={`question-correct-${format}`}
                checked={opt.is_correct}
                onChange={() => setCorrect(index, true)}
                disabled={disabled}
                aria-label={t("tests_field_option_is_correct")}
              />
            ) : (
              <input
                type="checkbox"
                checked={opt.is_correct}
                onChange={(e) => setCorrect(index, e.target.checked)}
                disabled={disabled}
                aria-label={t("tests_field_option_is_correct")}
              />
            )}
            <span>{t("tests_field_option_is_correct")}</span>
          </label>
          <Button
            type="button"
            size="icon-xs"
            variant="ghost"
            onClick={() => remove(index)}
            disabled={disabled || options.length <= 2}
            aria-label={t("tests_remove_option")}
          >
            <Trash2 className="size-3" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        size="sm"
        variant="outline"
        onClick={add}
        disabled={disabled || options.length >= 4}
      >
        <Plus className="mr-1 size-4" />
        {t("tests_add_option")}
      </Button>
    </div>
  );
}

const FORMAT_LABELS: Record<QuestionFormat, "tests_format_mcq" | "tests_format_multi_answer" | "tests_format_short" | "tests_format_fill_blank" | "tests_format_essay"> = {
  mcq: "tests_format_mcq",
  multi_answer: "tests_format_multi_answer",
  short: "tests_format_short",
  fill_blank: "tests_format_fill_blank",
  essay: "tests_format_essay",
};

const ALL_FORMATS: QuestionFormat[] = ["mcq", "multi_answer", "short", "fill_blank", "essay"];

function buildOptionsFromQuestion(q: QuestionWithOptions): AdminQuestionOptionInput[] {
  if (q.options.length === 0) {
    return [
      { key: "a", text: "", is_correct: true, sort_order: 1 },
      { key: "b", text: "", is_correct: false, sort_order: 2 },
    ];
  }
  return q.options.map((o) => ({
    key: o.key,
    text: o.text,
    image_url: o.image_url ?? undefined,
    is_correct: o.is_correct,
    sort_order: o.sort_order,
  }));
}

function buildInput(
  format: QuestionFormat,
  body: string,
  sortOrder: string,
  difficulty: string,
  explanation: string,
  imageUrl: string,
  correctAnswer: string,
  options: AdminQuestionOptionInput[]
): AdminQuestionInput {
  const base: AdminQuestionInput = {
    format,
    body: body.trim(),
    sort_order: Number(sortOrder) || 0,
  };
  if (difficulty) base.difficulty = difficulty;
  if (explanation.trim()) base.explanation = explanation.trim();
  if (imageUrl.trim()) base.image_url = imageUrl.trim();
  if (format === "short" || format === "fill_blank") {
    base.correct_answer = correctAnswer.trim();
  }
  if (format === "mcq" || format === "multi_answer") {
    base.options = options.map((o, i) => ({
      key: o.key,
      text: o.text,
      image_url: o.image_url,
      is_correct: o.is_correct,
      sort_order: i + 1,
    }));
  }
  return base;
}

function validate(
  format: QuestionFormat,
  body: string,
  correctAnswer: string,
  options: AdminQuestionOptionInput[]
): { ok: true } | { ok: false; key: string } {
  if (!body.trim()) {
    return { ok: false, key: "tests_validation_body_required" };
  }
  if (format === "mcq") {
    const correct = options.filter((o) => o.is_correct).length;
    if (correct !== 1) return { ok: false, key: "tests_validation_mcq_one_correct" };
  }
  if (format === "multi_answer") {
    const correct = options.filter((o) => o.is_correct).length;
    if (correct < 1) return { ok: false, key: "tests_validation_multi_answer_one_correct" };
  }
  if (format === "short" || format === "fill_blank") {
    if (!correctAnswer.trim()) {
      return { ok: false, key: "tests_validation_correct_answer_required" };
    }
  }
  return { ok: true };
}

export function QuestionEditor({ testId, question, onCancel, onSaved }: QuestionEditorProps) {
  const { t } = useTranslation();
  const isEdit = Boolean(question);
  const [format, setFormat] = useState<QuestionFormat>(question?.question.format ?? "mcq");
  const [body, setBody] = useState(question?.question.body ?? "");
  const [sortOrder, setSortOrder] = useState(String(question?.question.sort_order ?? 1));
  const [difficulty, setDifficulty] = useState<string>(question?.question.difficulty ?? "");
  const [explanation, setExplanation] = useState(question?.question.explanation ?? "");
  const [imageUrl, setImageUrl] = useState(question?.question.image_url ?? "");
  const [correctAnswer, setCorrectAnswer] = useState(question?.question.correct_answer ?? "");
  const [options, setOptions] = useState<AdminQuestionOptionInput[]>(
    question ? buildOptionsFromQuestion(question) : [
      { key: "a", text: "", is_correct: true, sort_order: 1 },
      { key: "b", text: "", is_correct: false, sort_order: 2 },
    ]
  );
  const [errorKey, setErrorKey] = useState<string | null>(null);
  const save = useSaveQuestion(testId);

  useEffect(() => {
    if (!question) {
      setFormat("mcq");
      setBody("");
      setSortOrder("1");
      setDifficulty("");
      setExplanation("");
      setImageUrl("");
      setCorrectAnswer("");
      setOptions([
        { key: "a", text: "", is_correct: true, sort_order: 1 },
        { key: "b", text: "", is_correct: false, sort_order: 2 },
      ]);
    }
  }, [question]);

  async function handleSave() {
    const result = validate(format, body, correctAnswer, options);
    if (!result.ok) {
      setErrorKey(result.key);
      return;
    }
    setErrorKey(null);
    const input = buildInput(format, body, sortOrder, difficulty, explanation, imageUrl, correctAnswer, options);
    try {
      await save.mutateAsync({ question: question?.question.id, input });
      toast.success(t("tests_save_success"));
      onSaved?.();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error_generic"));
    }
  }

  const showOptions = format === "mcq" || format === "multi_answer";
  const showCorrectAnswer = format === "short" || format === "fill_blank";
  const errorMessage = errorKey ? t(errorKey as Parameters<typeof t>[0]) : null;

  return (
    <div className="space-y-4 rounded-lg border bg-card p-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="grid gap-2">
          <Label htmlFor="question-format">{t("format")}</Label>
          <select
            id="question-format"
            data-slot="select"
            value={format}
            onChange={(e) => setFormat(e.target.value as QuestionFormat)}
            disabled={save.isPending}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50 disabled:pointer-events-none disabled:opacity-50"
          >
            {ALL_FORMATS.map((f) => (
              <option key={f} value={f}>
                {t(FORMAT_LABELS[f])}
              </option>
            ))}
          </select>
        </div>
        <div className="grid gap-2">
          <Label htmlFor="question-sort-order">{t("tests_field_sort_order")}</Label>
          <Input
            id="question-sort-order"
            type="number"
            min={1}
            value={sortOrder}
            onChange={(e) => setSortOrder(e.target.value)}
            disabled={save.isPending}
          />
        </div>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="question-body">{t("tests_field_body")}</Label>
        <textarea
          id="question-body"
          data-slot="textarea"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={3}
          disabled={save.isPending}
          className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50 disabled:pointer-events-none disabled:opacity-50"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="grid gap-2">
          <Label htmlFor="question-difficulty">{t("difficulty")}</Label>
          <select
            id="question-difficulty"
            value={difficulty || "none"}
            onChange={(e) => setDifficulty(e.target.value === "none" ? "" : e.target.value)}
            disabled={save.isPending}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50 disabled:pointer-events-none disabled:opacity-50"
          >
            <option value="none">—</option>
            <option value="easy">{t("tests_field_difficulty_easy")}</option>
            <option value="medium">{t("tests_field_difficulty_medium")}</option>
            <option value="hard">{t("tests_field_difficulty_hard")}</option>
          </select>
        </div>
        <div className="grid gap-2">
          <Label htmlFor="question-image-url">{t("tests_field_image_url")}</Label>
          <Input
            id="question-image-url"
            value={imageUrl}
            onChange={(e) => setImageUrl(e.target.value)}
            placeholder="https://..."
            disabled={save.isPending}
          />
        </div>
      </div>

      {showOptions && (
        <div className="grid gap-2">
          <Label>{t("tests_field_option_text")}</Label>
          <OptionEditor
            format={format as "mcq" | "multi_answer"}
            options={options}
            onChange={setOptions}
            disabled={save.isPending}
          />
        </div>
      )}

      {showCorrectAnswer && (
        <div className="grid gap-2">
          <Label htmlFor="question-correct-answer">{t("tests_field_correct_answer")}</Label>
          <Input
            id="question-correct-answer"
            value={correctAnswer}
            onChange={(e) => setCorrectAnswer(e.target.value)}
            disabled={save.isPending}
          />
        </div>
      )}

      <div className="grid gap-2">
        <Label htmlFor="question-explanation">{t("tests_field_explanation")}</Label>
        <textarea
          id="question-explanation"
          value={explanation}
          onChange={(e) => setExplanation(e.target.value)}
          rows={2}
          disabled={save.isPending}
          className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50 disabled:pointer-events-none disabled:opacity-50"
        />
      </div>

      {errorMessage && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage}
        </p>
      )}

      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={save.isPending}
        >
          {t("cancel")}
        </Button>
        <Button type="button" onClick={handleSave} disabled={save.isPending}>
          {save.isPending ? t("saving") : t("tests_save_question")}
        </Button>
      </div>
    </div>
  );
}
