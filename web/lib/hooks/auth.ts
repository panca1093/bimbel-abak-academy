"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { apiFetch, authFetch } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { LoginResponse, User } from "@/lib/types";

export interface LoginInput {
  identifier: string;
  password: string;
}

export function useLogin() {
  const setSession = useAuthStore((s) => s.setSession);
  return useMutation({
    mutationFn: (input: LoginInput) =>
      apiFetch<LoginResponse>(`/auth/login`, {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: (data) => {
      if (data.access_token && data.user) {
        setSession(data.access_token, data.user);
      }
    },
  });
}

export interface RegisterInput {
  name: string;
  email: string;
  password: string;
}

export function useRegister() {
  return useMutation({
    mutationFn: (input: RegisterInput) =>
      apiFetch<void>(`/auth/register`, {
        method: "POST",
        body: JSON.stringify(input),
      }),
  });
}

export interface VerifyOtpInput {
  identifier: string;
  code: string;
  pending_token?: string;
}

export function useVerifyOtp() {
  const setSession = useAuthStore((s) => s.setSession);
  return useMutation({
    mutationFn: (input: VerifyOtpInput) =>
      apiFetch<LoginResponse>(`/auth/otp/verify`, {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: (data) => {
      if (data.access_token && data.user) {
        setSession(data.access_token, data.user);
      }
    },
  });
}

export function useLogout() {
  const clear = useAuthStore((s) => s.clear);
  return useMutation({
    mutationFn: () => authFetch<void>(`/auth/logout`, { method: "POST" }),
    onSettled: () => {
      clear();
    },
  });
}

export function useMe(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: () => authFetch<User>(`/auth/me`),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  });
}