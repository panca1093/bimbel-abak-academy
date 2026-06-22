"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { AdminAccount, AdminAccountRole, AdminAccountStatus, AdminCreateAccountInput } from "@/lib/types";

export const adminAccountsKeys = {
  all: ["admin", "accounts"] as const,
  list: (role?: AdminAccountRole, status?: AdminAccountStatus) => {
    const filters: string[] = [];
    if (role) filters.push(`role=${role}`);
    if (status) filters.push(`status=${status}`);
    const key = filters.length ? filters.join("&") : "all";
    return [...adminAccountsKeys.all, "list", key] as const;
  },
};

export function useAdminAccounts(role?: AdminAccountRole, status?: AdminAccountStatus) {
  return useQuery({
    queryKey: adminAccountsKeys.list(role, status),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (role) params.set("role", role);
      if (status) params.set("status", status);
      const query = params.toString();
      const path = query ? `/admin/system/accounts?${query}` : "/admin/system/accounts";
      const res = await authFetch<{ data: AdminAccount[]; next_cursor?: string }>(path);
      return res.data ?? [];
    },
  });
}

export function useCreateAdminAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateAccountInput) =>
      authFetch<AdminAccount>("/admin/system/accounts", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAccountsKeys.all });
    },
  });
}

export function useChangeAccountRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: AdminAccountRole }) =>
      authFetch<{ message: string }>(`/admin/system/accounts/${encodeURIComponent(id)}/role`, {
        method: "PATCH",
        body: JSON.stringify({ role }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAccountsKeys.all });
    },
  });
}

export function useChangeAccountStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: AdminAccountStatus }) =>
      authFetch<{ message: string }>(`/admin/system/accounts/${encodeURIComponent(id)}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAccountsKeys.all });
    },
  });
}

export function useResetAccountPassword() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<{ message: string }>(`/admin/system/accounts/${encodeURIComponent(id)}/reset-password`, {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminAccountsKeys.all });
    },
  });
}
