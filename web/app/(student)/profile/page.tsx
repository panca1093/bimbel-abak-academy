"use client";

import { useEffect, useMemo, useState } from "react";
import { Key, Loader2 } from "lucide-react";
import {
  useChangePassword,
  useProfile,
  useUpdateProfile,
} from "@/lib/hooks/students";
import { useTranslation } from "@/lib/i18n";
import { useAuthStore } from "@/stores/auth";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

function initials(name?: string): string {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const first = parts[0]?.[0] ?? "";
  const second = parts[1]?.[0] ?? "";
  return (first + second).toUpperCase() || name.trim()[0].toUpperCase();
}

function formatJoined(iso?: string, lang: "id" | "en" = "id"): string {
  if (!iso) return "—";
  try {
    const d = new Date(iso);
    return d.toLocaleDateString(lang === "id" ? "id-ID" : "en-US", {
      month: "long",
      year: "numeric",
    });
  } catch {
    return iso;
  }
}

function Field({
  id,
  label,
  value,
  onChange,
  locked,
  hint,
  isLoading,
  type = "text",
}: {
  id: string;
  label: string;
  value?: string;
  onChange?: (v: string) => void;
  locked?: boolean;
  hint?: string;
  isLoading?: boolean;
  type?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      {isLoading ? (
        <Skeleton className="h-9 w-full rounded-md" />
      ) : (
        <Input
          id={id}
          type={type}
          value={value ?? ""}
          placeholder="—"
          disabled={locked || !onChange}
          readOnly={locked || !onChange}
          onChange={onChange ? (e) => onChange(e.target.value) : undefined}
        />
      )}
      {hint && <div className="text-xs text-ink-400">{hint}</div>}
    </div>
  );
}

