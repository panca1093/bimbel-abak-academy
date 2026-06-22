"use client";

import { useState } from "react";
import { ShieldCheck, ClipboardList, Filter, Clock } from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { useAdminAuditLog } from "@/lib/hooks/admin-audit";
import type { AuditLogEntry } from "@/lib/types";

function actionLabel(action: string): string {
  const verb = action.split(".").pop() ?? action;
  return verb.replace(/_/g, " ");
}

function actionTone(action: string): string {
  if (action.includes("create") || action.includes("add"))
    return "bg-success-bg text-success border-success";
  if (action.includes("delete") || action.includes("remove"))
    return "bg-danger-bg text-danger border-danger";
  if (action.includes("change") || action.includes("update"))
    return "bg-info-bg text-info border-info";
  return "bg-ink-100 text-ink-600 border-line";
}

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

  const { data: entries = [], isLoading, error } = useAdminAuditLog(
    search.trim() ? { q: search } : undefined
  );

  if (isLoading) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader icon={ClipboardList} title="Audit Log" description="Memuat…" />
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
      <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={ClipboardList}
          title="Audit Log"
          description="Terjadi kesalahan"
        />
        <div className="py-12 text-center text-ink-500">{msg}</div>
      </div>
    );
  }

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
              placeholder="Cari aksi…"
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
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {entries.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-ink-500">
                    Tidak ada catatan audit.
                  </td>
                </tr>
              )}
              {entries.map((a) => (
                <tr key={a.id} className="hover:bg-surface-2">
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Clock className="size-3" />
                      {new Date(a.created_at).toLocaleString("id-ID", {
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
                          {initials(a.actor_name ?? "System")}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-ink-900">
                          {a.actor_name ?? "System"}
                        </div>
                        {a.actor_email && (
                          <div className="text-[11px] text-ink-500">{a.actor_email}</div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold capitalize",
                        actionTone(a.action)
                      )}
                    >
                      {actionLabel(a.action)}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <div className="font-medium text-ink-900">{a.target_id}</div>
                    <div className="text-[11px] text-ink-500">{a.target_type}</div>
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
