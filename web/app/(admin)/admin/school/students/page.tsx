"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import {
  Plus,
  Users,
  MoreHorizontal,
  Lock,
  Search,
  Copy,
  Check,
} from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";
import {
  useAdminStudents,
  useRegisterStudent,
  useChangeStudentStatus,
  useReissueStudentCredentials,
} from "@/lib/hooks/admin-students";
import { useAdminSchools } from "@/lib/hooks/admin-schools";
import { useAuthStore } from "@/stores/auth";
import type {
  AdminStudent,
  StudentRegistrationInput,
  StudentRegistrationResult,
  StudentCredentials,
} from "@/lib/types";

const STATUS_TONE: Record<string, string> = {
  active: "bg-success-bg text-success border-success",
  deactivated: "bg-danger-bg text-danger border-danger",
};

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

export default function SchoolStudentsPage() {
  const { t, lang } = useTranslation();
  const dateLocale = lang === "en" ? "en-US" : "id-ID";

  // Role-gated school picker (super_admin only)
  const currentRole = useAuthStore((s) => s.user?.role);
  const isSuperAdmin = currentRole === "super_admin";
  const { data: schoolsData, isLoading: schoolsLoading } = useAdminSchools();
  const [selectedSchoolId, setSelectedSchoolId] = useState<string>("");

  // Filters
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  // Cursor pagination
  const [accumulated, setAccumulated] = useState<AdminStudent[]>([]);
  const [activeCursor, setActiveCursor] = useState<string | undefined>(undefined);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);

  // Guard: reset pagination on filter change
  const filterKey = `${statusFilter}:${search}:${selectedSchoolId}`;
  const pageFilterKeyRef = useRef(filterKey);

  useEffect(() => {
    if (filterKey !== pageFilterKeyRef.current) {
      setAccumulated([]);
      setActiveCursor(undefined);
      setNextCursor(undefined);
      pageFilterKeyRef.current = filterKey;
    }
  }, [filterKey]);

  const query = useAdminStudents({
    status: statusFilter === "all" ? undefined : statusFilter,
    q: search || undefined,
    cursor: activeCursor,
    limit: 20,
    ...(isSuperAdmin && selectedSchoolId ? { schoolId: selectedSchoolId } : {}),
    enabled: !isSuperAdmin || Boolean(selectedSchoolId),
  });

  // Accumulate pages as they arrive
  useEffect(() => {
    if (!query.data) return;
    if (filterKey !== pageFilterKeyRef.current) return;

    setAccumulated((prev) => {
      if (activeCursor === undefined) return query.data!.data;
      const ids = new Set(prev.map((s) => s.id));
      const fresh = query.data!.data.filter((s) => !ids.has(s.id));
      return [...prev, ...fresh];
    });
    setNextCursor(query.data.next_cursor);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.data]);

  // Register dialog
  const [registerOpen, setRegisterOpen] = useState(false);
  const [registerForm, setRegisterForm] = useState<StudentRegistrationInput>({
    name: "",
    nis: "",
    email: "",
    dob: "",
    gender: "",
    grade: undefined,
    alamat_domisili: "",
    target_exam: "",
  });
  const [registerResult, setRegisterResult] =
    useState<StudentRegistrationResult | null>(null);

  // Reissue dialog
  const [reissueTarget, setReissueTarget] = useState<AdminStudent | null>(null);
  const [reissueResult, setReissueResult] =
    useState<StudentCredentials | null>(null);

  // Copy-to-clipboard state
  const [copied, setCopied] = useState<"username" | "password" | null>(null);

  const registerStudent = useRegisterStudent();
  const changeStatus = useChangeStudentStatus();
  const reissueCreds = useReissueStudentCredentials();

  const stats = useMemo(
    () => ({
      total: accumulated.length,
      active: accumulated.filter((s) => s.status === "active").length,
      deactivated: accumulated.filter((s) => s.status === "deactivated").length,
    }),
    [accumulated]
  );

  const handleRegister = async () => {
    if (!registerForm.name || !registerForm.nis) {
      toast.error(t("accounts_toast_required"));
      return;
    }
    try {
      const result = await registerStudent.mutateAsync({
        input: registerForm,
        schoolId: isSuperAdmin ? selectedSchoolId : undefined,
      });
      setRegisterResult(result);
      toast.success(t("students_register_success"));
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : t("students_register_failed");
      toast.error(msg);
    }
  };

  const handleStatusToggle = async (student: AdminStudent) => {
    const newStatus =
      student.status === "active" ? "deactivated" : "active";
    const success =
      newStatus === "active"
        ? t("students_toast_activated")
        : t("students_toast_deactivated");
    try {
      await changeStatus.mutateAsync({
        id: student.id,
        status: newStatus,
        schoolId: isSuperAdmin ? selectedSchoolId : undefined,
      });
      toast.success(success);
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : t("students_toast_status_failed");
      toast.error(msg);
    }
  };

  const handleReissue = async () => {
    if (!reissueTarget) return;
    try {
      const result = await reissueCreds.mutateAsync({
        id: reissueTarget.id,
        schoolId: isSuperAdmin ? selectedSchoolId : undefined,
      });
      setReissueResult(result);
      toast.success(t("students_credential_reissued"));
    } catch (err: unknown) {
      const msg =
        err instanceof Error
          ? err.message
          : t("students_credential_reissue_failed");
      toast.error(msg);
    }
  };

  const handleCopy = async (text: string, field: "username" | "password") => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(field);
      setTimeout(() => setCopied(null), 2000);
    } catch {
      // Clipboard unavailable — silently ignore
    }
  };

  const handleCloseRegister = () => {
    setRegisterOpen(false);
    // Discard plaintext credentials
    setRegisterResult(null);
    setRegisterForm({
      name: "",
      nis: "",
      email: "",
      dob: "",
      gender: "",
      grade: undefined,
      alamat_domisili: "",
      target_exam: "",
    });
  };

  const handleCloseReissue = () => {
    setReissueTarget(null);
    setReissueResult(null);
  };

  const handleLoadMore = () => {
    if (nextCursor) {
      setActiveCursor(nextCursor);
    }
  };

  // Loading state (first page only)
  if (query.isLoading && accumulated.length === 0) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Users}
          title={t("school_students_title")}
          description={t("sys_loading")}
        />
        <div className="py-12 text-center text-ink-500">
          {t("sys_loading_data")}
        </div>
      </div>
    );
  }

  if (query.error && accumulated.length === 0) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Users}
          title={t("school_students_title")}
          description={t("sys_error_title")}
        />
        <div className="py-12 text-center text-ink-500">
          {t("sys_error_load")}
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={Users}
        title={t("school_students_title")}
        description={t("students_subtitle")}
        actions={
          <Button size="sm" onClick={() => setRegisterOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("students_register_title")}
          </Button>
        }
      />

      {/* School picker (super_admin only) */}
      {isSuperAdmin && (
        <div className="mb-6">
          <p className="text-xs text-ink-500">{t("select_school")}</p>
          {schoolsLoading ? (
            <div className="mt-1 h-9 w-[240px] animate-pulse rounded-md bg-surface-2" />
          ) : (
            <Select value={selectedSchoolId} onValueChange={setSelectedSchoolId}>
              <SelectTrigger className="mt-1 h-9 w-[240px] text-xs" aria-label={t("select_school")}>
                <SelectValue placeholder={t("select_school")} />
              </SelectTrigger>
              <SelectContent>
                {(schoolsData?.data ?? []).map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
      )}

      {/* Stats */}
      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label={t("accounts_stat_total")}
          value={String(stats.total)}
        />
        <StatCard
          label={t("status_label_active")}
          value={String(stats.active)}
        />
        <StatCard
          label={t("status_label_inactive")}
          value={String(stats.deactivated)}
        />
      </div>

      {/* Filters */}
      <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip
            active={statusFilter === "all"}
            onClick={() => setStatusFilter("all")}
          >
            {t("tab_all")}
          </FilterChip>
          <FilterChip
            active={statusFilter === "active"}
            onClick={() => setStatusFilter("active")}
          >
            {t("status_label_active")}
          </FilterChip>
          <FilterChip
            active={statusFilter === "deactivated"}
            onClick={() => setStatusFilter("deactivated")}
          >
            {t("status_label_inactive")}
          </FilterChip>
        </div>
        <div className="flex items-center gap-2 lg:ml-auto">
          <Select
            value={statusFilter}
            onValueChange={(v) => setStatusFilter(v)}
          >
            <SelectTrigger className="h-9 w-[140px] text-xs">
              <SelectValue placeholder={t("accounts_status_placeholder")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("accounts_status_all")}</SelectItem>
              <SelectItem value="active">
                {t("status_label_active")}
              </SelectItem>
              <SelectItem value="deactivated">
                {t("status_label_inactive")}
              </SelectItem>
            </SelectContent>
          </Select>
          <Search className="size-4 text-ink-400" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("students_search_placeholder")}
            className="h-9 w-[200px] text-xs"
          />
        </div>
      </div>

      {/* Table */}
      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">{t("students_field_name")}</th>
                <th className="px-4 py-3">{t("th_status")}</th>
                <th className="px-4 py-3">{t("students_field_grade")}</th>
                <th className="px-4 py-3">{t("accounts_th_created")}</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {accumulated.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    className="px-4 py-8 text-center text-sm text-ink-500"
                  >
                    {t("students_empty")}
                  </td>
                </tr>
              )}
              {accumulated.map((s) => (
                <tr key={s.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Avatar size="sm">
                        <AvatarFallback className="bg-brand-50 text-brand-700 text-xs">
                          {initials(s.name)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-ink-900">
                          {s.name}
                        </div>
                        <div className="text-[11px] text-ink-500">
                          NIS: {s.nis} · @{s.username}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold capitalize",
                        STATUS_TONE[s.status] ??
                          "bg-surface-2 text-ink-500 border-line"
                      )}
                    >
                      {s.status === "active"
                        ? t("status_label_active")
                        : t("status_label_inactive")}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {s.grade || "—"}
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {s.created_at
                      ? new Date(s.created_at).toLocaleString(dateLocale, {
                          day: "2-digit",
                          month: "short",
                          year: "numeric",
                        })
                      : "—"}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon-xs">
                          <MoreHorizontal className="size-4 text-ink-500" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => setReissueTarget(s)}
                        >
                          <Lock className="mr-2 size-4" />
                          {t("students_credential_reissue")}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleStatusToggle(s)}
                        >
                          <Lock className="mr-2 size-4" />
                          {s.status === "active"
                            ? t("students_status_toggle_deactivated")
                            : t("students_status_toggle_active")}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Load more */}
        {nextCursor && (
          <div className="border-t border-line px-4 py-3 text-center">
            <Button
              variant="outline"
              size="sm"
              onClick={handleLoadMore}
              disabled={query.isFetching}
            >
              {query.isFetching ? t("sys_loading") : t("load_more")}
            </Button>
          </div>
        )}
      </div>

      {/* Register dialog */}
      <Dialog open={registerOpen} onOpenChange={(open) => {
        if (!open) handleCloseRegister();
      }}>
        <DialogContent className="sm:max-w-lg">
          {registerResult ? (
            /* Credential panel — one-time display */
            <>
              <DialogHeader>
                <DialogTitle className="font-serif">
                  {t("students_credential_display")}
                </DialogTitle>
                <DialogDescription>
                  {t("students_credential_warning")}
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label>{t("students_credential_username")}</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="flex-1 rounded-md border border-line bg-surface-2 px-3 py-2 text-sm font-mono">
                      {registerResult.username}
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handleCopy(registerResult.username, "username")
                      }
                    >
                      {copied === "username" ? (
                        <Check className="size-4 text-success" />
                      ) : (
                        <Copy className="size-4" />
                      )}
                      <span className="ml-1">
                        {copied === "username"
                          ? t("students_credential_copied")
                          : t("students_credential_copy")}
                      </span>
                    </Button>
                  </div>
                </div>
                <div>
                  <Label>{t("students_credential_password")}</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="flex-1 rounded-md border border-line bg-surface-2 px-3 py-2 text-sm font-mono">
                      {registerResult.temp_password}
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handleCopy(registerResult.temp_password, "password")
                      }
                    >
                      {copied === "password" ? (
                        <Check className="size-4 text-success" />
                      ) : (
                        <Copy className="size-4" />
                      )}
                      <span className="ml-1">
                        {copied === "password"
                          ? t("students_credential_copied")
                          : t("students_credential_copy")}
                      </span>
                    </Button>
                  </div>
                </div>
              </div>
              <DialogFooter className="mt-4">
                <Button onClick={handleCloseRegister}>
                  {t("cancel")}
                </Button>
              </DialogFooter>
            </>
          ) : (
            /* Registration form */
            <>
              <DialogHeader>
                <DialogTitle className="font-serif">
                  {t("students_register_title")}
                </DialogTitle>
                <DialogDescription>
                  {t("students_register_desc")}
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label>
                    {t("students_field_name")}{" "}
                    <span className="text-danger">*</span>
                  </Label>
                  <Input
                    value={registerForm.name}
                    onChange={(e) =>
                      setRegisterForm((f) => ({
                        ...f,
                        name: e.target.value,
                      }))
                    }
                    placeholder={t("students_field_name")}
                  />
                </div>
                <div>
                  <Label>
                    {t("students_field_nis")}{" "}
                    <span className="text-danger">*</span>
                  </Label>
                  <Input
                    value={registerForm.nis}
                    onChange={(e) =>
                      setRegisterForm((f) => ({
                        ...f,
                        nis: e.target.value,
                      }))
                    }
                    placeholder={t("students_field_nis")}
                  />
                </div>
                <div>
                  <Label>{t("students_field_email")}</Label>
                  <Input
                    type="email"
                    value={registerForm.email ?? ""}
                    onChange={(e) =>
                      setRegisterForm((f) => ({
                        ...f,
                        email: e.target.value || undefined,
                      }))
                    }
                    placeholder={t("accounts_placeholder_email")}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>{t("students_field_dob")}</Label>
                    <Input
                      type="date"
                      value={registerForm.dob ?? ""}
                      onChange={(e) =>
                        setRegisterForm((f) => ({
                          ...f,
                          dob: e.target.value || undefined,
                        }))
                      }
                    />
                  </div>
                  <div>
                    <Label>{t("students_field_gender")}</Label>
                    <Select
                      value={registerForm.gender ?? ""}
                      onValueChange={(v) =>
                        setRegisterForm((f) => ({
                          ...f,
                          gender: v || undefined,
                        }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue
                          placeholder={t("accounts_placeholder_pick_role")}
                        />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="male">
                          {lang === "id" ? "Laki-laki" : "Male"}
                        </SelectItem>
                        <SelectItem value="female">
                          {lang === "id" ? "Perempuan" : "Female"}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>{t("students_field_grade")}</Label>
                    <Input
                      value={registerForm.grade ?? ""}
                      onChange={(e) =>
                        setRegisterForm((f) => ({
                          ...f,
                          grade: e.target.value ? Number(e.target.value) : undefined,
                        }))
                      }
                      placeholder={t("students_field_grade")}
                    />
                  </div>
                  <div>
                    <Label>{t("students_field_target_exam")}</Label>
                    <Input
                      value={registerForm.target_exam ?? ""}
                      onChange={(e) =>
                        setRegisterForm((f) => ({
                          ...f,
                          target_exam: e.target.value || undefined,
                        }))
                      }
                      placeholder={t("students_field_target_exam")}
                    />
                  </div>
                </div>
                <div>
                  <Label>{t("students_field_alamat_domisili")}</Label>
                  <Input
                    value={registerForm.alamat_domisili ?? ""}
                    onChange={(e) =>
                      setRegisterForm((f) => ({
                        ...f,
                        alamat_domisili: e.target.value || undefined,
                      }))
                    }
                    placeholder={t("students_field_alamat_domisili")}
                  />
                </div>
              </div>
              <DialogFooter className="mt-4">
                <Button variant="outline" onClick={handleCloseRegister}>
                  {t("cancel")}
                </Button>
                <Button
                  onClick={handleRegister}
                  disabled={registerStudent.isPending}
                >
                  {registerStudent.isPending
                    ? t("saving")
                    : t("students_register_title")}
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>

      {/* Reissue credentials dialog */}
      <Dialog
        open={reissueTarget !== null}
        onOpenChange={(open) => {
          if (!open) handleCloseReissue();
        }}
      >
        <DialogContent className="sm:max-w-lg">
          {reissueResult ? (
            /* Credential panel — one-time display */
            <>
              <DialogHeader>
                <DialogTitle className="font-serif">
                  {t("students_credential_display")}
                </DialogTitle>
                <DialogDescription>
                  {t("students_credential_warning")}
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label>{t("students_credential_username")}</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="flex-1 rounded-md border border-line bg-surface-2 px-3 py-2 text-sm font-mono">
                      {reissueResult.username}
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handleCopy(reissueResult.username, "username")
                      }
                    >
                      {copied === "username" ? (
                        <Check className="size-4 text-success" />
                      ) : (
                        <Copy className="size-4" />
                      )}
                      <span className="ml-1">
                        {copied === "username"
                          ? t("students_credential_copied")
                          : t("students_credential_copy")}
                      </span>
                    </Button>
                  </div>
                </div>
                <div>
                  <Label>{t("students_credential_password")}</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="flex-1 rounded-md border border-line bg-surface-2 px-3 py-2 text-sm font-mono">
                      {reissueResult.temp_password}
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handleCopy(reissueResult.temp_password, "password")
                      }
                    >
                      {copied === "password" ? (
                        <Check className="size-4 text-success" />
                      ) : (
                        <Copy className="size-4" />
                      )}
                      <span className="ml-1">
                        {copied === "password"
                          ? t("students_credential_copied")
                          : t("students_credential_copy")}
                      </span>
                    </Button>
                  </div>
                </div>
              </div>
              <DialogFooter className="mt-4">
                <Button onClick={handleCloseReissue}>
                  {t("cancel")}
                </Button>
              </DialogFooter>
            </>
          ) : (
            /* Confirmation step */
            <>
              <DialogHeader>
                <DialogTitle className="font-serif">
                  {t("students_credential_reissue")}
                </DialogTitle>
                <DialogDescription>
                  {t("students_credential_reissue_warning")}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter className="mt-4">
                <Button variant="outline" onClick={handleCloseReissue}>
                  {t("cancel")}
                </Button>
                <Button
                  onClick={handleReissue}
                  disabled={reissueCreds.isPending}
                >
                  {reissueCreds.isPending
                    ? t("saving")
                    : t("students_credential_reissue")}
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-lg border px-3 py-[7px] text-xs font-semibold transition-colors",
        active
          ? "border-brand-600 bg-brand-600 text-white"
          : "border-line bg-surface text-ink-600 hover:text-ink-900"
      )}
    >
      {children}
    </button>
  );
}
