"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminRevenue } from "@/lib/hooks/admin-revenue";
import { formatRupiah } from "@/lib/format";
import type { AdminRevenue, RevenueByTypeItem } from "@/lib/types";

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan.";
}

function orderCount(revenue?: AdminRevenue): number {
  if (!revenue) return 0;
  return Object.values(revenue.by_type).reduce((sum, item) => sum + (item.count ?? 0), 0);
}

function averageOrderValue(revenue?: AdminRevenue): number {
  const count = orderCount(revenue);
  if (!count || !revenue) return 0;
  return revenue.total / count;
}

function typeEntries(revenue?: AdminRevenue): [string, RevenueByTypeItem][] {
  if (!revenue) return [];
  return Object.entries(revenue.by_type).sort((a, b) => b[1].total - a[1].total);
}

function maxTypeTotal(revenue?: AdminRevenue): number {
  const entries = typeEntries(revenue);
  if (entries.length === 0) return 1;
  return Math.max(...entries.map(([, item]) => item.total));
}

export default function RevenuePage() {
  const { data: revenue, isLoading, isError, error } = useAdminRevenue();

  const entries = typeEntries(revenue);
  const max = maxTypeTotal(revenue);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Pendapatan</h1>
      </div>

      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          Gagal memuat revenue: {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && revenue && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Total pendapatan</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-semibold">{formatRupiah(revenue.total)}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Pesanan</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-semibold">{orderCount(revenue)}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Rata-rata nilai pesanan</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-semibold">{formatRupiah(averageOrderValue(revenue))}</div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Pendapatan berdasarkan jenis</CardTitle>
            </CardHeader>
            <CardContent>
              {entries.length === 0 ? (
                <div className="text-sm text-muted-foreground">Belum ada data pendapatan.</div>
              ) : (
                <div className="space-y-4">
                  {entries.map(([type, item]) => {
                    const pct = max > 0 ? (item.total / max) * 100 : 0;
                    return (
                      <div key={type} className="space-y-1">
                        <div className="flex items-center justify-between text-sm">
                          <span className="capitalize font-medium">{type}</span>
                          <span className="text-muted-foreground">
                            {formatRupiah(item.total)} · {item.count} pesanan
                          </span>
                        </div>
                        <div className="h-3 w-full overflow-hidden rounded-full bg-muted">
                          <div
                            className="h-full rounded-full bg-primary"
                            style={{ width: `${pct}%` }}
                            data-testid={`bar-${type}`}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Produk terlaris</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto rounded-lg border">
                <table className="w-full text-sm">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-4 py-3 text-left font-medium">Produk</th>
                      <th className="px-4 py-3 text-left font-medium">Pesanan</th>
                      <th className="px-4 py-3 text-right font-medium">Pendapatan</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-t">
                      <td colSpan={3} className="px-4 py-8 text-center text-muted-foreground">
                        Belum ada data produk terlaris untuk periode ini.
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
