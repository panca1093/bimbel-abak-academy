"use client";

import Link from "next/link";
import { ArrowRight, Receipt } from "lucide-react";
import { useOrders } from "@/lib/hooks/orders";
import { useTranslation } from "@/lib/i18n";
import { formatRupiah } from "@/lib/format";
import type { Order } from "@/lib/types";
import { OrderStatusBadge } from "@/components/orders/OrderStatusBadge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";

function formatDate(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(d);
}

function OrdersList({ orders }: { orders: Order[] }) {
  const { t } = useTranslation();
  if (orders.length === 0) {
    return (
      <div className="rounded-lg border border-line bg-surface px-6 py-16 text-center">
        <Receipt className="mx-auto size-10 text-ink-300" strokeWidth={1.5} />
        <p className="mt-3 text-sm font-medium text-ink-700">{t("orders_empty")}</p>
        <p className="mt-1 text-xs text-ink-500">
          {t("orders_empty_desc")}
        </p>
        <Button asChild variant="outline" size="sm" className="mt-5">
          <Link href="/catalog">{t("orders_start_shopping")}</Link>
        </Button>
      </div>
    );
  }

  const sorted = [...orders].sort((a, b) => {
    const ta = a.created_at ? Date.parse(a.created_at) : 0;
    const tb = b.created_at ? Date.parse(b.created_at) : 0;
    return tb - ta;
  });

  return (
    <ul className="flex flex-col gap-3">
      {sorted.map((o) => (
        <li key={o.id}>
          <Link
            href={`/orders/${o.id}`}
            className="group flex items-center gap-4 rounded-lg border border-line bg-surface p-4 shadow-[var(--sh-sm)] transition-colors hover:border-ink-400"
          >
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <div className="flex items-center gap-2">
                <span className="font-mono text-xs font-semibold text-ink-500">#{o.id.slice(-8)}</span>
                <OrderStatusBadge status={o.status} />
              </div>
              <div className="flex items-center gap-3 text-sm text-ink-600">
                <span className="truncate">{formatDate(o.created_at)}</span>
                {o.items && o.items.length > 0 && (
                  <span className="truncate text-ink-400">
                    {o.items
                      .slice(0, 2)
                      .map((it) => it.name)
                      .join(", ")}
                    {o.items.length > 2 ? ` +${o.items.length - 2}` : ""}
                  </span>
                )}
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="font-serif text-base font-bold text-ink-900">
                {formatRupiah(o.total)}
              </span>
              <ArrowRight className="size-4 text-ink-400 transition-transform group-hover:translate-x-0.5" />
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}

function OrdersSkeleton() {
  return (
    <ul className="flex flex-col gap-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <li
          key={i}
          className="flex items-center gap-4 rounded-lg border border-line bg-surface p-4"
        >
          <div className="flex flex-1 flex-col gap-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-48" />
          </div>
          <Skeleton className="h-5 w-24" />
        </li>
      ))}
    </ul>
  );
}

export default function OrdersPage() {
  const { t } = useTranslation();
  const { data, isLoading, isError, error, refetch } = useOrders();

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-6 flex items-center gap-3">
        <Receipt className="size-6 text-success" />
        <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">{t("orders_title")}</h1>
      </header>

      {isError ? (
        <div className="rounded-lg border border-danger/30 bg-danger-bg px-5 py-4 text-sm text-danger">
          <p>{t("orders_load_failed")} {(error as Error)?.message}</p>
          <button onClick={() => refetch()} className="mt-2 underline">
            {t("retry")}
          </button>
        </div>
      ) : isLoading ? (
        <OrdersSkeleton />
      ) : (
        <OrdersList orders={data ?? []} />
      )}
    </div>
  );
}