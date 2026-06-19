"use client";

import Link from "next/link";
import { ArrowLeft, ShoppingCart, X } from "lucide-react";
import { useCart, useRemoveCartItem, useValidatePromo } from "@/lib/hooks/orders";
import { formatRupiah } from "@/lib/format";
import type { OrderItem } from "@/lib/types";
import { CartLineItem } from "@/components/cart/CartLineItem";
import { PromoInput } from "@/components/cart/PromoInput";
import { SnapCheckout } from "@/components/cart/SnapCheckout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

export default function CartPage() {
  const { data: cart, isLoading, isError, error, refetch } = useCart();
  const removeItem = useRemoveCartItem();
  const validatePromo = useValidatePromo();

  const items: OrderItem[] = cart?.items ?? [];
  const subtotal = cart?.subtotal ?? items.reduce((s, it) => s + it.jumlah, 0);
  const discount = cart?.discount ?? 0;
  const total = cart?.total ?? Math.max(0, subtotal - discount);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10">
      <Link
        href="/catalog"
        className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-ink-500 transition-colors hover:text-ink-900"
      >
        <ArrowLeft className="size-4" /> Lanjutkan belanja
      </Link>

      <header className="mb-6 flex items-center gap-3">
        <ShoppingCart className="size-6 text-success" />
        <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">Keranjang</h1>
        {items.length > 0 && (
          <Badge variant="outline" className="border-transparent bg-success-bg text-success">
            {items.length} item
          </Badge>
        )}
      </header>

      {isLoading ? (
        <CartSkeleton />
      ) : isError ? (
        <ErrorState message={error instanceof Error ? error.message : "Gagal memuat keranjang"} onRetry={refetch} />
      ) : items.length === 0 ? (
        <EmptyCart />
      ) : (
        <div className="grid gap-6 lg:grid-cols-[1fr_360px] lg:items-start">
          <section className="flex flex-col gap-3">
            {items.map((it) => (
              <CartLineItem
                key={it.id}
                item={it}
                onRemove={() => {
                  if (!cart) return;
                  removeItem.mutate({ orderId: cart.id, itemId: it.id });
                }}
                removing={removeItem.isPending}
              />
            ))}
          </section>

          <aside className="lg:sticky lg:top-6">
            <Card className="p-5">
              <h2 className="font-serif text-lg font-semibold text-ink-900">Ringkasan Pesanan</h2>

              <PromoInput
                onValidate={(code) => validatePromo.mutate({ code, orderId: cart?.id, subtotal })}
                isValidating={validatePromo.isPending}
                discount={validatePromo.data?.discount}
                finalTotal={validatePromo.data?.final_total}
                error={validatePromo.isError ? "Kode promo tidak valid" : undefined}
              />

              <div className="mt-4 space-y-2 border-t border-line pt-4 text-sm">
                <Row label="Subtotal" value={formatRupiah(subtotal)} />
                {discount > 0 && <Row label="Diskon" value={`−${formatRupiah(discount)}`} tone="text-success" />}
              </div>

              <div className="mt-4 flex items-center justify-between border-t border-line pt-4">
                <span className="font-semibold text-ink-900">Total</span>
                <span className="font-serif text-2xl font-bold text-success">{formatRupiah(total)}</span>
              </div>

              <SnapCheckout orderId={cart?.id} />

              <p className="mt-3 text-center text-xs text-ink-400">
                Midtrans · pembayaran aman terenkripsi
              </p>
            </Card>
          </aside>
        </div>
      )}
    </div>
  );
}

function Row({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-ink-500">{label}</span>
      <span className={`font-semibold ${tone ?? "text-ink-900"}`}>{value}</span>
    </div>
  );
}

function CartSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="h-24 w-full rounded-lg" />
      ))}
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <Card className="flex flex-col items-center gap-3 p-10 text-center">
      <X className="size-8 text-danger" />
      <p className="text-sm text-ink-600">{message}</p>
      <Button variant="outline" size="sm" onClick={onRetry}>Coba lagi</Button>
    </Card>
  );
}

function EmptyCart() {
  return (
    <Card className="flex flex-col items-center gap-4 p-12 text-center">
      <div className="flex size-16 items-center justify-center rounded-full bg-paper">
        <ShoppingCart className="size-7 text-ink-400" />
      </div>
      <div>
        <h2 className="font-serif text-lg font-semibold text-ink-900">Keranjang masih kosong</h2>
        <p className="mt-1 text-sm text-ink-500">Yuk jelajahi katalog dan tambahkan buku atau kursus favoritmu.</p>
      </div>
      <Button asChild>
        <Link href="/catalog">Lihat Katalog</Link>
      </Button>
    </Card>
  );
}