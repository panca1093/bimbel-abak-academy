"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import { examKeys } from "@/lib/hooks/exam";
import type {
  ExamListItem,
  ExamDetail,
  CreateExamPayload,
  UpdateExamPayload,
  GradingSessionItem,
  GradingEssayItem,
} from "@/lib/types";

export const adminExamsKeys = {
  all: ["adminExams"] as const,
  lists: () => [...adminExamsKeys.all, "list"] as const,
  list: (filter: AdminExamsFilters | undefined) =>
    [...adminExamsKeys.lists(), filter ?? {}] as const,
  details: () => [...adminExamsKeys.all, "detail"] as const,
  detail: (id: string) => [...adminExamsKeys.details(), id] as const,
  gradingLists: () => [...adminExamsKeys.all, "grading"] as const,
  grading: (examId: string) => [...adminExamsKeys.gradingLists(), examId] as const,
  sessionEssays: (sessionId: string) =>
    [...adminExamsKeys.all, "sessionEssays", sessionId] as const,
};

export interface GradeEssayInput {
  question_id: string;
  score: number;
  comment?: string;
}

export interface AdminExamsFilters {
  cursor?: string;
  limit?: number;
}

function buildListPath(filters?: AdminExamsFilters): string {
  if (!filters) return "/admin/exams";
  const params = new URLSearchParams();
  if (filters.cursor) params.set("cursor", filters.cursor);
  if (filters.limit !== undefined) params.set("limit", String(filters.limit));
  const query = params.toString();
  return query ? `/admin/exams?${query}` : "/admin/exams";
}

export function useExams(filter?: AdminExamsFilters) {
  return useQuery({
    queryKey: adminExamsKeys.list(filter),
    queryFn: () =>
      authFetch<{ data: ExamListItem[]; next_cursor?: string }>(buildListPath(filter)),
  });
}

export function useExam(id: string | undefined) {
  return useQuery({
    queryKey: adminExamsKeys.detail(id ?? ""),
    queryFn: () =>
      authFetch<ExamDetail>(`/admin/exams/${encodeURIComponent(id!)}`),
    enabled: Boolean(id),
  });
}

export function useCreateExam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateExamPayload) =>
      authFetch<{ exam: ExamDetail; product: { id: string } }>("/admin/exams", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminExamsKeys.details() });
    },
  });
}

export function useUpdateExam(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateExamPayload) =>
      authFetch<ExamDetail>(`/admin/exams/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminExamsKeys.detail(id) });
    },
  });
}

export function useReplaceExamTests(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (testIds: string[]) =>
      authFetch<void>(`/admin/exams/${encodeURIComponent(id)}/tests`, {
        method: "PUT",
        body: JSON.stringify(testIds),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminExamsKeys.detail(id) });
    },
  });
}

export function useUpdateExamPrice(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (price: number) =>
      authFetch<{ price: number }>(
        `/admin/exams/${encodeURIComponent(id)}/price`,
        {
          method: "PATCH",
          body: JSON.stringify({ price }),
        },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminExamsKeys.detail(id) });
    },
  });
}

export function usePublishExam(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      authFetch<{ message: string }>(
        `/admin/exams/${encodeURIComponent(id)}/publish`,
        {
          method: "POST",
        },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.lists() });
      qc.invalidateQueries({ queryKey: adminExamsKeys.detail(id) });
    },
  });
}

export function useGradingSessions(examId: string | undefined) {
  return useQuery({
    queryKey: adminExamsKeys.grading(examId ?? ""),
    queryFn: () =>
      authFetch<{ data: GradingSessionItem[] }>(
        `/admin/exams/${encodeURIComponent(examId!)}/grading`,
      ),
    enabled: Boolean(examId),
  });
}

export function useSessionEssays(sessionId: string | undefined) {
  return useQuery({
    queryKey: adminExamsKeys.sessionEssays(sessionId ?? ""),
    queryFn: () =>
      authFetch<{ data: GradingEssayItem[] }>(
        `/admin/sessions/${encodeURIComponent(sessionId!)}/essays`,
      ),
    enabled: Boolean(sessionId),
  });
}

export function useGradeEssay(sessionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: GradeEssayInput) =>
      authFetch<{ status: string; score: number }>(
        `/admin/sessions/${encodeURIComponent(sessionId)}/grade`,
        {
          method: "POST",
          body: JSON.stringify(input),
        },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminExamsKeys.sessionEssays(sessionId) });
      qc.invalidateQueries({ queryKey: adminExamsKeys.gradingLists() });
      qc.invalidateQueries({ queryKey: examKeys.result(sessionId) });
    },
  });
}