import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "@/lib/types";

export interface AuthState {
  token: string | null;
  user: User | null;
  setSession: (token: string, user: User) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setSession: (token, user) => set({ token, user }),
      clear: () => set({ token: null, user: null }),
    }),
    { name: "abak-auth" }
  )
);