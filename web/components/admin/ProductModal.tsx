"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { Product, ProductType, ProductStatus, AdminCreateProductInput, AdminUpdateProductInput } from "@/lib/types";

interface ProductModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  product?: Product | null;
  onSubmit: (input: AdminCreateProductInput | AdminUpdateProductInput) => void;
  isPending: boolean;
}

const PRODUCT_TYPES: ProductType[] = ["book", "course", "package"];
const PRODUCT_STATUSES: ProductStatus[] = ["draft", "published", "hidden", "archived"];

const TYPE_LABELS: Record<ProductType, string> = {
  book: "Buku",
  course: "Kursus",
  package: "Paket",
};

const STATUS_LABELS: Record<ProductStatus, string> = {
  draft: "Draft",
  published: "Dipublikasikan",
  hidden: "Disembunyikan",
  archived: "Diarsipkan",
};

export function ProductModal({ open, onOpenChange, product, onSubmit, isPending }: ProductModalProps) {
  const isEdit = Boolean(product);
  const [name, setName] = useState("");
  const [type, setType] = useState<ProductType | "">("");
  const [price, setPrice] = useState("");
  const [stock, setStock] = useState("");
  const [status, setStatus] = useState<ProductStatus>("draft");
  const [description, setDescription] = useState("");

  useEffect(() => {
    if (open) {
      if (product) {
        setName(product.name ?? "");
        setType(product.type ?? "");
        setPrice(String(product.price ?? ""));
        setStock(product.stock != null ? String(product.stock) : "");
        setStatus(product.status ?? "draft");
        setDescription(product.description ?? "");
      } else {
        setName("");
        setType("");
        setPrice("");
        setStock("");
        setStatus("draft");
        setDescription("");
      }
    }
  }, [open, product]);

  const showStock = type === "book";
  const canSubmit = name.trim() !== "" && (isEdit || type !== "") && price !== "";

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit || isPending) return;

    const base = {
      name: name.trim(),
      description: description.trim() || undefined,
      price: Number(price),
    };

    if (isEdit) {
      const input: AdminUpdateProductInput = {
        ...base,
        status,
        ...(showStock ? { stock: Number(stock) } : {}),
      };
      onSubmit(input);
      return;
    }

    if (!type) return;
    const input: AdminCreateProductInput = {
      ...base,
      type,
      ...(showStock ? { stock: Number(stock) } : {}),
    };
    onSubmit(input);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{isEdit ? "Edit produk" : "Buat produk"}</DialogTitle>
            <DialogDescription>
              {isEdit ? "Perbarui metadata produk." : "Tambahkan produk baru ke katalog."}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="product-name">Nama</Label>
              <Input
                id="product-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Nama produk"
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="product-type">Jenis</Label>
                <select
                  id="product-type"
                  value={type}
                  onChange={(e) => setType(e.target.value as ProductType)}
                  disabled={isEdit || isPending}
                  className="h-9 w-full rounded-md border border-input bg-transparent px-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:opacity-50"
                >
                  <option value="" disabled>Pilih jenis</option>
                  {PRODUCT_TYPES.map((t) => (
                    <option key={t} value={t}>
                      {TYPE_LABELS[t]}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="product-price">Harga (IDR)</Label>
                <Input
                  id="product-price"
                  type="number"
                  min={0}
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  placeholder="0"
                  disabled={isPending}
                />
              </div>
            </div>

            {isEdit && (
              <div className="grid gap-2">
                <Label htmlFor="product-status">Status</Label>
                <select
                  id="product-status"
                  value={status}
                  onChange={(e) => setStatus(e.target.value as ProductStatus)}
                  disabled={isPending}
                  className="h-9 w-full rounded-md border border-input bg-transparent px-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:opacity-50"
                >
                  {PRODUCT_STATUSES.map((s) => (
                    <option key={s} value={s}>
                      {STATUS_LABELS[s]}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {showStock && (
              <div className="grid gap-2">
                <Label htmlFor="product-stock">Stok</Label>
                <Input
                  id="product-stock"
                  type="number"
                  min={0}
                  value={stock}
                  onChange={(e) => setStock(e.target.value)}
                  placeholder="0"
                  disabled={isPending}
                />
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="product-description">Deskripsi</Label>
              <textarea
                id="product-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Deskripsi singkat"
                disabled={isPending}
                rows={3}
                className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:opacity-50"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Batal
            </Button>
            <Button type="submit" disabled={!canSubmit || isPending}>
              {isPending ? "Menyimpan..." : "Simpan"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
