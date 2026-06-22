import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "@/lib/types";

export interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  setSession: (token: string, refreshToken: string, user: User) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      setSession: (token, refreshToken, user) => set({ token, refreshToken, user }),
      clear: () => set({ token: null, refreshToken: null, user: null }),
    }),
    { name: "abak-auth" }
  )
);