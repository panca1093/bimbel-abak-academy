"use client";

import { useState } from "react";
import { toast } from "sonner";
import {
  useAdminPromoCodes,
  useCreatePromoCode,
  useUpdatePromoCode,
  useDeletePromoCode,
} from "@/lib/hooks/admin-promos";
import { PromoModal } from "@/components/admin/PromoModal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { formatRupiah } from "@/lib/format";
import type { PromoCode, AdminCreatePromoCodeInput, AdminUpdatePromoCodeInput } from "@/lib/types";

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan.";
}

function discountLabel(promo: PromoCode): string {
  if (promo.discount_percent != null) return `${promo.discount_percent}%`;
  if (promo.discount_amount != null) return formatRupiah(promo.discount_amount);
  return "-";
}

function usageText(promo: PromoCode): string {
  const max = promo.max_uses != null ? String(promo.max_uses) : "∞";
  return `${promo.used_count} / ${max}`;
}

function expiryText(iso?: string): string {
  if (!iso) return "No expiry";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString("id-ID");
}

export default function PromosPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editingPromo, setEditingPromo] = useState<PromoCode | null>(null);

  const { data: promos, isLoading, isError, error } = useAdminPromoCodes();
  const create = useCreatePromoCode();
  const update = useUpdatePromoCode();
  const remove = useDeletePromoCode();

  function openCreate() {
    setEditingPromo(null);
    setModalOpen(true);
  }

  function openEdit(promo: PromoCode) {
    setEditingPromo(promo);
    setModalOpen(true);
  }

  async function handleCreate(input: AdminCreatePromoCodeInput) {
    try {
      await create.mutateAsync(input);
      toast.success("Kode promo dibuat.");
      setModalOpen(false);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleUpdate(input: AdminUpdatePromoCodeInput) {
    if (!editingPromo) return;
    try {
      await update.mutateAsync({ id: editingPromo.id, input });
      toast.success("Perubahan disimpan.");
      setModalOpen(false);
      setEditingPromo(null);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleSubmit(input: AdminCreatePromoCodeInput | AdminUpdatePromoCodeInput) {
    if (editingPromo) {
      await handleUpdate(input as AdminUpdatePromoCodeInput);
    } else {
      await handleCreate(input as AdminCreatePromoCodeInput);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Hapus kode promo ini? Tindakan ini tidak dapat dibatalkan.")) return;
    try {
      await remove.mutateAsync(id);
      toast.success("Kode promo dihapus.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Promo</h1>
        <Button onClick={openCreate}>Buat kode promo</Button>
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
          Gagal memuat kode promo: {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="overflow-x-auto rounded-lg border">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">Kode</th>
                <th className="px-4 py-3 text-left font-medium">Diskon</th>
                <th className="px-4 py-3 text-left font-medium">Digunakan / Maks</th>
                <th className="px-4 py-3 text-left font-medium">Kadaluarsa</th>
                <th className="px-4 py-3 text-right font-medium">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {promos?.map((promo) => (
                <tr key={promo.id} className="border-t hover:bg-muted/40">
                  <td className="px-4 py-3 font-medium">{promo.code}</td>
                  <td className="px-4 py-3">{discountLabel(promo)}</td>
                  <td className="px-4 py-3">{usageText(promo)}</td>
                  <td className="px-4 py-3">{expiryText(promo.expires_at)}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button size="sm" variant="outline" onClick={() => openEdit(promo)}>
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleDelete(promo.id)}
                        disabled={remove.isPending}
                      >
                        Hapus
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
              {(promos?.length ?? 0) === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">
                    Belum ada kode promo.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <PromoModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        promo={editingPromo}
        onSubmit={handleSubmit}
        isPending={create.isPending || update.isPending}
      />
    </div>
  );
}
