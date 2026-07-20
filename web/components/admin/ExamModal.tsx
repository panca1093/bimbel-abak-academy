"use client";

import { useEffect, useMemo, useRef, useState } from "react";
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
import { toast } from "sonner";
import { fetchCertificatePreview, useCreateExam, useUpdateExam } from "@/lib/hooks/admin-exams";
import { usePresignUpload } from "@/lib/hooks/students";
import type { ExamListItem, CreateExamPayload, UpdateExamPayload, ExamResultConfig } from "@/lib/types";

interface ExamModalProps {
  open: boolean;
  onClose: () => void;
  exam?: ExamListItem | null;
  onSaved?: (exam: ExamListItem) => void;
}

type TimerMode = "overall" | "per_test";
type CertificateTemplate = "classic" | "modern" | "elegant" | "custom";

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

function nonNegativeInt(value: string): number | null {
  const n = value === "" ? NaN : Number(value);
  if (!Number.isInteger(n)) return null;
  return Math.max(0, n);
}

function inputValueFromNumber(n: number | null | undefined): string {
  return n != null ? String(n) : "";
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
  const [certificateTemplate, setCertificateTemplate] = useState<CertificateTemplate>("classic");
  const [certificateBackgroundUrl, setCertificateBackgroundUrl] = useState("");
  const [backgroundUploading, setBackgroundUploading] = useState(false);
  const [mode, setMode] = useState("standard");
  const [resultConfig, setResultConfig] = useState<ExamResultConfig>("hidden");
  const [resultReleaseAt, setResultReleaseAt] = useState("");
  const [checkInWindow, setCheckInWindow] = useState("");
  const [graceWindow, setGraceWindow] = useState("");
  const [maxAttempts, setMaxAttempts] = useState("");
  const previewUrlRef = useRef<string | null>(null);
  const backgroundInputRef = useRef<HTMLInputElement>(null);
  const presign = usePresignUpload();

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
      setCertificateTemplate((exam.certificate_template as CertificateTemplate) ?? "classic");
      setCertificateBackgroundUrl(exam.certificate_background_url ?? "");
      setMode(exam.mode ?? "standard");
      setResultConfig((exam.result_config as ExamResultConfig) ?? "hidden");
      setResultReleaseAt(scheduledAtInputValue(exam.result_release_at));
      setCheckInWindow(inputValueFromNumber(exam.check_in_window_minutes));
      setGraceWindow(inputValueFromNumber(exam.grace_window_minutes));
      setMaxAttempts(inputValueFromNumber(exam.max_attempts));
    } else {
      setTitle("");
      setScheduledAt("");
      setTimerMode("overall");
      setDuration("");
      setIsFree(false);
      setRequiresCheckin(false);
      setAllowLeaderboard(false);
      setRandomize(false);
      setCertificateTemplate("classic");
      setCertificateBackgroundUrl("");
      setMode("standard");
      setResultConfig("hidden");
      setResultReleaseAt("");
      setCheckInWindow("");
      setGraceWindow("");
      setMaxAttempts("");
    }
  }, [open, exam]);

  const isPending = create.isPending || update.isPending;
  const durationRequired = timerMode === "overall";

  async function handleBackgroundSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setBackgroundUploading(true);
    try {
      const presigned = await presign.mutateAsync({ filename: file.name, content_type: file.type });
      const res = await fetch(presigned.url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      });
      if (!res.ok) throw new Error(`Upload failed: ${res.status}`);
      setCertificateBackgroundUrl(presigned.key);
    } catch {
      // upload failed; leave existing background untouched
    } finally {
      setBackgroundUploading(false);
      if (backgroundInputRef.current) backgroundInputRef.current.value = "";
    }
  }

  const canSubmit = useMemo(
    () =>
      title.trim() !== "" &&
      (!durationRequired || (duration !== "" && Number(duration) > 0)) &&
      (certificateTemplate !== "custom" || certificateBackgroundUrl.trim() !== "") &&
      !isPending,
    [title, duration, durationRequired, certificateTemplate, certificateBackgroundUrl, isPending],
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
      certificate_template: certificateTemplate,
      certificate_background_url: certificateBackgroundUrl || null,
      mode,
      result_config: resultConfig,
      result_release_at: scheduledAtIso(resultReleaseAt),
      check_in_window_minutes: nonNegativeInt(checkInWindow),
      grace_window_minutes: nonNegativeInt(graceWindow),
      max_attempts: nonNegativeInt(maxAttempts),
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
        onSaved?.(result as ExamListItem);
      }
      onClose();
    } catch {
      toast.error(t("error_generic"));
    }
  }

  async function handlePreview() {
    if (!exam || isPending) return;
    try {
      const blob = await fetchCertificatePreview(exam.id, certificateTemplate);
      const url = URL.createObjectURL(blob);
      if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current);
      previewUrlRef.current = url;
      window.open(url, "_blank");
      setTimeout(() => {
        if (previewUrlRef.current === url) {
          URL.revokeObjectURL(url);
          previewUrlRef.current = null;
        }
      }, 30000);
    } catch {
      toast.error(t("error_generic"));
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
              <Label>{t("exam_packages_modal_mode")}</Label>
              <div className="grid grid-cols-3 gap-2">
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="mode"
                    value="standard"
                    checked={mode === "standard"}
                    onChange={() => setMode("standard")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_mode_standard")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="mode"
                    value="utbk"
                    checked={mode === "utbk"}
                    onChange={() => setMode("utbk")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_mode_utbk")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="mode"
                    value="ielts"
                    checked={mode === "ielts"}
                    onChange={() => setMode("ielts")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_mode_ielts")}</span>
                </label>
              </div>
              {mode !== "standard" && (
                <p className="text-xs text-muted-foreground">
                  {t("exam_packages_modal_mode_hint")}
                </p>
              )}
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
                    value="per_test"
                    checked={timerMode === "per_test"}
                    onChange={() => setTimerMode("per_test")}
                    disabled={isPending}
                  />
                  <span>{t("exam_packages_modal_timer_per_test")}</span>
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

            <div className="grid gap-2">
              <Label>{t("exam_packages_modal_certificate_template")}</Label>
              <div className="grid grid-cols-4 gap-2">
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="certificate_template"
                    value="classic"
                    checked={certificateTemplate === "classic"}
                    onChange={() => setCertificateTemplate("classic")}
                    disabled={isPending}
                  />
                  <span>{t("certificate_template_classic")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="certificate_template"
                    value="modern"
                    checked={certificateTemplate === "modern"}
                    onChange={() => setCertificateTemplate("modern")}
                    disabled={isPending}
                  />
                  <span>{t("certificate_template_modern")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="certificate_template"
                    value="elegant"
                    checked={certificateTemplate === "elegant"}
                    onChange={() => setCertificateTemplate("elegant")}
                    disabled={isPending}
                  />
                  <span>{t("certificate_template_elegant")}</span>
                </label>
                <label className="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm">
                  <input
                    type="radio"
                    name="certificate_template"
                    value="custom"
                    checked={certificateTemplate === "custom"}
                    onChange={() => setCertificateTemplate("custom")}
                    disabled={isPending}
                  />
                  <span>{t("certificate_template_custom")}</span>
                </label>
              </div>
              {certificateTemplate === "custom" && (
                <div className="grid gap-2 rounded-md border border-input p-3">
                  <Label htmlFor="background-upload" className="text-xs font-medium">
                    {t("certificate_background_upload_label")}
                  </Label>
                  <input
                    ref={backgroundInputRef}
                    id="background-upload"
                    type="file"
                    accept="image/*"
                    onChange={handleBackgroundSelect}
                    disabled={isPending || backgroundUploading}
                    className="hidden"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => backgroundInputRef.current?.click()}
                    disabled={isPending || backgroundUploading}
                  >
                    {backgroundUploading ? t("certificate_background_uploading") : t("certificate_background_upload_button")}
                  </Button>
                  {certificateBackgroundUrl && (
                    <p className="text-xs text-muted-foreground">{certificateBackgroundUrl}</p>
                  )}
                </div>
              )}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handlePreview}
                disabled={!isEdit || isPending}
              >
                {t("admin_exam_certificate_preview")}
              </Button>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="exam-result-config">{t("exam_packages_modal_result_config")}</Label>
              <select
                id="exam-result-config"
                value={resultConfig}
                onChange={(e) => setResultConfig(e.target.value as ExamResultConfig)}
                disabled={isPending}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50 disabled:pointer-events-none disabled:opacity-50"
              >
                <option value="hidden">{t("exam_packages_modal_result_config_hidden")}</option>
                <option value="score_only">{t("exam_packages_modal_result_config_score_only")}</option>
                <option value="score_pembahasan">{t("exam_packages_modal_result_config_score_pembahasan")}</option>
              </select>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="exam-result-release-at">{t("exam_packages_modal_result_release_at")}</Label>
              <Input
                id="exam-result-release-at"
                type="datetime-local"
                value={resultReleaseAt}
                onChange={(e) => setResultReleaseAt(e.target.value)}
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="exam-check-in-window">{t("exam_packages_modal_check_in_window")}</Label>
                <Input
                  id="exam-check-in-window"
                  type="number"
                  min={0}
                  value={checkInWindow}
                  onChange={(e) => setCheckInWindow(e.target.value)}
                  disabled={isPending}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="exam-grace-window">{t("exam_packages_modal_grace_window")}</Label>
                <Input
                  id="exam-grace-window"
                  type="number"
                  min={0}
                  value={graceWindow}
                  onChange={(e) => setGraceWindow(e.target.value)}
                  disabled={isPending}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="exam-max-attempts">{t("exam_packages_modal_max_attempts")}</Label>
                <Input
                  id="exam-max-attempts"
                  type="number"
                  min={0}
                  value={maxAttempts}
                  onChange={(e) => setMaxAttempts(e.target.value)}
                  disabled={isPending}
                />
              </div>
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
