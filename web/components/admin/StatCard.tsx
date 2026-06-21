import type { LucideIcon } from "lucide-react";

type StatAccent = "primary" | "secondary" | "error" | "tertiary";

interface StatCardProps {
  label: string;
  value: string;
  icon?: LucideIcon;
  trend?: string;
  accent?: StatAccent;
}

const ACCENT: Record<StatAccent, { box: string; on: string; trend: string }> = {
  primary:   { box: "var(--md-sys-color-primary-container)",   on: "var(--md-sys-color-on-primary-container)",   trend: "var(--md-sys-color-primary)" },
  secondary: { box: "var(--md-sys-color-secondary-container)", on: "var(--md-sys-color-on-secondary-container)", trend: "var(--md-sys-color-on-surface-variant)" },
  error:     { box: "var(--md-sys-color-error-container)",     on: "var(--md-sys-color-on-error-container)",     trend: "var(--md-sys-color-error)" },
  tertiary:  { box: "var(--md-sys-color-tertiary-container)",  on: "var(--md-sys-color-on-tertiary-container)",  trend: "var(--md-sys-color-on-surface-variant)" },
};

export function StatCard({ label, value, icon: Icon, trend, accent = "primary" }: StatCardProps) {
  const c = ACCENT[accent];
  return (
    <div className="md-card-elevated">
      <div className="flex items-start justify-between">
        <div>
          <div className="text-label color-on-surface-variant mb-2">{label}</div>
          <div className="text-title-large mb-2" style={{ fontWeight: 600 }}>{value}</div>
          {trend ? <div className="text-label" style={{ color: c.trend }}>{trend}</div> : null}
        </div>
        {Icon ? (
          <div
            className="flex size-10 items-center justify-center rounded-[12px]"
            style={{ backgroundColor: c.box, color: c.on }}
          >
            <Icon size={20} />
          </div>
        ) : null}
      </div>
    </div>
  );
}
