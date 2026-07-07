"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { AdminSchoolInput, AdminSchoolUpdateInput, School } from "@/lib/types";

export const adminSchoolsKeys = {
  all: ["admin", "schools"] as const,
  list: (cursor?: string, limit?: number) =>
    [...adminSchoolsKeys.all, "list", cursor ?? "initial", limit ?? 20] as const,
};

export function useAdminSchools(cursor?: string, limit?: number) {
  return useQuery({
    queryKey: adminSchoolsKeys.list(cursor, limit),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      const query = params.toString();
      const path = query ? `/admin/schools?${query}` : "/admin/schools";
      return authFetch<{ data: School[]; next_cursor?: string }>(path);
    },
  });
}

export function useCreateSchool() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminSchoolInput) =>
      authFetch<School>("/admin/schools", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminSchoolsKeys.all });
    },
  });
}

export function useUpdateSchool() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & AdminSchoolUpdateInput) =>
      authFetch<School>(`/admin/schools/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminSchoolsKeys.all });
    },
  });
}

export function useChangeSchoolStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      authFetch<School>(`/admin/schools/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminSchoolsKeys.all });
    },
  });
}
