"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { ExamTopic } from "@/lib/types";

export const adminTopicsKeys = {
  all: ["admin", "topics"] as const,
  lists: () => [...adminTopicsKeys.all, "list"] as const,
  list: (subject: string | undefined) =>
    [...adminTopicsKeys.lists(), subject ?? "all"] as const,
  detail: (id: string) => [...adminTopicsKeys.all, "detail", id] as const,
};

export interface AdminCreateTopicInput {
  name: string;
  subject: string;
}

export interface AdminUpdateTopicInput {
  name?: string;
  subject?: string;
}

function buildListPath(subject?: string): string {
  return subject ? `/admin/topics?subject=${encodeURIComponent(subject)}` : "/admin/topics";
}

export function useTopics(subject?: string) {
  return useQuery({
    queryKey: adminTopicsKeys.list(subject),
    queryFn: () => authFetch<{ data: ExamTopic[]; next_cursor?: string }>(buildListPath(subject)),
  });
}

export function useCreateTopic() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateTopicInput) =>
      authFetch<ExamTopic>("/admin/topics", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminTopicsKeys.lists() });
    },
  });
}

export function useUpdateTopic(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminUpdateTopicInput) =>
      authFetch<ExamTopic>(`/admin/topics/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminTopicsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminTopicsKeys.detail(id) });
    },
  });
}

export function useDeleteTopic() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<void>(`/admin/topics/${encodeURIComponent(id)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminTopicsKeys.lists() });
    },
  });
}
