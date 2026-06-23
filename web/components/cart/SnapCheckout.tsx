"use client";

import { Lock, Loader2, ExternalLink } from "lucide-react";
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

  const handleCheckout = () => {
    if (!orderId) return;
    checkout.mutate(orderId, {
      onSuccess: (data) => {
        if (data.payment_url) {
          window.open(data.payment_url, "_blank");
          toast.info("Selesaikan pembayaran di tab baru, lalu refresh halaman ini.");
        } else if (data.snap_token) {
          navigator.clipboard.writeText(data.snap_token);
          toast.info("Snap token disalin. Buka Midtrans untuk melanjutkan pembayaran.");
        } else {
          toast.error("Gagal memulai pembayaran. Coba lagi nanti.");
        }
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
      disabled={!orderId || checkout.isPending || disabled}
    >
      {checkout.isPending ? (
        <Loader2 className="size-4 animate-spin" />
      ) : (
        <ExternalLink className="size-4" />
      )}
      Bayar di Tab Baru
    </Button>
  );
}
