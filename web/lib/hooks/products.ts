"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api";
import type { Product, ProductType } from "@/lib/types";

export const productsKeys = {
  all: ["products"] as const,
  list: (type?: ProductType) => [...productsKeys.all, "list", type ?? "all"] as const,
  detail: (id: string) => [...productsKeys.all, "detail", id] as const,
};

export function useProducts(type?: ProductType) {
  return useQuery({
    queryKey: productsKeys.list(type),
    queryFn: () => {
      const qs = type ? `?type=${encodeURIComponent(type)}` : "";
      return apiFetch<Product[]>(`/products${qs}`);
    },
  });
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: productsKeys.detail(id),
    queryFn: () => apiFetch<Product>(`/products/${encodeURIComponent(id)}`),
    enabled: Boolean(id),
  });
}