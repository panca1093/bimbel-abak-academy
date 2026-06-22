"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  ShieldCheck,
  Shield,
  MoreHorizontal,
  Edit,
  Lock,
  Mail,
  Search,
} from "lucide-react";
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

type SystemRole = "super_admin" | "admin_exam" | "admin_school" | "admin_store";
type AccountStatus = "active" | "suspended" | "pending";

interface Account {
  id: string;
  name: string;
  email: string;
  role: SystemRole;
  status: AccountStatus;
  lastActive?: string;
}

const INITIAL_ACCOUNTS: Account[] = [
  {
    id: "ACC-9001",
    name: "Saifullah Panca",
    email: "saifullah.panca@amartha.com",
    role: "super_admin",
    status: "active",
    lastActive: "2026-06-19T09:30",
  },
  {
    id: "ACC-9002",
    name: "Rina Wijayanti",
    email: "rina.w@example.com",
    role: "admin_exam",
    status: "active",
    lastActive: "2026-06-18T16:45",
  },
  {
    id: "ACC-9003",
    name: "Hendra Gunawan",
    email: "hendra.g@example.com",
    role: "admin_store",
    status: "active",
    lastActive: "2026-06-17T11:20",
  },
  {
    id: "ACC-9004",
    name: "Sri Wahyuni",
    email: "sri.w@example.com",
    role: "admin_school",
    status: "pending",
  },
  {
    id: "ACC-9005",
    name: "Budi Admin Lama",
    email: "budi.lama@example.com",
    role: "admin_exam",
    status: "suspended",
  },
];

const ROLE_LABEL: Record<SystemRole, string> = {
  super_admin: "Super Admin",
  admin_exam: "Admin Exam",
  admin_school: "School Operator",
  admin_store: "Store Manager",
};

const ROLE_TONE: Record<SystemRole, string> = {
  super_admin: "bg-danger-bg text-danger border-danger",
  admin_exam: "bg-info-bg text-info border-info",
  admin_school: "bg-violet-bg text-violet border-violet",
  admin_store: "bg-success-bg text-success border-success",
};

const STATUS_TONE: Record<AccountStatus, string> = {
  active: "bg-success-bg text-success border-success",
  suspended: "bg-danger-bg text-danger border-danger",
  pending: "bg-warn-bg text-warn border-warn",
};

const ROLES: SystemRole[] = ["super_admin", "admin_exam", "admin_school", "admin_store"];

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
  const [roleFilter, setRoleFilter] = useState<SystemRole | "all">("all");
  const [statusFilter, setStatusFilter] = useState<AccountStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);

  const rows = useMemo(() => {
    return INITIAL_ACCOUNTS.filter((a) => {
      const q = search.toLowerCase();
      const matchesSearch =
        search.trim() === "" ||
        a.name.toLowerCase().includes(q) ||
        a.email.toLowerCase().includes(q);
      const matchesRole = roleFilter === "all" || a.role === roleFilter;
      const matchesStatus = statusFilter === "all" || a.status === statusFilter;
      return matchesSearch && matchesRole && matchesStatus;
    });
  }, [search, roleFilter, statusFilter]);

  const stats = useMemo(() => {
    return {
      total: INITIAL_ACCOUNTS.length,
      active: INITIAL_ACCOUNTS.filter((a) => a.status === "active").length,
      pending: INITIAL_ACCOUNTS.filter((a) => a.status === "pending").length,
      suspended: INITIAL_ACCOUNTS.filter((a) => a.status === "suspended").length,
    };
  }, []);

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
        <StatCard label="Pending" value={String(stats.pending)} />
        <StatCard label="Suspended" value={String(stats.suspended)} />
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
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as AccountStatus | "all")}>
            <SelectTrigger className="h-9 w-[140px] text-xs">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua status</SelectItem>
              <SelectItem value="active">Aktif</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="suspended">Suspended</SelectItem>
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
                <th className="px-4 py-3">Terakhir aktif</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
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
                          {a.email}
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
                      className={cn("text-[11px] font-semibold capitalize", STATUS_TONE[a.status])}
                    >
                      {a.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {a.lastActive
                      ? new Date(a.lastActive).toLocaleString("id-ID", {
                          day: "2-digit",
                          month: "short",
                          hour: "2-digit",
                          minute: "2-digit",
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
                        <DropdownMenuItem onClick={() => setCreateOpen(true)}>
                          <Edit className="mr-2 size-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem>
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

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">{t("create")} akun</DialogTitle>
            <DialogDescription>
              Form undangan akun admin akan tersedia di iterasi berikutnya.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Nama</Label>
              <Input placeholder="Nama lengkap" />
            </div>
            <div>
              <Label>Email</Label>
              <Input type="email" placeholder="email@example.com" />
            </div>
            <div>
              <Label>Peran</Label>
              <Select>
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
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                {t("cancel")}
              </Button>
              <Button onClick={() => setCreateOpen(false)}>{t("create")}</Button>
            </div>
          </div>
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
