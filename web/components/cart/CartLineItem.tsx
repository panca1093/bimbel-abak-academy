"use client";

import { Book, PlayCircle, Trash2, Trophy } from "lucide-react";
import type { OrderItem } from "@/lib/types";
import { formatRupiah } from "@/lib/format";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const TYPE_META: Record<string, { label: string; tone: string; bg: string; Icon: typeof Book }> = {
  book: { label: "Buku", tone: "text-warn", bg: "bg-warn-bg", Icon: Book },
  course: { label: "Kursus", tone: "text-success", bg: "bg-success-bg", Icon: PlayCircle },
  package: { label: "Kompetisi", tone: "text-violet", bg: "bg-violet-bg", Icon: Trophy },
};

export interface CartLineItemProps {
  item: OrderItem;
  onRemove: () => void;
  removing?: boolean;
}

export function CartLineItem({ item, onRemove, removing }: CartLineItemProps) {
  const meta = TYPE_META[item.product_type] ?? TYPE_META.book;
  const { Icon } = meta;
  const isBook = item.product_type === "book";
  const lineTotal = item.jumlah ?? item.unit_price * item.qty;

  return (
    <div className="flex gap-4 rounded-lg border border-line bg-surface p-4 shadow-[var(--sh-sm)]">
      <div
        className="flex size-16 shrink-0 items-center justify-center rounded-md bg-paper"
        aria-hidden
      >
        <Icon className="size-7 text-ink-400" strokeWidth={1.5} />
      </div>

      <div className="flex flex-1 flex-col gap-1">
        <div className="flex items-start justify-between gap-2">
          <div className="flex flex-col gap-1">
            <span className="text-[15px] font-semibold leading-snug text-ink-900">{item.name}</span>
            <Badge variant="outline" className={cn("w-fit border-transparent", meta.bg, meta.tone)}>
              {meta.label}
            </Badge>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 text-ink-400 hover:text-danger"
            onClick={onRemove}
            disabled={removing}
            aria-label={`Hapus ${item.name} dari keranjang`}
          >
            <Trash2 className="size-4" />
          </Button>
        </div>

        <div className="mt-1 flex items-end justify-between gap-2 text-sm">
          <div className="flex flex-col text-ink-500">
            <span>{formatRupiah(item.unit_price)}</span>
            {isBook && item.qty > 1 && <span className="text-xs">Qty: {item.qty}</span>}
          </div>
          <span className="font-serif text-base font-bold text-ink-900">{formatRupiah(lineTotal)}</span>
        </div>
      </div>
    </div>
  );
}