function Switch({
  on,
  onClick,
  ariaLabel,
}: {
  on: boolean;
  onClick: () => void;
  ariaLabel?: string;
}) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      onClick={onClick}
      className={`relative h-6 w-11 shrink-0 rounded-full border-0 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 ${
        on ? "bg-brand-500" : "bg-line"
      }`}
    >
      <span
        className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all ${
          on ? "left-[22px]" : "left-0.5"
        }`}
      />
    </button>
  );
}

function ChangePasswordDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
}) {
  const { t } = useTranslation();
  const change = useChangePassword();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");

  useEffect(() => {
    if (!open) {
      setCurrent("");
      setNext("");
      setConfirm("");
    }
  }, [open]);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (next.length < 8) {
      toast.error(t("weak_password"));
      return;
    }
    if (next !== confirm) {
      toast.error(t("passwords_do_not_match"));
      return;
    }
    change.mutate(
      { current_password: current, new_password: next },
      {
        onSuccess: () => {
          toast.success(t("password_changed"));
          onOpenChange(false);
        },
        onError: (err) => {
          toast.error(err instanceof Error ? err.message : t("change_password_failed"));
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("change_password")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="flex flex-col gap-4 py-2">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="current-password">{t("current_password")}</Label>
            <Input
              id="current-password"
              type="password"
              value={current}
              onChange={(e) => setCurrent(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="new-password">{t("new_password")}</Label>
            <Input
              id="new-password"
              type="password"
              value={next}
              onChange={(e) => setNext(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="confirm-password">{t("confirm_password")}</Label>
            <Input
              id="confirm-password"
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t("cancel")}
            </Button>
            <Button type="submit" disabled={change.isPending}>
              {change.isPending ? (
                <Loader2 className="mr-2 size-4 animate-spin" />
              ) : null}
              {t("update")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function ProfilePage() {
  const { t, lang } = useTranslation();
  const { data: profile, isLoading, isError, error, refetch } = useProfile();
  const updateProfile = useUpdateProfile();
  const token = useAuthStore((s) => s.token);
  const setSession = useAuthStore((s) => s.setSession);

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [targetExam, setTargetExam] = useState("");
  const [emailNotif, setEmailNotif] = useState(true);
  const [waNotif, setWaNotif] = useState(true);
  const [pushNotif, setPushNotif] = useState(false);
  const [passwordOpen, setPasswordOpen] = useState(false);

  useEffect(() => {
    if (profile) {
      setName(profile.name ?? "");
      setPhone(profile.phone ?? "");
      setAddress(profile.alamat_domisili ?? "");
      setTargetExam(profile.target_exam ?? "");
    }
  }, [profile]);

  const displayName = profile?.name ?? "";
  const metaLine = useMemo(() => {
    if (!profile) return "";
    const joined = formatJoined(profile.created_at, lang);
    return profile.username
      ? `@${profile.username} · ${t("joined")} ${joined}`
      : `${t("joined")} ${joined}`;
  }, [profile, lang, t]);

  function handleSave(e: React.FormEvent) {
    e.preventDefault();
    updateProfile.mutate(
      {
        name: name || undefined,
        phone: phone || undefined,
        alamat_domisili: address || undefined,
        target_exam: targetExam || undefined,
      },
      {
        onSuccess: (updated) => {
          toast.success(t("saved"));
          if (updated && token) {
            setSession(token, updated);
          }
        },
        onError: (err) => {
          toast.error(err instanceof Error ? err.message : t("save_failed"));
        },
      }
    );
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10">
        <header className="mb-6">
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("profile_title")}
          </h1>
        </header>
        <Card className="border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <div className="flex-1 text-sm text-ink-700">
              {t("load_failed")}
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

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-6">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
          {t("profile_title")}
        </h1>
      </header>

      <div className="grid grid-cols-1 items-start gap-5 lg:grid-cols-[1fr_320px]">
        <Card className="p-6">
          <div className="mb-6 flex items-center gap-4">
            {isLoading ? (
              <Skeleton className="size-16 rounded-full" />
            ) : (
              <Avatar className="size-16 rounded-full bg-brand-100 text-brand-700">
                <AvatarFallback className="size-16 rounded-full bg-brand-100 text-lg font-semibold text-brand-700">
                  {initials(displayName)}
                </AvatarFallback>
              </Avatar>
            )}
            <div className="min-w-0">
              {isLoading ? (
                <>
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="mt-2 h-4 w-56" />
                </>
              ) : (
                <>
                  <div className="font-serif text-xl font-semibold text-ink-900">
                    {displayName || t("unnamed")}
                  </div>
                  <div className="truncate text-sm text-ink-500">{metaLine}</div>
                </>
              )}
            </div>
          </div>

          <form onSubmit={handleSave} className="flex flex-col gap-4" noValidate>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field
                id="name"
                label={t("full_name")}
                value={name}
                onChange={setName}
                isLoading={isLoading}
              />
              <Field
                id="phone"
                label={t("phone")}
                value={phone}
                onChange={setPhone}
                isLoading={isLoading}
              />
              <Field
                id="email"
                label={t("email")}
                value={profile?.email}
                locked
                hint={t("email_locked")}
                isLoading={isLoading}
              />
              <Field
                id="school"
                label={t("school")}
                value="—"
                locked
                isLoading={isLoading}
              />
              <Field
                id="grade"
                label={t("grade")}
                value={profile?.grade}
                locked
                isLoading={isLoading}
              />
              <Field
                id="nis"
                label={t("nis")}
                value={profile?.nis}
                locked
                isLoading={isLoading}
              />
              <Field
                id="address"
                label={t("address")}
                value={address}
                onChange={setAddress}
                isLoading={isLoading}
              />
              <Field
                id="target_exam"
                label={t("target_exam")}
                value={targetExam}
                onChange={setTargetExam}
                isLoading={isLoading}
              />
            </div>

            <div className="flex flex-wrap gap-3 pt-3">
              <Button
                type="submit"
                disabled={updateProfile.isPending || isLoading}
              >
                {updateProfile.isPending ? (
                  <Loader2 className="mr-2 size-4 animate-spin" />
                ) : null}
                {updateProfile.isPending ? t("saving") : t("save_changes")}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => setPasswordOpen(true)}
                disabled={isLoading}
              >
                <Key className="mr-2 size-4" />
                {t("change_password")}
              </Button>
            </div>
          </form>
        </Card>

        <Card className="p-5">
          <h3 className="mb-4 text-[15px] font-semibold text-ink-900">
            {t("notif_prefs")}
          </h3>
          <div className="flex flex-col gap-3">
            <label className="flex items-center justify-between gap-3">
              <div>
                <div className="text-sm text-ink-900">{t("email_notif")}</div>
                <div className="text-xs text-ink-400">
                  {profile?.email ?? "—"}
                </div>
              </div>
              <Switch on={emailNotif} onClick={() => setEmailNotif((v) => !v)} />
            </label>
            <div className="h-px bg-line" />
            <label className="flex items-center justify-between gap-3">
              <div>
                <div className="text-sm text-ink-900">{t("wa_notif")}</div>
                <div className="text-xs text-ink-400">
                  {profile?.phone ?? "—"}
                </div>
              </div>
              <Switch on={waNotif} onClick={() => setWaNotif((v) => !v)} />
            </label>
            <div className="h-px bg-line" />
            <label className="flex items-center justify-between gap-3">
              <span className="text-sm text-ink-900">{t("push_notif")}</span>
              <Switch on={pushNotif} onClick={() => setPushNotif((v) => !v)} />
            </label>
          </div>
        </Card>
      </div>

      <ChangePasswordDialog open={passwordOpen} onOpenChange={setPasswordOpen} />
    </div>
  );
}
