import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useShippingRates, usePatchCart, ordersKeys } from "./orders";
import type { CourierRate } from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
  ApiError: class extends Error {
    code: string;
    status: number;
    constructor(code: string, message: string, status: number) {
      super(message);
      this.code = code;
      this.status = status;
    }
  },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: {
    getState: () => ({ token: "test-token" }),
  },
}));

const sampleRates: CourierRate[] = [
  { courier: "JNE", service: "REG", estimated_days: 3, price: 15000 },
  { courier: "TIKI", service: "ONS", estimated_days: 1, price: 25000 },
];

describe("orders hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  describe("useShippingRates", () => {
    it("posts to /orders/shipping with destination_postal_code and weight_grams", async () => {
      mockAuthFetch.mockResolvedValueOnce({ rates: sampleRates });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useShippingRates(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          destination_postal_code: "12345",
          weight_grams: 500,
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/orders/shipping", {
        method: "POST",
        body: JSON.stringify({
          destination_postal_code: "12345",
          weight_grams: 500,
        }),
      });
    });

    it("returns CourierRate[] from the response", async () => {
      mockAuthFetch.mockResolvedValueOnce({ rates: sampleRates });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useShippingRates(), { wrapper });

      let data: CourierRate[] | undefined;
      await act(async () => {
        data = await result.current.mutateAsync({
          destination_postal_code: "12345",
          weight_grams: 500,
        });
      });

      expect(data).toEqual(sampleRates);
    });
  });

  describe("usePatchCart", () => {
    it("patches /orders/:id with courier, shipping_cost, and address fields", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "order updated" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => usePatchCart(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          orderId: "o1",
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: "12345",
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/orders/o1", {
        method: "PATCH",
        body: JSON.stringify({
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: "12345",
        }),
      });
    });

    it("invalidates cart query on success", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "order updated" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => usePatchCart(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          orderId: "o1",
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: "12345",
        });
      });

      expect(spy).toHaveBeenCalledWith({ queryKey: ordersKeys.cart() });
    });

    it("supports nullable kode_pos field", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "order updated" });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => usePatchCart(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          orderId: "o1",
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: null,
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/orders/o1", {
        method: "PATCH",
        body: JSON.stringify({
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: null,
        }),
      });
    });

    it("can include optional promo_code field", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "order updated" });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => usePatchCart(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          orderId: "o1",
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: "12345",
          promo_code: "SAVE10",
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/orders/o1", {
        method: "PATCH",
        body: JSON.stringify({
          courier: "JNE",
          service: "REG",
          shipping_cost: 15000,
          province_id: "p1",
          city_id: "c1",
          district_id: "d1",
          kode_pos: "12345",
          promo_code: "SAVE10",
        }),
      });
    });
  });
});

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
    queryClient,
  };
}
