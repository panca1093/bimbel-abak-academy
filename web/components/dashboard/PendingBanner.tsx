"use client";

import Link from "next/link";
import { AlertCircle } from "lucide-react";
import { formatRupiah } from "@/lib/format";
import { Button } from "@/components/ui/button";

export interface PendingBannerProps {
  id: string;
  product?: string;
  amount: number;
}

export function PendingBanner({ id, product, amount }: PendingBannerProps) {
  return (
    <div
      className="mb-8 flex items-center gap-4 rounded-lg border px-5 py-4"
      style={{
        borderColor: "var(--color-warn)",
        background: "var(--color-warn-bg)",
      }}
    >
      <AlertCircle className="size-5 shrink-0 text-warn" />
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-warn">Pembayaran tertunda</p>
        <p className="truncate text-sm text-ink-600">
          {product ? `${product} · ` : ""}
          {formatRupiah(amount)}
        </p>
      </div>
      <Button asChild size="sm">
        <Link href={`/orders/${id}`}>Bayar Sekarang</Link>
      </Button>
    </div>
  );
}