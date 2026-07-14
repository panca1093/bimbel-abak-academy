import Link from "next/link";
import { Book, PlayCircle, ClipboardList } from "lucide-react";
import type { Product, ProductType } from "@/lib/types";
import { formatRupiah } from "@/lib/format";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const TYPE_META: Record<
  ProductType,
  { label: string; tone: string; bg: string; Icon: typeof Book }
> = {
  book: { label: "Buku", tone: "text-warn", bg: "bg-warn-bg", Icon: Book },
  course: { label: "Kursus", tone: "text-success", bg: "bg-success-bg", Icon: PlayCircle },
  exam: { label: "Ujian", tone: "text-info", bg: "bg-info-bg", Icon: ClipboardList },
};

const COVER_GRADIENT: Record<ProductType, string> = {
  book: "linear-gradient(135deg, #fbf1e2 0%, #f6e6cf 100%)",
  course: "linear-gradient(135deg, #e5f5ec 0%, #d4eede 100%)",
  exam: "linear-gradient(135deg, #e7eefb 0%, #d3e2f8 100%)",
};

export interface ProductCardProps {
  product: Product;
  className?: string;
}

export function ProductCard({ product, className }: ProductCardProps) {
  const meta = TYPE_META[product.type];
  const { Icon } = meta;

  return (
    <Link
      href={`/catalog/${product.id}`}
      className={cn(
        "group flex flex-col overflow-hidden rounded-lg border border-line bg-surface shadow-[var(--sh-sm)] transition-all hover:-translate-y-0.5 hover:shadow-[var(--sh-md)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
        className,
      )}
    >
      <div
        className="relative flex h-32 items-center justify-center"
        style={product.image_url ? { backgroundImage: `url(${product.image_url})`, backgroundSize: "cover", backgroundPosition: "center" } : { background: COVER_GRADIENT[product.type] }}
      >
        {!product.image_url && (
          <Icon className="size-10 text-white/90 drop-shadow-sm" strokeWidth={1.5} />
        )}
        <div className="absolute left-3 top-3">
          <Badge variant="outline" className={cn("border-transparent", meta.bg, meta.tone)}>
            {meta.label}
          </Badge>
        </div>
      </div>
      <div className="flex flex-1 flex-col gap-1 p-4">
        <div className="line-clamp-2 text-[15px] font-semibold leading-snug text-ink-900">
          {product.name}
        </div>
        {product.description && (
          <p className="line-clamp-2 text-xs leading-relaxed text-ink-500">
            {product.description}
          </p>
        )}
        <div className="mt-auto pt-3 font-serif text-lg font-bold text-success">
          {formatRupiah(product.price)}
        </div>
      </div>
    </Link>
  );
}