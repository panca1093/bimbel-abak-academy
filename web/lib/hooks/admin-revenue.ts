"use client";

import { useQuery } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { AdminRevenue } from "@/lib/types";

export const adminRevenueKeys = {
  all: ["admin", "revenue"] as const,
  list: () => [...adminRevenueKeys.all, "list"] as const,
};

export function useAdminRevenue() {
  return useQuery({
    queryKey: adminRevenueKeys.list(),
    queryFn: async () => authFetch<AdminRevenue>("/admin/revenue"),
  });
}
