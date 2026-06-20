"use client";

/** A small circular progress ring, used e.g. for popular lesson items. */
export function Ring({
  value,
  size = 48,
  thickness = 6,
}: {
  value: number; // 0–1
  size?: number;
  thickness?: number;
}) {
  const r = (size - thickness) / 2;
  const circ = 2 * Math.PI * r;
  const filled = circ * Math.min(value, 1);
  const cx = size / 2;
  const cy = size / 2;

  return (
    <svg
      width={size}
      height={size}
      viewBox={`0 0 ${size} ${size}`}
      className="shrink-0"
      role="img"
      aria-label={`${Math.round(value * 100)}%`}
    >
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke="var(--color-brand-100)"
        strokeWidth={thickness}
      />
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke="var(--color-brand-600)"
        strokeWidth={thickness}
        strokeDasharray={`${filled} ${circ - filled}`}
        strokeDashoffset={0}
        strokeLinecap="round"
        transform={`rotate(-90 ${cx} ${cy})`}
      />
    </svg>
  );
}
