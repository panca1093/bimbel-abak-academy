"use client";

import type { LucideIcon } from "lucide-react";

interface UnderMaintenanceProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  estimatedTimeline?: string;
}

export function UnderMaintenance({
  icon: Icon,
  title,
  description = "Fitur ini sedang dalam pengembangan",
  estimatedTimeline,
}: UnderMaintenanceProps) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="flex size-20 items-center justify-center rounded-[20px] bg-[var(--md-sys-color-surface-container-high)] text-[var(--md-sys-color-on-surface-variant)]">
        <Icon size={40} />
      </div>
      <h2 className="text-title-large mt-6 font-semibold text-[var(--md-sys-color-on-surface)]">
        {title}
      </h2>
      <p className="text-body-medium mt-2 max-w-sm text-[var(--md-sys-color-on-surface-variant)]">
        {description}
      </p>
      {estimatedTimeline ? (
        <p className="text-label mt-4 text-[var(--md-sys-color-on-surface-variant)]">
          {estimatedTimeline}
        </p>
      ) : null}
    </div>
  );
}
