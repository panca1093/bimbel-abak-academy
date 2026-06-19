"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { PromoCode, AdminCreatePromoCodeInput, AdminUpdatePromoCodeInput } from "@/lib/types";

export const adminPromosKeys = {
  all: ["admin", "promos"] as const,
  list: () => [...adminPromosKeys.all, "list"] as const,
};

export function useAdminPromoCodes() {
  return useQuery({
    queryKey: adminPromosKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: PromoCode[] }>("/admin/promo-codes");
      return res.data ?? [];
    },
  });
}

export function useCreatePromoCode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreatePromoCodeInput) =>
      authFetch<PromoCode>("/admin/promo-codes", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminPromosKeys.list() });
    },
  });
}

export function useUpdatePromoCode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: AdminUpdatePromoCodeInput }) =>
      authFetch<{ message: string }>(`/admin/promo-codes/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminPromosKeys.list() });
    },
  });
}

export function useDeletePromoCode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<void>(`/admin/promo-codes/${encodeURIComponent(id)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminPromosKeys.list() });
    },
  });
}
