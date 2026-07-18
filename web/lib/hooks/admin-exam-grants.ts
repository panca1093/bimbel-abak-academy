"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { CrossSchoolStudent } from "@/lib/types";

// ── Query keys ────────────────────────────────────────────────────────────

export const examGrantKeys = {
  all: ["admin", "examGrants"] as const,
  search: (q?: string, schoolId?: string, jenjang?: string, grade?: string) =>
    [...examGrantKeys.all, "search", q ?? "", schoolId ?? "", jenjang ?? "", grade ?? ""] as const,
};

// ── Types ─────────────────────────────────────────────────────────────────

export interface SearchStudentsAcrossSchoolsOpts {
  q?: string;
  schoolId?: string;
  jenjang?: string;
  grade?: string;
  enabled?: boolean;
}

export interface GrantExamAccessInput {
  exam_id: string;
  student_ids: string[];
}

export interface GrantExamAccessResponse {
  granted_count: number;
  granted_students: Array<{
    id: string;
    name: string;
    username: string;
  }>;
}

// ── Hooks ─────────────────────────────────────────────────────────────────

/**
 * Search students across all schools (super_admin only).
 * GET /admin/exam-grants/students/search
 * All filter params are optional.
 */
export function useSearchStudentsAcrossSchools(opts?: SearchStudentsAcrossSchoolsOpts) {
  const { q, schoolId, jenjang, grade, enabled } = opts ?? {};
  return useQuery({
    queryKey: examGrantKeys.search(q, schoolId, jenjang, grade),
    enabled: enabled ?? true,
    queryFn: async () => {
      const params = new URLSearchParams();
      if (q) params.set("q", q);
      if (schoolId) params.set("school_id", schoolId);
      if (jenjang) params.set("jenjang", jenjang);
      if (grade) params.set("grade", grade);
      const query = params.toString();
      const path = query ? `/admin/exam-grants/students/search?${query}` : "/admin/exam-grants/students/search";
      return authFetch<{ data: CrossSchoolStudent[]; next_cursor?: string }>(path);
    },
  });
}

/**
 * Grant free exam access to selected students (super_admin only).
 * POST /admin/exam-grants
 */
export function useGrantExamAccess() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: GrantExamAccessInput) =>
      authFetch<GrantExamAccessResponse>("/admin/exam-grants", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examGrantKeys.all });
    },
  });
}
