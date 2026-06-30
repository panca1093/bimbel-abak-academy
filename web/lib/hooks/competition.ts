"use client";

import { useQuery } from "@tanstack/react-query";
import { API_BASE, authFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { RegistrationDetail, RegistrationListItem } from "@/lib/types";

export const competitionKeys = {
  all: ["competition"] as const,
  lists: () => [...competitionKeys.all, "list"] as const,
  list: () => [...competitionKeys.lists()] as const,
  details: () => [...competitionKeys.all, "detail"] as const,
  detail: (id: string) => [...competitionKeys.details(), id] as const,
};

export function useRegistrations() {
  return useQuery({
    queryKey: competitionKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: RegistrationListItem[] }>(
        "/competition/registrations"
      );
      return res.data ?? [];
    },
  });
}

export function useRegistration(id: string | undefined) {
  return useQuery({
    queryKey: competitionKeys.detail(id ?? ""),
    queryFn: () =>
      authFetch<RegistrationDetail>(
        `/competition/registrations/${encodeURIComponent(id!)}`
      ),
    enabled: Boolean(id),
  });
}

export async function downloadCard(id: string): Promise<void> {
  const token = useAuthStore.getState().token;
  const res = await fetch(
    `${API_BASE}/competition/registrations/${encodeURIComponent(id)}/card`,
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