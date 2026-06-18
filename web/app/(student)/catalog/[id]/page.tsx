"use client";

import { use } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Book, ShoppingCart, PlayCircle, Trophy } from "lucide-react";
import { toast } from "sonner";

import { useProduct } from "@/lib/hooks/products";
import { useAddToCart } from "@/lib/hooks/orders";
import { useCartStore } from "@/stores/cart";
import { useAuthStore } from "@/stores/auth";
import { formatRupiah } from "@/lib/format";
import { ApiError } from "@/lib/api";
import type { ProductType } from "@/lib/types";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const TYPE_META: Record<
  ProductType,
  { label: string; tone: string; bg: string; Icon: typeof Book }
> = {
  book: { label: "Buku", tone: "text-warn", bg: "bg-warn-bg", Icon: Book },
  course: { label: "Kursus", tone: "text-success", bg: "bg-success-bg", Icon: PlayCircle },
  package: { label: "Kompetisi", tone: "text-violet", bg: "bg-violet-bg", Icon: Trophy },
};

const COVER_GRADIENT: Record<ProductType, string> = {
  book: "linear-gradient(135deg, #fbf1e2 0%, #f6e6cf 100%)",
  course: "linear-gradient(135deg, #e5f5ec 0%, #d4eede 100%)",
  package: "linear-gradient(135deg, #efe9fb 0%, #e0d4f7 100%)",
};

export default function ProductDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const router = useRouter();
  const { data: product, isLoading, isError, error, refetch } = useProduct(id);
  const addToCart = useAddToCart();
  const bumpBadge = useCartStore((s) => s.setCount);
  const token = useAuthStore((s) => s.token);

  if (isLoading) return <DetailSkeleton />;

  if (isError || !product) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-16 md:px-6">
        <div className="rounded-lg border border-danger/30 bg-danger-bg px-5 py-4 text-sm text-danger">
          <p>Gagal memuat produk. {(error as Error)?.message}</p>
          <button onClick={() => refetch()} className="mt-2 underline">
            Coba lagi
          </button>
        </div>
      </div>
    );
  }

  const meta = TYPE_META[product.type];
  const { Icon } = meta;

  const handleAdd = (thenRoute?: () => void) => {
    if (!token) {
      router.push("/login");
      return;
    }
    addToCart.mutate(
      { productId: product.id, qty: 1 },
      {
        onSuccess: () => {
          bumpBadge(useCartStore.getState().count + 1);
          toast.success("Ditambahkan ke keranjang", {
            description: product.name,
          });
          thenRoute?.();
        },
        onError: (err) => {
          const msg = err instanceof ApiError ? err.message : "Gagal menambahkan ke keranjang.";
          toast.error("Gagal menambahkan", { description: msg });
        },
      },
    );
  };

  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 md:py-10">
      <Button asChild variant="ghost" size="sm" className="mb-4">
        <Link href="/catalog">
          <ArrowLeft className="size-4" />
          Katalog
        </Link>
      </Button>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-[1fr_340px] md:gap-8">
        <div className="flex flex-col gap-6">
          <div
            className="flex h-64 items-center justify-center overflow-hidden rounded-lg border border-line md:h-72"
            style={
              product.image_url
                ? { backgroundImage: `url(${product.image_url})`, backgroundSize: "cover", backgroundPosition: "center" }
                : { background: COVER_GRADIENT[product.type], border: 0 }
            }
          >
            {!product.image_url && (
              <Icon className="size-16 text-white/90 drop-shadow-sm" strokeWidth={1.5} />
            )}
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-3">
              <Badge variant="outline" className={cn("border-transparent", meta.bg, meta.tone)}>
                {meta.label}
              </Badge>
              <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">
                {product.name}
              </h1>
            </div>
            <p className="max-w-2xl text-sm leading-relaxed text-ink-600 md:text-[15px]">
              {product.description ?? "Tidak ada deskripsi."}
            </p>
            {product.type === "book" && (
              <p className="text-xs text-ink-500">
                Stok: {product.stock ?? 0} · dikirim ke alamat Anda
              </p>
            )}
          </div>
        </div>

        <aside className="md:sticky md:top-6 md:self-start">
          <div className="rounded-lg border border-line bg-surface p-5 shadow-[var(--sh-sm)]">
            <div className="font-serif text-3xl font-bold text-success">
              {formatRupiah(product.price)}
            </div>
            {product.type === "book" && (
              <div className="mt-1 text-xs text-ink-500">
                Stok: {product.stock ?? 0}
              </div>
            )}
            <div className="my-4 h-px bg-line" />
            <div className="flex flex-col gap-3">
              <Button
                size="lg"
                className="w-full"
                disabled={addToCart.isPending}
                onClick={() => handleAdd()}
              >
                <ShoppingCart className="size-4" />
                Tambah ke Keranjang
              </Button>
              <Button
                size="lg"
                variant="secondary"
                className="w-full"
                disabled={addToCart.isPending}
                onClick={() => handleAdd(() => router.push("/cart"))}
              >
                Beli Sekarang
              </Button>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}

function DetailSkeleton() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 md:py-10">
      <Skeleton className="mb-4 h-8 w-24" />
      <div className="grid grid-cols-1 gap-6 md:grid-cols-[1fr_340px] md:gap-8">
        <div className="flex flex-col gap-6">
          <Skeleton className="h-72 w-full rounded-lg" />
          <div className="flex flex-col gap-3">
            <Skeleton className="h-6 w-2/3" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </div>
        </div>
        <Skeleton className="h-56 w-full rounded-lg" />
      </div>
    </div>
  );
}