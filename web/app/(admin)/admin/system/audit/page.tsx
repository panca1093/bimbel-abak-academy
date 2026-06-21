"use client";

import { useMemo, useState } from "react";
import { ShieldCheck, ClipboardList, Filter, Clock } from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";

type AuditAction = "create" | "update" | "delete" | "login" | "publish";

interface AuditEntry {
  id: string;
  actor: string;
  email: string;
  action: AuditAction;
  target: string;
  targetType: string;
  ip?: string;
  at: string;
}

const INITIAL_AUDITS: AuditEntry[] = [
  {
    id: "AUD-3001",
    actor: "Rina Wijayanti",
    email: "rina.w@example.com",
    action: "publish",
    target: "Tryout SNBT #12",
    targetType: "Exam",
    ip: "203.0.113.42",
    at: "2026-06-20T10:15:00",
  },
  {
    id: "AUD-3002",
    actor: "Hendra Gunawan",
    email: "hendra.g@example.com",
    action: "update",
    target: "Modul Fisika Kelas 12",
    targetType: "Product",
    ip: "203.0.113.88",
    at: "2026-06-20T09:42:00",
  },
  {
    id: "AUD-3003",
    actor: "Sri Wahyuni",
    email: "sri.w@example.com",
    action: "create",
    target: "Andi Pratama",
    targetType: "Student",
    at: "2026-06-20T08:30:00",
  },
  {
    id: "AUD-3004",
    actor: "Saifullah Panca",
    email: "saifullah.panca@amartha.com",
    action: "login",
    target: "Super Admin Portal",
    targetType: "System",
    ip: "203.0.113.10",
    at: "2026-06-19T18:20:00",
  },
  {
    id: "AUD-3005",
    actor: "Rina Wijayanti",
    email: "rina.w@example.com",
    action: "delete",
    target: "Soal AQ-1099",
    targetType: "Question",
    ip: "203.0.113.42",
    at: "2026-06-19T14:10:00",
  },
];

const ACTION_TONE: Record<AuditAction, string> = {
  create: "bg-success-bg text-success border-success",
  update: "bg-info-bg text-info border-info",
  delete: "bg-danger-bg text-danger border-danger",
  login: "bg-ink-100 text-ink-600 border-line",
  publish: "bg-warn-bg text-warn border-warn",
};

const ACTION_LABEL: Record<AuditAction, string> = {
  create: "Create",
  update: "Update",
  delete: "Delete",
  login: "Login",
  publish: "Publish",
};

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

export default function SystemAuditPage() {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");

  const rows = useMemo(() => {
    if (search.trim() === "") return INITIAL_AUDITS;
    const q = search.toLowerCase();
    return INITIAL_AUDITS.filter(
      (a) =>
        a.actor.toLowerCase().includes(q) ||
        a.target.toLowerCase().includes(q) ||
        a.targetType.toLowerCase().includes(q)
    );
  }, [search]);

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={ClipboardList}
        title="Audit Log"
        description="Riwayat aktivitas admin di seluruh platform."
        actions={
          <div className="flex items-center gap-2">
            <Filter className="size-4 text-ink-400" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari aktor atau target…"
              className="h-9 w-[240px] text-xs"
            />
          </div>
        }
      />

      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">Waktu</th>
                <th className="px-4 py-3">Aktor</th>
                <th className="px-4 py-3">Aksi</th>
                <th className="px-4 py-3">Target</th>
                <th className="px-4 py-3">IP</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((a) => (
                <tr key={a.id} className="hover:bg-surface-2">
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Clock className="size-3" />
                      {new Date(a.at).toLocaleString("id-ID", {
                        day: "2-digit",
                        month: "short",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Avatar size="sm">
                        <AvatarFallback className="bg-brand-50 text-brand-700 text-xs">
                          {initials(a.actor)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-ink-900">{a.actor}</div>
                        <div className="text-[11px] text-ink-500">{a.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn("text-[11px] font-semibold", ACTION_TONE[a.action])}
                    >
                      {ACTION_LABEL[a.action]}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <div className="font-medium text-ink-900">{a.target}</div>
                    <div className="text-[11px] text-ink-500">{a.targetType}</div>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-ink-600">
                    {a.ip || "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
