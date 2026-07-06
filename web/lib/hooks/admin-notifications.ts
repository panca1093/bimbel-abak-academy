"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";

export interface AdminNotification {
  id: string;
  type: string;
  order_id: string;
  student_name: string;
  amount: number;
  created_at: string;
  read: boolean;
}

export interface AdminNotificationsResponse {
  data: AdminNotification[];
  next_cursor: string;
}

export interface AdminNotifsFilters {
  unreadOnly?: boolean;
  cursor?: string;
}

export const adminNotifsKeys = {
  all: ["admin", "notifications"] as const,
  list: (filters?: AdminNotifsFilters) => {
    const parts: string[] = ["admin", "notifications", "list"];
    if (filters?.unreadOnly) parts.push("unread");
    if (filters?.cursor) parts.push(filters.cursor);
    return parts as unknown as readonly string[];
  },
};

function buildNotifsQuery(filters?: AdminNotifsFilters): string {
  const params = new URLSearchParams();
  if (filters?.unreadOnly) params.set("unread_only", "true");
  if (filters?.cursor) params.set("cursor", filters.cursor);
  const qs = params.toString();
  return qs ? `/admin/notifications?${qs}` : "/admin/notifications";
}

export function useAdminNotifications(filters?: AdminNotifsFilters) {
  return useQuery({
    queryKey: adminNotifsKeys.list(filters),
    queryFn: () => authFetch<AdminNotificationsResponse>(buildNotifsQuery(filters)),
  });
}

export function useMarkNotificationRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<{ message: string }>(`/admin/notifications/${encodeURIComponent(id)}/read`, {
        method: "PATCH",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminNotifsKeys.all });
    },
  });
}
