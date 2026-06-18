const rupiahFormatter = new Intl.NumberFormat("id-ID", {
  maximumFractionDigits: 0,
  useGrouping: true,
});

export function formatRupiah(n: number): string {
  const safe = Number.isFinite(n) ? n : 0;
  return `Rp${rupiahFormatter.format(Math.round(safe))}`;
}