"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { Dashboard, User } from "@/lib/types";

export const studentsKeys = {
  all: ["students"] as const,
  dashboard: () => [...studentsKeys.all, "dashboard"] as const,
  profile: () => [...studentsKeys.all, "profile"] as const,
};

export function useDashboard() {
  return useQuery({
    queryKey: studentsKeys.dashboard(),
    queryFn: () => authFetch<Dashboard>(`/students/dashboard`),
  });
}

export function useProfile() {
  return useQuery({
    queryKey: studentsKeys.profile(),
    queryFn: () => authFetch<User>(`/students/profile`),
  });
}

export interface UpdateProfileInput {
  name?: string;
  email?: string;
  username?: string;
}

export function useUpdateProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateProfileInput) =>
      authFetch<User>(`/students/profile`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: studentsKeys.profile() });
    },
  });
}

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (input: ChangePasswordInput) =>
      authFetch<void>(`/auth/password/change`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
  });
}