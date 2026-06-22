"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  ShieldCheck,
  MoreHorizontal,
  Lock,
  Mail,
  Search,
} from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
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
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";
import {
  useAdminAccounts,
  useCreateAdminAccount,
  useChangeAccountRole,
  useChangeAccountStatus,
  useResetAccountPassword,
} from "@/lib/hooks/admin-accounts";
import type { AdminAccount, AdminAccountRole, AdminAccountStatus } from "@/lib/types";

const ROLE_LABEL: Record<AdminAccountRole, string> = {
  super_admin: "Super Admin",
  admin_exam: "Admin Exam",
  admin_school: "School Operator",
  admin_store: "Store Manager",
};

const ROLE_TONE: Record<AdminAccountRole, string> = {
  super_admin: "bg-danger-bg text-danger border-danger",
  admin_exam: "bg-info-bg text-info border-info",
  admin_school: "bg-violet-bg text-violet border-violet",
  admin_store: "bg-success-bg text-success border-success",
};

const STATUS_TONE: Record<AdminAccountStatus, string> = {
  active: "bg-success-bg text-success border-success",
  deactivated: "bg-danger-bg text-danger border-danger",
};

const ROLES: AdminAccountRole[] = ["super_admin", "admin_exam", "admin_school", "admin_store"];

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

