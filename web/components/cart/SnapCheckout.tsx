"use client";

import { Lock, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useCheckout, ordersKeys } from "@/lib/hooks/orders";
import { Button } from "@/components/ui/button";

export interface SnapCheckoutProps {
  orderId?: string;
  disabled?: boolean;
}

export function SnapCheckout({ orderId, disabled }: SnapCheckoutProps) {
  const router = useRouter();
  const qc = useQueryClient();
  const checkout = useCheckout();

  const isPending = checkout.isPending;

  const handleCheckout = () => {
    if (!orderId) return;
    checkout.mutate(orderId, {
      onSuccess: (data) => {
        if (typeof window === "undefined" || !window.snap) {
          toast.error("Payment gateway belum siap. Muat ulang halaman lalu coba lagi.");
          return;
        }
        window.snap.pay(data.snap_token, {
          onSuccess: () => {
            qc.invalidateQueries({ queryKey: ordersKeys.cart() });
            qc.invalidateQueries({ queryKey: ordersKeys.list() });
            router.push(`/orders/${data.order_id}`);
          },
          onPending: () => {
            qc.invalidateQueries({ queryKey: ordersKeys.cart() });
            qc.invalidateQueries({ queryKey: ordersKeys.list() });
            router.push(`/orders/${data.order_id}`);
          },
          onError: () => {
            toast.error("Pembayaran gagal. Coba metode lain atau ulangi sebentar lagi.");
          },
          onClose: () => {
            toast.info("Kamu menutup pembayaran. Selesaikan pembayaran agar pesanan diproses.");
          },
        });
      },
      onError: (err) => {
        toast.error(err instanceof Error ? err.message : "Gagal memulai checkout.");
      },
    });
  };

  return (
    <Button
      type="button"
      className="mt-4 w-full"
      size="lg"
      onClick={handleCheckout}
      disabled={!orderId || isPending || disabled}
    >
      {isPending ? <Loader2 className="size-4 animate-spin" /> : <Lock className="size-4" />}
      Bayar Sekarang
    </Button>
  );
}