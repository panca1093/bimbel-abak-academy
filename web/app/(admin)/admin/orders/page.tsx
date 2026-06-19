"use client";

import { useMemo, useState } from "react";
import { toast } from "sonner";
import {
  useAdminOrders,
  useConfirmOrder,
  useShipOrder,
  useRefundOrder,
  useReconcileOrder,
} from "@/lib/hooks/admin-orders";
import { OrderStatusBadge } from "@/components/orders/OrderStatusBadge";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { formatRupiah } from "@/lib/format";
import type { Order, OrderStatus, AdminOrderFilterStatus } from "@/lib/types";

const FILTER_OPTIONS: AdminOrderFilterStatus[] = ["all", "pending", "paid", "failed", "refunded"];

const FILTER_LABELS: Record<AdminOrderFilterStatus, string> = {
  all: "All",
  pending: "Pending",
  paid: "Paid",
  failed: "Failed",
  refunded: "Refunded",
};

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan.";
}

function orderNumber(order: Order): string {
  return `#${order.id.slice(-8)}`;
}

function productSummary(order: Order): string {
  if (!order.items || order.items.length === 0) return "-";
  const names = order.items.slice(0, 2).map((it) => it.name);
  const suffix = order.items.length > 2 ? ` +${order.items.length - 2}` : "";
  return names.join(", ") + suffix;
}

function hasBookItem(order: Order): boolean {
  return (order.items ?? []).some((it) => it.product_type === "book");
}

function isShipped(order: Order): boolean {
  return Boolean(order.tracking_number || order.shipped_at);
}

function shippingBadge(order: Order) {
  if (!hasBookItem(order)) return null;
  if (isShipped(order)) {
    return <Badge className="bg-green-100 text-green-800 border-green-200">Shipped</Badge>;
  }
  return <Badge variant="outline">Pending</Badge>;
}

function actionAllowed(status: OrderStatus, action: "confirm" | "ship" | "refund" | "reconcile"): boolean {
  switch (action) {
    case "confirm":
      return status === "payment_pending";
    case "ship":
      return status === "paid" || status === "processing";
    case "refund":
      return status === "paid" || status === "processing" || status === "completed";
    case "reconcile":
      return status === "payment_pending";
  }
}

export default function OrdersPage() {
  const [filter, setFilter] = useState<AdminOrderFilterStatus>("all");
  const { data: orders, isLoading, isError, error } = useAdminOrders(filter);
  const confirm = useConfirmOrder();
  const ship = useShipOrder();
  const refund = useRefundOrder();
  const reconcile = useReconcileOrder();

  const filtered = useMemo(() => {
    if (!orders) return [];
    return orders;
  }, [orders]);

  async function handleConfirm(id: string) {
    try {
      await confirm.mutateAsync(id);
      toast.success("Pesanan dikonfirmasi.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleShip(id: string) {
    const trackingNumber = window.prompt("Masukkan nomor resi pengiriman:");
    if (!trackingNumber) return;
    try {
      await ship.mutateAsync({ id, trackingNumber });
      toast.success("Pesanan dikirim.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleRefund(id: string) {
    if (!window.confirm("Yakin ingin mengembalikan dana pesanan ini?")) return;
    try {
      await refund.mutateAsync(id);
      toast.success("Pesanan direfund.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleReconcile(id: string) {
    try {
      await reconcile.mutateAsync(id);
      toast.success("Status pembayaran direkonsiliasi.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Orders</h1>
      </div>

      <div className="flex flex-wrap gap-2">
        {FILTER_OPTIONS.map((f) => (
          <Button
            key={f}
            variant={filter === f ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter(f)}
          >
            {FILTER_LABELS[f]}
          </Button>
        ))}
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          Gagal memuat pesanan: {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="overflow-x-auto rounded-lg border">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">Order</th>
                <th className="px-4 py-3 text-left font-medium">Buyer</th>
                <th className="px-4 py-3 text-left font-medium">Product</th>
                <th className="px-4 py-3 text-left font-medium">Amount</th>
                <th className="px-4 py-3 text-left font-medium">Payment</th>
                <th className="px-4 py-3 text-left font-medium">Shipping</th>
                <th className="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((order) => (
                <tr key={order.id} className="border-t">
                  <td className="px-4 py-3 font-mono font-medium">{orderNumber(order)}</td>
                  <td className="px-4 py-3">{order.student_id}</td>
                  <td className="px-4 py-3 max-w-xs truncate">{productSummary(order)}</td>
                  <td className="px-4 py-3">{formatRupiah(order.total)}</td>
                  <td className="px-4 py-3">
                    <OrderStatusBadge status={order.status} />
                  </td>
                  <td className="px-4 py-3">{shippingBadge(order)}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {actionAllowed(order.status, "confirm") && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleConfirm(order.id)}
                          disabled={confirm.isPending}
                        >
                          Confirm
                        </Button>
                      )}
                      {actionAllowed(order.status, "ship") && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleShip(order.id)}
                          disabled={ship.isPending}
                        >
                          Ship
                        </Button>
                      )}
                      {actionAllowed(order.status, "refund") && (
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleRefund(order.id)}
                          disabled={refund.isPending}
                        >
                          Refund
                        </Button>
                      )}
                      {actionAllowed(order.status, "reconcile") && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleReconcile(order.id)}
                          disabled={reconcile.isPending}
                        >
                          Reconcile
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">
                    Tidak ada pesanan.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
