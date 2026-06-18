import { create } from "zustand";

export interface CartState {
  count: number;
  setCount: (n: number) => void;
}

export const useCartStore = create<CartState>()((set) => ({
  count: 0,
  setCount: (n) => set({ count: n }),
}));