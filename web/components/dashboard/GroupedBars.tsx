"use client";

/** A simple grouped vertical bar chart for exam progress. */
export function GroupedBars({
  groups,
  labels,
  series,
  height = 160,
}: {
  groups: [number, number][];
  labels: string[];
  series: { label: string; color: string }[];
  height?: number;
}) {
  if (groups.length === 0) return null;
  const yMax = Math.max(...groups.flat(), 1);

  return (
    <div className="flex items-end gap-4" style={{ height }}>
      {groups.map(([a, b], i) => {
        const hA = (a / yMax) * (height - 20);
        const hB = (b / yMax) * (height - 20);
        return (
          <div key={i} className="flex flex-1 flex-col items-center gap-1">
            <div className="flex w-full items-end justify-center gap-[3px]">
              <div
                className="w-3 rounded-t"
                style={{ height: Math.max(hA, 2), background: series[0]?.color ?? "var(--color-brand-600)" }}
                title={`${series[0]?.label ?? ""}: ${a}`}
              />
              <div
                className="w-3 rounded-t"
                style={{ height: Math.max(hB, 2), background: series[1]?.color ?? "var(--color-brand-200)" }}
                title={`${series[1]?.label ?? ""}: ${b}`}
              />
            </div>
            <span className="text-center text-[10px] text-ink-500">{labels[i]}</span>
          </div>
        );
      })}
    </div>
  );
}