export default function SystemAccountsPage() {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<AdminAccountRole | "all">("all");
  const [statusFilter, setStatusFilter] = useState<AdminAccountStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState({
    name: "",
    email: "",
    role: "admin_store" as AdminAccountRole,
    password: "",
  });
  const [roleChangeTarget, setRoleChangeTarget] = useState<AdminAccount | null>(null);
  const [roleChangeRole, setRoleChangeRole] = useState<AdminAccountRole>("admin_store");

  const { data: accounts = [], isLoading, error } = useAdminAccounts(
    roleFilter === "all" ? undefined : roleFilter,
    statusFilter === "all" ? undefined : statusFilter
  );

  const createAccount = useCreateAdminAccount();
  const changeRole = useChangeAccountRole();
  const changeStatus = useChangeAccountStatus();
  const resetPwd = useResetAccountPassword();

  const rows = useMemo(() => {
    if (search.trim() === "") return accounts;
    const q = search.toLowerCase();
    return accounts.filter(
      (a) => a.name.toLowerCase().includes(q) || (a.email ?? "").toLowerCase().includes(q)
    );
  }, [search, accounts]);

  const stats = useMemo(
    () => ({
      total: accounts.length,
      active: accounts.filter((a) => a.status === "active").length,
      deactivated: accounts.filter((a) => a.status === "deactivated").length,
    }),
    [accounts]
  );

  const handleCreate = async () => {
    if (!createForm.name || !createForm.email || !createForm.password) {
      toast.error("Semua field harus diisi");
      return;
    }
    try {
      await createAccount.mutateAsync({
        name: createForm.name,
        email: createForm.email,
        role: createForm.role,
        password: createForm.password,
      });
      toast.success("Akun berhasil dibuat");
      setCreateOpen(false);
      setCreateForm({ name: "", email: "", role: "admin_store", password: "" });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal membuat akun";
      toast.error(msg);
    }
  };

  const handleRoleChange = async () => {
    if (!roleChangeTarget) return;
    try {
      await changeRole.mutateAsync({ id: roleChangeTarget.id, role: roleChangeRole });
      toast.success("Peran berhasil diubah");
      setRoleChangeTarget(null);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal mengubah peran";
      toast.error(msg);
    }
  };

  const handleStatusToggle = async (account: AdminAccount) => {
    const newStatus: AdminAccountStatus = account.status === "active" ? "deactivated" : "active";
    const label = newStatus === "active" ? "Aktifkan" : "Nonaktifkan";
    try {
      await changeStatus.mutateAsync({ id: account.id, status: newStatus });
      toast.success(`${label} berhasil`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : `Gagal ${label.toLowerCase()}`;
      toast.error(msg);
    }
  };

  const handleResetPassword = async (account: AdminAccount) => {
    try {
      await resetPwd.mutateAsync(account.id);
      toast.success("Email reset password telah dikirim");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal mereset password";
      toast.error(msg);
    }
  };

  if (isLoading) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader icon={ShieldCheck} title="Akun Pengguna" description="Memuat…" />
        <div className="py-12 text-center text-ink-500">Memuat data…</div>
      </div>
    );
  }

  if (error) {
    const msg =
      (error as { code?: string })?.code === "forbidden"
        ? "Akses ditolak. Hanya Super Admin yang dapat mengakses halaman ini."
        : "Gagal memuat data. Coba refresh halaman.";
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader icon={ShieldCheck} title="Akun Pengguna" description="Terjadi kesalahan" />
        <div className="py-12 text-center text-ink-500">{msg}</div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={ShieldCheck}
        title="Akun Pengguna"
        description="Kelola akun dan hak akses admin."
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total akun" value={String(stats.total)} />
        <StatCard label="Aktif" value={String(stats.active)} />
        <StatCard label="Nonaktif" value={String(stats.deactivated)} />
      </div>

      <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip active={roleFilter === "all"} onClick={() => setRoleFilter("all")}>
            {t("tab_all")}
          </FilterChip>
          {ROLES.map((r) => (
            <FilterChip
              key={r}
              active={roleFilter === r}
              onClick={() => setRoleFilter(r)}
            >
              {ROLE_LABEL[r]}
            </FilterChip>
          ))}
        </div>
        <div className="flex items-center gap-2 lg:ml-auto">
          <Select
            value={statusFilter}
            onValueChange={(v) => setStatusFilter(v as AdminAccountStatus | "all")}
          >
            <SelectTrigger className="h-9 w-[140px] text-xs">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua status</SelectItem>
              <SelectItem value="active">Aktif</SelectItem>
              <SelectItem value="deactivated">Nonaktif</SelectItem>
            </SelectContent>
          </Select>
          <Search className="size-4 text-ink-400" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama / email…"
            className="h-9 w-[200px] text-xs"
          />
        </div>
      </div>

      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">Akun</th>
                <th className="px-4 py-3">Peran</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Dibuat</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-ink-500">
                    Tidak ada akun ditemukan.
                  </td>
                </tr>
              )}
              {rows.map((a) => (
                <tr key={a.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Avatar size="sm">
                        <AvatarFallback className="bg-brand-50 text-brand-700 text-xs">
                          {initials(a.name)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-ink-900">{a.name}</div>
                        <div className="flex items-center gap-1 text-[11px] text-ink-500">
                          <Mail className="size-3" />
                          {a.email ?? "—"}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn("text-[11px] font-semibold", ROLE_TONE[a.role])}
                    >
                      {ROLE_LABEL[a.role]}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold capitalize",
                        STATUS_TONE[a.status]
                      )}
                    >
                      {a.status === "active" ? "Aktif" : "Nonaktif"}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {a.created_at
                      ? new Date(a.created_at).toLocaleString("id-ID", {
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
                          onClick={() => {
                            setRoleChangeTarget(a);
                            setRoleChangeRole(a.role);
                          }}
                        >
                          <Mail className="mr-2 size-4" />
                          Ganti peran
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleStatusToggle(a)}>
                          <Lock className="mr-2 size-4" />
                          {a.status === "active" ? "Nonaktifkan" : "Aktifkan"}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleResetPassword(a)}>
                          <Lock className="mr-2 size-4" />
                          Reset password
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create account dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">Buat akun admin</DialogTitle>
            <DialogDescription>Isi data akun admin baru.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Nama</Label>
              <Input
                value={createForm.name}
                onChange={(e) => setCreateForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="Nama lengkap"
              />
            </div>
            <div>
              <Label>Email</Label>
              <Input
                type="email"
                value={createForm.email}
                onChange={(e) => setCreateForm((f) => ({ ...f, email: e.target.value }))}
                placeholder="email@example.com"
              />
            </div>
            <div>
              <Label>Peran</Label>
              <Select
                value={createForm.role}
                onValueChange={(v) => setCreateForm((f) => ({ ...f, role: v as AdminAccountRole }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih peran" />
                </SelectTrigger>
                <SelectContent>
                  {ROLES.map((r) => (
                    <SelectItem key={r} value={r}>
                      {ROLE_LABEL[r]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Password</Label>
              <Input
                type="password"
                value={createForm.password}
                onChange={(e) => setCreateForm((f) => ({ ...f, password: e.target.value }))}
                placeholder="Minimal 8 karakter"
              />
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              {t("cancel")}
            </Button>
            <Button onClick={handleCreate} disabled={createAccount.isPending}>
              {createAccount.isPending ? "Menyimpan…" : t("create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Role change dialog */}
      <Dialog
        open={roleChangeTarget !== null}
        onOpenChange={(open) => {
          if (!open) setRoleChangeTarget(null);
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="font-serif">Ganti peran</DialogTitle>
            <DialogDescription>
              {roleChangeTarget ? `Ubah peran untuk ${roleChangeTarget.name}` : ""}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Peran baru</Label>
              <Select
                value={roleChangeRole}
                onValueChange={(v) => setRoleChangeRole(v as AdminAccountRole)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih peran" />
                </SelectTrigger>
                <SelectContent>
                  {ROLES.map((r) => (
                    <SelectItem key={r} value={r}>
                      {ROLE_LABEL[r]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="outline" onClick={() => setRoleChangeTarget(null)}>
              {t("cancel")}
            </Button>
            <Button onClick={handleRoleChange} disabled={changeRole.isPending}>
              {changeRole.isPending ? "Menyimpan…" : "Simpan"}
            </Button>
          </DialogFooter>
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
