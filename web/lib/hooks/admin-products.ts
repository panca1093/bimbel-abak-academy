"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { Product, AdminCreateProductInput, AdminUpdateProductInput } from "@/lib/types";

export const adminProductsKeys = {
  all: ["admin", "products"] as const,
  list: () => [...adminProductsKeys.all, "list"] as const,
};

export function useAdminProducts() {
  return useQuery({
    queryKey: adminProductsKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: Product[]; next_cursor?: string }>("/admin/products");
      return res.data ?? [];
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateProductInput) =>
      authFetch<Product>("/admin/products", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminProductsKeys.list() });
    },
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: AdminUpdateProductInput }) =>
      authFetch<Product>(`/admin/products/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminProductsKeys.list() });
    },
  });
}

export function usePublishProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<{ message: string }>(`/admin/products/${encodeURIComponent(id)}/publish`, {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminProductsKeys.list() });
    },
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      authFetch<void>(`/admin/products/${encodeURIComponent(id)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminProductsKeys.list() });
    },
  });
}
