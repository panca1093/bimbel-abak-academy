"use client";

import { useQuery } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { AuditLogEntry } from "@/lib/types";

export interface AuditLogFilters {
  actor_id?: string;
  from?: string;
  to?: string;
  target_type?: string;
  q?: string;
}

export const adminAuditKeys = {
  all: ["admin", "audit"] as const,
  list: (filters?: AuditLogFilters) => {
    if (!filters) return [...adminAuditKeys.all, "list"] as const;
    const parts: string[] = [];
    if (filters.actor_id) parts.push(`actor=${filters.actor_id}`);
    if (filters.from) parts.push(`from=${filters.from}`);
    if (filters.to) parts.push(`to=${filters.to}`);
    if (filters.target_type) parts.push(`type=${filters.target_type}`);
    if (filters.q) parts.push(`q=${filters.q}`);
    const key = parts.length ? parts.join("&") : "all";
    return [...adminAuditKeys.all, "list", key] as const;
  },
};

export function useAdminAuditLog(filters?: AuditLogFilters) {
  return useQuery({
    queryKey: adminAuditKeys.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.actor_id) params.set("actor_id", filters.actor_id);
      if (filters?.from) params.set("from", filters.from);
      if (filters?.to) params.set("to", filters.to);
      if (filters?.target_type) params.set("target_type", filters.target_type);
      if (filters?.q) params.set("q", filters.q);
      const query = params.toString();
      const path = query ? `/admin/system/audit?${query}` : "/admin/system/audit";
      const res = await authFetch<{ data: AuditLogEntry[]; next_cursor?: string }>(path);
      return res.data ?? [];
    },
  });
}
