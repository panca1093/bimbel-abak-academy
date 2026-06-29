"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useTranslation } from "@/lib/i18n";
import { useCreateExam, useUpdateExam } from "@/lib/hooks/admin-exams";
import type { ExamListItem, CreateExamPayload, UpdateExamPayload } from "@/lib/types";

interface ExamModalProps {
  open: boolean;
  onClose: () => void;
  exam?: ExamListItem | null;
  onSaved?: (exam: ExamListItem) => void;
}

type TimerMode = "overall" | "per_question";

function scheduledAtInputValue(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function scheduledAtIso(value: string): string | null {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return null;
  return d.toISOString();
}

export function ExamModal({ open, onClose, exam, onSaved }: ExamModalProps) {
  const { t } = useTranslation();
  const isEdit = Boolean(exam);
  const create = useCreateExam();
  const update = useUpdateExam(exam?.id ?? "");

  const [title, setTitle] = useState("");
  const [scheduledAt, setScheduledAt] = useState("");
  const [timerMode, setTimerMode] = useState<TimerMode>("overall");
  const [duration, setDuration] = useState("");
  const [isFree, setIsFree] = useState(false);
  const [requiresCheckin, setRequiresCheckin] = useState(false);
  const [allowLeaderboard, setAllowLeaderboard] = useState(false);
  const [randomize, setRandomize] = useState(false);

  useEffect(() => {
    if (!open) return;
    if (exam) {
      setTitle(exam.title ?? "");
      setScheduledAt(scheduledAtInputValue(exam.scheduled_at));
      setTimerMode((exam.timer_mode as TimerMode) ?? "overall");
      setDuration(exam.duration_minutes != null ? String(exam.duration_minutes) : "");
      setIsFree(Boolean(exam.is_free));
      setRequiresCheckin(Boolean(exam.requires_checkin));
      setAllowLeaderboard(Boolean(exam.allow_leaderboard));
      setRandomize(Boolean(exam.randomize));
    } else {
      setTitle("");
      setScheduledAt("");
      setTimerMode("overall");
      setDuration("");
      setIsFree(false);
      setRequiresCheckin(false);
      setAllowLeaderboard(false);
      setRandomize(false);
    }
  }, [open, exam]);

  const isPending = create.isPending || update.isPending;
  const durationRequired = timerMode === "overall";
  const canSubmit = useMemo(
    () =>
      title.trim() !== "" &&
      (!durationRequired || (duration !== "" && Number(duration) > 0)) &&
      !isPending,
    [title, duration, durationRequired, isPending],
  );

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit || isPending) return;

    const base = {
      title: title.trim(),
      scheduled_at: scheduledAtIso(scheduledAt),
      timer_mode: timerMode,
      is_free: isFree,
      requires_checkin: requiresCheckin,
      allow_leaderboard: allowLeaderboard,
      randomize,
    };

    try {
      if (isEdit) {
        const payload: UpdateExamPayload = {
          ...base,
          duration_minutes:
            timerMode === "overall" && duration !== "" ? Number(duration) : null,
        };
        const saved = await update.mutateAsync(payload);
        onSaved?.(saved);
      } else {
        const payload: CreateExamPayload = {
          ...base,
          duration_minutes:
            timerMode === "overall" && duration !== "" ? Number(duration) : null,
        };
        const result = await create.mutateAsync(payload);
        if (result?.exam) onSaved?.(result.exam as ExamListItem);
      }
      onClose();
    } catch {
      // mutation hook handles errors via global state; close on success only
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { if (!next) onClose(); }}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>
              {isEdit
                ? t("exam_packages_modal_edit_title")
                : t("exam_packages_modal_create_title")}
            </DialogTitle>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="exam-title">{t("exam_packages_modal_title")}</Label>
              <Input
                id="exam-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t("exam_packages_modal_title_placeholder")}
                disabled={isPending}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="exam-scheduled-at">
                {t("exam_packages_modal_scheduled_at")}
              </Label>
              <Input
                id="exam-scheduled-at"
                type="datetime-local"
                value={scheduledAt}
                onChange={(e) => setScheduledAt(e.target.value)}
                disabled={isPending}
              />
            </div>

            <div className="grid gap-2">
              <Label>{t("exam_packages_modal_timer_mode")}</Label>
              <div className="grid grid-cols-2 gap-2">
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="timer_mode"
                    value="overall"
                    checked={timerMode === "overall"}
                    onChange={() => setTimerMode("overall")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_timer_overall")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="timer_mode"
                    value="per_question"
                    checked={timerMode === "per_question"}
                    onChange={() => setTimerMode("per_question")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_timer_per_question")}</span>
                </label>
              </div>
            </div>

            {timerMode === "overall" && (
              <div className="grid gap-2">
                <Label htmlFor="exam-duration-minutes">
                  {t("exam_packages_modal_duration_minutes")}
                </Label>
                <Input
                  id="exam-duration-minutes"
                  type="number"
                  min={1}
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                  placeholder="60"
                  disabled={isPending}
                />
              </div>
            )}

            <div className="grid gap-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={isFree}
                  onChange={(e) => setIsFree(e.target.checked)}
                  disabled={isPending}
                />
                <span>{t("exam_packages_modal_is_free")}</span>
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={requiresCheckin}
                  onChange={(e) => setRequiresCheckin(e.target.checked)}
                  disabled={isPending}
                />
                <span>{t("exam_packages_modal_requires_checkin")}</span>
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={allowLeaderboard}
                  onChange={(e) => setAllowLeaderboard(e.target.checked)}
                  disabled={isPending}
                />
                <span>{t("exam_packages_modal_allow_leaderboard")}</span>
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={randomize}
                  onChange={(e) => setRandomize(e.target.checked)}
                  disabled={isPending}
                />
                <span>{t("exam_packages_modal_randomize")}</span>
              </label>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isPending}
            >
              {t("exam_packages_modal_cancel")}
            </Button>
            <Button type="submit" disabled={!canSubmit || isPending}>
              {isPending ? t("saving") : t("exam_packages_modal_save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
