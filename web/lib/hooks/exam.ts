"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { API_BASE, authFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type {
  RegistrationDetail,
  RegistrationListItem,
  SessionStartPayload,
  SessionState,
  SessionAnswerInput,
  SubmitResult,
  CheckInResult,
  SessionResult,
  ExamLeaderboardEntry,
} from "@/lib/types";

export const examKeys = {
  all: ["exam"] as const,
  lists: () => [...examKeys.all, "list"] as const,
  list: () => [...examKeys.lists()] as const,
  details: () => [...examKeys.all, "detail"] as const,
  detail: (id: string) => [...examKeys.details(), id] as const,
  sessions: () => [...examKeys.all, "session"] as const,
  session: (id: string) => [...examKeys.sessions(), id] as const,
  results: () => [...examKeys.all, "result"] as const,
  result: (id: string) => [...examKeys.results(), id] as const,
  leaderboards: () => [...examKeys.all, "leaderboard"] as const,
  leaderboard: (id: string, filter?: { cursor?: string; limit?: number }) =>
    [...examKeys.leaderboards(), id, filter ?? {}] as const,
};

export function useRegistrations() {
  return useQuery({
    queryKey: examKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: RegistrationListItem[] }>(
        "/exam/registrations"
      );
      return res.data ?? [];
    },
  });
}

export function useRegistration(id: string | undefined) {
  return useQuery({
    queryKey: examKeys.detail(id ?? ""),
    queryFn: () =>
      authFetch<RegistrationDetail>(
        `/exam/registrations/${encodeURIComponent(id!)}`
      ),
    enabled: Boolean(id),
  });
}

export async function downloadCard(id: string): Promise<void> {
  const token = useAuthStore.getState().token;
  const res = await fetch(
    `${API_BASE}/exam/registrations/${encodeURIComponent(id)}/card`,
    {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    }
  );
  if (!res.ok) {
    throw new ApiError(
      `HTTP_${res.status}`,
      `Failed to download card: ${res.status}`,
      res.status
    );
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

// ── Session hooks (FR26) ─────────────────────────────────────────────────

export function useCheckIn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ token }: { token: string }) =>
      authFetch<CheckInResult>(
        `/exam/checkin`,
        { method: "POST", body: JSON.stringify({ token }) },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.details() });
    },
  });
}

export function useStartSession() {
  return useMutation({
    mutationFn: (registrationId: string) =>
      authFetch<SessionStartPayload>("/exam/sessions", {
        method: "POST",
        body: JSON.stringify({ registration_id: registrationId }),
      }),
  });
}

export function useReconnectSession(sessionId: string | undefined) {
  return useQuery({
    queryKey: examKeys.session(sessionId ?? ""),
    queryFn: () =>
      authFetch<SessionState>(
        `/exam/sessions/${encodeURIComponent(sessionId!)}`,
      ),
    enabled: Boolean(sessionId),
  });
}

export function useSaveAnswers(sessionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (answers: SessionAnswerInput[]) =>
      authFetch<void>(`/exam/sessions/${encodeURIComponent(sessionId)}/answers`, {
        method: "PATCH",
        body: JSON.stringify({ answers }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.session(sessionId) });
    },
  });
}

export function useSubmitSession(sessionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      authFetch<SubmitResult>(
        `/exam/sessions/${encodeURIComponent(sessionId)}/submit`,
        { method: "POST" },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.session(sessionId) });
    },
  });
}

export function useSessionResult(sessionId: string | undefined) {
  return useQuery({
    queryKey: examKeys.result(sessionId ?? ""),
    queryFn: () =>
      authFetch<SessionResult>(
        `/exam/sessions/${encodeURIComponent(sessionId!)}/result`,
      ),
    enabled: Boolean(sessionId),
  });
}

export function useLogViolation(sessionId: string) {
  return useMutation({
    mutationFn: (violationType: string) =>
      authFetch<void>(
        `/exam/sessions/${encodeURIComponent(sessionId)}/violations`,
        {
          method: "POST",
          body: JSON.stringify({ violation_type: violationType }),
        },
      ),
  });
}

export function useSessionLeaderboard(
  sessionId: string | undefined,
  filter?: { cursor?: string; limit?: number },
) {
  return useQuery({
    queryKey: examKeys.leaderboard(sessionId ?? "", filter),
    queryFn: () => {
      const base = `/exam/sessions/${encodeURIComponent(sessionId!)}/leaderboard`;
      if (!filter) return authFetch<{ data: ExamLeaderboardEntry[] }>(base);
      const params = new URLSearchParams();
      if (filter.cursor) params.set("cursor", filter.cursor);
      if (filter.limit !== undefined) params.set("limit", String(filter.limit));
      return authFetch<{ data: ExamLeaderboardEntry[] }>(`${base}?${params.toString()}`);
    },
    enabled: Boolean(sessionId),
  });
}
