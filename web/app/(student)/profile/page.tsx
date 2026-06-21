"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Camera, Key, Loader2, Pencil } from "lucide-react";
import {
  useChangePassword,
  usePresignUpload,
  useProfile,
  useSchools,
  useUpdatePhoto,
  useUpdateProfile,
} from "@/lib/hooks/students";
import { useTranslation } from "@/lib/i18n";
import { useAuthStore } from "@/stores/auth";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

const GRADES = ["7", "8", "9", "10", "11", "12"];

const PROFILE_INPUT_CLASS =
  "h-11 w-full rounded-md border border-line bg-surface px-3.5 text-sm text-ink-900 shadow-none transition-[border-color,box-shadow] outline-none placeholder:text-ink-400 focus-visible:border-brand-400 focus-visible:ring-[3px] focus-visible:ring-brand-50 disabled:cursor-not-allowed disabled:bg-surface-2/60 disabled:text-ink-500 read-only:bg-surface-2/60";

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
      <Label htmlFor={id} className="text-xs font-semibold text-ink-600">
        {label}
      </Label>
      {isLoading ? (
        <Skeleton className="h-11 w-full rounded-md" />
      ) : (
        <Input
          id={id}
          type={type}
          value={value ?? ""}
          placeholder="—"
          disabled={locked || !onChange}
          readOnly={locked || !onChange}
          onChange={onChange ? (e) => onChange(e.target.value) : undefined}
          className={PROFILE_INPUT_CLASS}
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
      className={`relative h-6 w-11 shrink-0 rounded-full border-0 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-300 ${
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
          toast.error(
            err instanceof Error ? err.message : t("change_password_failed")
          );
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
  const { data: schools, isLoading: schoolsLoading } = useSchools();
  const updateProfile = useUpdateProfile();
  const presign = usePresignUpload();
  const updatePhoto = useUpdatePhoto();
  const token = useAuthStore((s) => s.token);
  const setSession = useAuthStore((s) => s.setSession);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [editMode, setEditMode] = useState(false);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [targetExam, setTargetExam] = useState("");
  const [grade, setGrade] = useState<string>("");
  const [schoolId, setSchoolId] = useState<string>("");
  const [emailNotif, setEmailNotif] = useState(true);
  const [waNotif, setWaNotif] = useState(true);
  const [pushNotif, setPushNotif] = useState(false);
  const [passwordOpen, setPasswordOpen] = useState(false);
  const [photoUploading, setPhotoUploading] = useState(false);

  const syncedVersion = useRef<string | null>(null);
  useEffect(() => {
    if (!profile || editMode) return;
    const version = `${profile.id}:${profile.updated_at}`;
    if (syncedVersion.current === version) return;
    syncedVersion.current = version;
    setName(profile.name ?? "");
    setPhone(profile.phone ?? "");
    setAddress(profile.alamat_domisili ?? "");
    setTargetExam(profile.target_exam ?? "");
    setGrade(profile.grade != null ? String(profile.grade) : "");
    setSchoolId(profile.school_id ?? "");
    setWaNotif(!!profile.phone);
  }, [profile, editMode]);

  const displayName = profile?.name ?? "";
  const metaLine = useMemo(() => {
    if (!profile) return "";
    const joined = formatJoined(profile.created_at, lang);
    return profile.username
      ? `@${profile.username} · ${t("joined")} ${joined}`
      : `${t("joined")} ${joined}`;
  }, [profile, lang, t]);

  function cancelEdit() {
    if (profile) {
      setName(profile.name ?? "");
      setPhone(profile.phone ?? "");
      setAddress(profile.alamat_domisili ?? "");
      setTargetExam(profile.target_exam ?? "");
      setGrade(profile.grade != null ? String(profile.grade) : "");
      setSchoolId(profile.school_id ?? "");
    }
    setEditMode(false);
  }

  function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!editMode) return;
    const gradeNum = grade ? parseInt(grade, 10) : undefined;
    if (gradeNum !== undefined && Number.isNaN(gradeNum)) {
      toast.error("Kelas tidak valid.");
      return;
    }
    updateProfile.mutate(
      {
        name: name || undefined,
        phone: phone || undefined,
        address: address || undefined,
        target_exam: targetExam || undefined,
        grade: gradeNum,
        school_id: schoolId || undefined,
      },
      {
        onSuccess: (updated) => {
          toast.success(t("saved"));
          setEditMode(false);
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

  async function handlePhotoSelect(e: React.ChangeEvent<HTMLInputElement>) {
    if (!editMode) return;
    const file = e.target.files?.[0];
    if (!file) return;
    setPhotoUploading(true);
    try {
      const presigned = await presign.mutateAsync({
        filename: file.name,
        content_type: file.type,
      });
      const uploadRes = await fetch(presigned.url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      });
      if (!uploadRes.ok) {
        throw new Error(`Upload failed: ${uploadRes.status}`);
      }
      await updatePhoto.mutateAsync(presigned.public_url);
      toast.success(t("photo_uploaded"));
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t("photo_upload_failed")
      );
    } finally {
      setPhotoUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  }

  if (isError) {
    return (
      <div className="fade-in">
        <header className="mb-6">
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("profile_title")}
          </h1>
        </header>
        <Card className="rounded-2xl border-danger/30 bg-danger-bg px-5 py-4 shadow-md">
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
    <div className="fade-in">
      <header className="mb-6">
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("profile_title")}
          </h1>
        </header>

        <div className="grid grid-cols-1 items-start gap-5 lg:grid-cols-[1fr_340px]">
          <Card className="rounded-2xl border-0 p-6 shadow-md transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
            <div className="mb-6 flex items-start justify-between gap-4">
              <div className="flex items-center gap-4">
                <div className="relative">
                  {isLoading ? (
                    <Skeleton className="size-20 rounded-full" />
                  ) : editMode ? (
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={photoUploading}
                      aria-label={t("upload_photo")}
                      className="relative flex size-20 cursor-pointer items-center justify-center rounded-full border-0 p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-300 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <Avatar
                        size="lg"
                        className={`size-20 rounded-full bg-gradient-to-br from-brand-100 to-brand-200 text-brand-700 ring-4 ring-surface shadow-sm ${
                          photoUploading ? "animate-pulse" : ""
                        }`}
                      >
                        {profile?.photo_url ? (
                          <AvatarImage
                            src={profile.photo_url}
                            alt={displayName}
                            className="object-cover"
                          />
                        ) : null}
                        <AvatarFallback className="rounded-full bg-transparent text-2xl font-semibold">
                          {initials(displayName)}
                        </AvatarFallback>
                      </Avatar>
                      <span className="absolute -right-1 -bottom-1 flex size-8 items-center justify-center rounded-full bg-brand-600 text-white shadow-md ring-2 ring-surface transition-transform hover:scale-110">
                        {photoUploading ? (
                          <Loader2 className="size-4 animate-spin" />
                        ) : (
                          <Camera className="size-4" />
                        )}
                      </span>
                    </button>
                  ) : (
                    <Avatar
                      size="lg"
                      className="size-20 rounded-full bg-gradient-to-br from-brand-100 to-brand-200 text-brand-700 ring-4 ring-surface shadow-sm"
                    >
                      {profile?.photo_url ? (
                        <AvatarImage
                          src={profile.photo_url}
                          alt={displayName}
                          className="object-cover"
                        />
                      ) : null}
                      <AvatarFallback className="rounded-full bg-transparent text-2xl font-semibold">
                        {initials(displayName)}
                      </AvatarFallback>
                    </Avatar>
                  )}
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handlePhotoSelect}
                  />
                </div>
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
                      {editMode && (
                        <button
                          type="button"
                          onClick={() => fileInputRef.current?.click()}
                          disabled={photoUploading}
                          className="mt-2 text-xs font-semibold text-brand-600 hover:text-brand-700 disabled:opacity-60"
                        >
                          {photoUploading ? t("saving") : t("upload_photo")}
                        </button>
                      )}
                    </>
                  )}
                </div>
              </div>
              {!isLoading && !editMode && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setEditMode(true)}
                >
                  <Pencil className="mr-2 size-4" />
                  {t("edit_profile")}
                </Button>
              )}
            </div>

            <form
              onSubmit={handleSave}
              className="flex flex-col gap-4"
              noValidate
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field
                  id="name"
                  label={t("full_name")}
                  value={name}
                  onChange={editMode ? setName : undefined}
                  isLoading={isLoading}
                />
                <Field
                  id="phone"
                  label={t("phone")}
                  value={phone}
                  onChange={editMode ? setPhone : undefined}
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
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="school" className="text-xs font-semibold text-ink-600">
                    {t("school")}
                  </Label>
                  {isLoading ? (
                    <Skeleton className="h-11 w-full rounded-md" />
                  ) : (
                    <Select
                      value={schoolId || "_empty_"}
                      onValueChange={(v) =>
                        setSchoolId(v === "_empty_" ? "" : v)
                      }
                      disabled={!editMode || schoolsLoading}
                    >
                      <SelectTrigger id="school" className={PROFILE_INPUT_CLASS}>
                        <SelectValue placeholder={t("select_school")} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_empty_">
                          {t("select_school")}
                        </SelectItem>
                        {schools?.map((s) => (
                          <SelectItem key={s.id} value={s.id}>
                            {s.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="grade" className="text-xs font-semibold text-ink-600">
                    {t("grade")}
                  </Label>
                  {isLoading ? (
                    <Skeleton className="h-11 w-full rounded-md" />
                  ) : (
                    <Select
                      value={grade || "_empty_"}
                      onValueChange={(v) => setGrade(v === "_empty_" ? "" : v)}
                      disabled={!editMode}
                    >
                      <SelectTrigger id="grade" className={PROFILE_INPUT_CLASS}>
                        <SelectValue placeholder={t("select_grade")} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_empty_">
                          {t("select_grade")}
                        </SelectItem>
                        {GRADES.map((g) => (
                          <SelectItem key={g} value={g}>
                            {g}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </div>
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
                  onChange={editMode ? setAddress : undefined}
                  isLoading={isLoading}
                />
                <Field
                  id="target_exam"
                  label={t("target_exam")}
                  value={targetExam}
                  onChange={editMode ? setTargetExam : undefined}
                  isLoading={isLoading}
                />
              </div>

              {editMode && (
                <div className="flex flex-wrap gap-3 pt-3">
                  <Button type="submit" disabled={updateProfile.isPending}>
                    {updateProfile.isPending ? (
                      <Loader2 className="mr-2 size-4 animate-spin" />
                    ) : null}
                    {updateProfile.isPending ? t("saving") : t("save_changes")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setPasswordOpen(true)}
                  >
                    <Key className="mr-2 size-4" />
                    {t("change_password")}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={cancelEdit}
                    disabled={updateProfile.isPending}
                  >
                    {t("cancel")}
                  </Button>
                </div>
              )}
            </form>
          </Card>

          <Card className="rounded-2xl border-0 p-5 shadow-md transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
            <h3 className="mb-4 text-[15px] font-semibold text-ink-900">
              {t("notif_prefs")}
            </h3>
            <div className="flex flex-col">
              <label className="flex items-center justify-between gap-3 py-3">
                <div>
                  <div className="text-sm text-ink-900">{t("email_notif")}</div>
                  <div className="text-xs text-ink-400">
                    {profile?.email ?? "—"}
                  </div>
                </div>
                <Switch
                  on={emailNotif}
                  onClick={() => setEmailNotif((v) => !v)}
                />
              </label>
              <label className="flex items-center justify-between gap-3 py-3">
                <div>
                  <div className="text-sm text-ink-900">{t("wa_notif")}</div>
                  <div className="text-xs text-ink-400">
                    {profile?.phone ?? "—"}
                  </div>
                </div>
                <Switch on={waNotif} onClick={() => setWaNotif((v) => !v)} />
              </label>
              <label className="flex items-center justify-between gap-3 py-3">
                <span className="text-sm text-ink-900">{t("push_notif")}</span>
                <Switch
                  on={pushNotif}
                  onClick={() => setPushNotif((v) => !v)}
                />
              </label>
            </div>
          </Card>
        </div>

      <ChangePasswordDialog
        open={passwordOpen}
        onOpenChange={setPasswordOpen}
      />
    </div>
  );
}
