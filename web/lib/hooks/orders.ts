"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { CheckoutResult, Order, PromoValidation } from "@/lib/types";

export const ordersKeys = {
  all: ["orders"] as const,
  list: () => [...ordersKeys.all, "list"] as const,
  cart: () => [...ordersKeys.all, "cart"] as const,
  detail: (id: string) => [...ordersKeys.all, "detail", id] as const,
};

export function useOrders() {
  return useQuery({
    queryKey: ordersKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: Order[]; next_cursor?: string }>(`/orders`);
      return res.data ?? [];
    },
  });
}

export function useOrder(id: string) {
  return useQuery({
    queryKey: ordersKeys.detail(id),
    queryFn: () => authFetch<Order>(`/orders/${encodeURIComponent(id)}`),
    enabled: Boolean(id),
  });
}

export function useCart() {
  return useQuery({
    queryKey: ordersKeys.cart(),
    queryFn: () => authFetch<Order>(`/orders`, { method: "POST" }),
  });
}

interface AddToCartInput {
  productId: string;
  qty?: number;
  cartId?: string;
}

interface AddItemBody {
  product_id: string;
  qty: number;
}

export function useAddToCart() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ productId, qty = 1, cartId }: AddToCartInput) => {
      let orderId = cartId;
      if (!orderId) {
        const cart = await authFetch<Order>("/orders", {
          method: "POST",
          body: JSON.stringify({ status: "cart" }),
        });
        orderId = cart.id;
      }
      return authFetch<Order>(`/orders/${encodeURIComponent(orderId!)}/items`, {
        method: "POST",
        body: JSON.stringify({ product_id: productId, qty } satisfies AddItemBody),
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ordersKeys.cart() });
      qc.invalidateQueries({ queryKey: ordersKeys.list() });
    },
  });
}

interface RemoveCartItemInput {
  orderId: string;
  itemId: string;
}

export function useRemoveCartItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ orderId, itemId }: RemoveCartItemInput) =>
      authFetch<void>(`/orders/${encodeURIComponent(orderId)}/items/${encodeURIComponent(itemId)}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ordersKeys.cart() });
      qc.invalidateQueries({ queryKey: ordersKeys.list() });
    },
  });
}

interface UpdateCartItemQtyInput {
  orderId: string;
  itemId: string;
  qty: number;
}

export function useUpdateCartItemQty() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ orderId, itemId, qty }: UpdateCartItemQtyInput) =>
      authFetch<void>(`/orders/${encodeURIComponent(orderId)}/items/${encodeURIComponent(itemId)}`, {
        method: "PATCH",
        body: JSON.stringify({ qty }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ordersKeys.cart() });
    },
  });
}

interface ValidatePromoInput {
  code: string;
  orderId?: string;
  subtotal?: number;
}

export function useValidatePromo() {
  return useMutation({
    mutationFn: (input: ValidatePromoInput) =>
      authFetch<PromoValidation>(`/promo-codes/validate`, {
        method: "POST",
        body: JSON.stringify(input),
      }),
  });
}

export function useCheckout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderId: string) =>
      authFetch<CheckoutResult>(`/orders/${encodeURIComponent(orderId)}/checkout`, {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ordersKeys.cart() });
      qc.invalidateQueries({ queryKey: ordersKeys.list() });
    },
  });
}

export function useRetryPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderId: string) =>
      authFetch<CheckoutResult>(`/orders/${encodeURIComponent(orderId)}/retry`, {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
      }),
    onSuccess: (data, orderId) => {
      qc.invalidateQueries({ queryKey: ordersKeys.detail(orderId) });
      qc.invalidateQueries({ queryKey: ordersKeys.list() });
    },
  });
}