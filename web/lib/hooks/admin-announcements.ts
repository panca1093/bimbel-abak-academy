"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";

export interface Announcement {
  id: string;
  title: string;
  message: string;
  type: string;
  recipients: string;
  status: string;
  scheduled_at: string | null;
  sent_at: string | null;
  recipient_count: number | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAnnouncementInput {
  title: string;
  message: string;
  type: string;
  recipients: string;
  status?: string;
  scheduled_at?: string | null;
}

export interface UpdateAnnouncementInput {
  title?: string;
  message?: string;
  type?: string;
  recipients?: string;
  scheduled_at?: string | null;
}

export const adminAnnouncementKeys = {
  all: ["admin", "announcements"] as const,
  list: () => [...adminAnnouncementKeys.all, "list"] as const,
};

export function useAdminAnnouncements() {
  return useQuery({
    queryKey: adminAnnouncementKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: Announcement[] }>("/admin/notifications/announcements");
      return res.data ?? [];
    },
  });
}

export function useCreateAnnouncement() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateAnnouncementInput) =>
      authFetch<Announcement>("/admin/notifications/announcements", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAnnouncementKeys.list() });
    },
  });
}

export function useUpdateAnnouncement() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateAnnouncementInput }) =>
      authFetch<Announcement>(`/admin/notifications/announcements/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAnnouncementKeys.list() });
    },
  });
}

export function useDeleteAnnouncement() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<{ message: string }>(`/admin/notifications/announcements/${encodeURIComponent(id)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAnnouncementKeys.list() });
    },
  });
}

export function useSendAnnouncement() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<Announcement>(`/admin/notifications/announcements/${encodeURIComponent(id)}/send`, {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAnnouncementKeys.list() });
    },
  });
}
