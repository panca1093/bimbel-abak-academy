"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type {
  AdminStudent,
  StudentCredentials,
  StudentRegistrationInput,
  StudentRegistrationResult,
} from "@/lib/types";

export const adminStudentsKeys = {
  all: ["admin", "students"] as const,
  list: (status?: string, q?: string, cursor?: string, limit?: number, schoolId?: string) =>
    [...adminStudentsKeys.all, "list", status ?? "all", q ?? "", cursor ?? "initial", limit ?? 20, schoolId ?? ""] as const,
};

export function useAdminStudents(
  opts?: { status?: string; q?: string; cursor?: string; limit?: number; schoolId?: string; enabled?: boolean }
) {
  const { status, q, cursor, limit, schoolId, enabled } = opts ?? {};
  return useQuery({
    queryKey: adminStudentsKeys.list(status, q, cursor, limit, schoolId),
    enabled: enabled ?? true,
    queryFn: async () => {
      const params = new URLSearchParams();
      if (status) params.set("status", status);
      if (q) params.set("q", q);
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (schoolId) params.set("school_id", schoolId);
      const query = params.toString();
      const path = query ? `/admin/students?${query}` : "/admin/students";
      return authFetch<{ data: AdminStudent[]; next_cursor?: string }>(path);
    },
  });
}

export function useRegisterStudent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: StudentRegistrationInput) =>
      authFetch<StudentRegistrationResult>("/admin/students", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminStudentsKeys.all });
    },
  });
}

export function useChangeStudentStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      authFetch<{ message: string }>(`/admin/students/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminStudentsKeys.all });
    },
  });
}

export function useReissueStudentCredentials() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<StudentCredentials>(`/admin/students/${encodeURIComponent(id)}/credentials`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminStudentsKeys.all });
    },
  });
}
