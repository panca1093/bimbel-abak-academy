"use client";

import { useEffect, useState } from "react";
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
import { Label } from "@/components/ui/label";
import { useTranslation } from "@/lib/i18n";
import type { Test, AdminCreateTestInput, AdminUpdateTestInput } from "@/lib/types";

interface TestModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  test?: Test | null;
  onSubmit: (input: AdminCreateTestInput | AdminUpdateTestInput) => void;
  isPending: boolean;
}

export function TestModal({ open, onOpenChange, test, onSubmit, isPending }: TestModalProps) {
  const { t } = useTranslation();
  const isEdit = Boolean(test);
  const [title, setTitle] = useState("");
  const [subject, setSubject] = useState("");
  const [topic, setTopic] = useState("");
  const [duration, setDuration] = useState("");
  const [audioUrl, setAudioUrl] = useState("");
  const [audioPlayLimit, setAudioPlayLimit] = useState("");

  useEffect(() => {
    if (!open) return;
    if (test) {
      setTitle(test.title ?? "");
      setSubject(test.subject ?? "");
      setTopic(test.topic ?? "");
      setDuration(test.duration_minutes != null ? String(test.duration_minutes) : "");
      setAudioUrl(test.audio_url ?? "");
      setAudioPlayLimit(test.audio_play_limit != null ? String(test.audio_play_limit) : "");
    } else {
      setTitle("");
      setSubject("");
      setTopic("");
      setDuration("");
      setAudioUrl("");
      setAudioPlayLimit("");
    }
  }, [open, test]);

  const canSubmit =
    title.trim() !== "" &&
    subject.trim() !== "" &&
    topic.trim() !== "" &&
    duration !== "" &&
    Number(duration) > 0 &&
    !isPending;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit || isPending) return;

    if (isEdit) {
      const input: AdminUpdateTestInput = {
        title: title.trim(),
        subject: subject.trim(),
        topic: topic.trim(),
        duration_minutes: Number(duration),
        ...(audioUrl.trim() ? { audio_url: audioUrl.trim() } : {}),
        ...(audioPlayLimit !== "" ? { audio_play_limit: Number(audioPlayLimit) } : {}),
      };
      onSubmit(input);
      return;
    }

    const input: AdminCreateTestInput = {
      title: title.trim(),
      subject: subject.trim(),
      topic: topic.trim(),
      duration_minutes: Number(duration),
      ...(audioUrl.trim() ? { audio_url: audioUrl.trim() } : {}),
      ...(audioPlayLimit !== "" ? { audio_play_limit: Number(audioPlayLimit) } : {}),
    };
    onSubmit(input);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{isEdit ? t("tests_edit") : t("tests_new")}</DialogTitle>
            <DialogDescription>
              {isEdit
                ? "Perbarui metadata tes."
                : "Tambahkan tes baru ke bank soal."}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="test-title">{t("tests_field_title")}</Label>
              <Input
                id="test-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t("tests_field_title")}
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="test-subject">{t("tests_field_subject")}</Label>
                <Input
                  id="test-subject"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  placeholder={t("tests_field_subject")}
                  disabled={isPending}
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="test-topic">{t("tests_field_topic")}</Label>
                <Input
                  id="test-topic"
                  value={topic}
                  onChange={(e) => setTopic(e.target.value)}
                  placeholder={t("tests_field_topic")}
                  disabled={isPending}
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="test-duration">{t("tests_field_duration")}</Label>
              <Input
                id="test-duration"
                type="number"
                min={1}
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                placeholder="60"
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="test-audio-url">{t("tests_field_audio_url")}</Label>
                <Input
                  id="test-audio-url"
                  value={audioUrl}
                  onChange={(e) => setAudioUrl(e.target.value)}
                  placeholder="https://..."
                  disabled={isPending}
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="test-audio-play-limit">{t("tests_field_audio_play_limit")}</Label>
                <Input
                  id="test-audio-play-limit"
                  type="number"
                  min={0}
                  value={audioPlayLimit}
                  onChange={(e) => setAudioPlayLimit(e.target.value)}
                  placeholder="0"
                  disabled={isPending}
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              {t("cancel")}
            </Button>
            <Button type="submit" disabled={!canSubmit || isPending}>
              {isPending ? t("saving") : t("save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}