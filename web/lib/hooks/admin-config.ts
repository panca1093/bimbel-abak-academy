"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { SystemConfig } from "@/lib/types";

export const adminConfigKeys = {
  all: ["admin", "config"] as const,
  detail: () => [...adminConfigKeys.all, "detail"] as const,
};

export function useAdminSystemConfig() {
  return useQuery({
    queryKey: adminConfigKeys.detail(),
    queryFn: () => authFetch<SystemConfig>("/admin/system/config"),
  });
}

export function useUpdateSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (values: Record<string, string>) =>
      authFetch<SystemConfig>("/admin/system/config", {
        method: "PUT",
        body: JSON.stringify(values),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminConfigKeys.all });
    },
  });
}
