"use client";

import { Construction } from "lucide-react";

interface ComingSoonProps {
  title?: string;
}

export function ComingSoon({ title = "Coming soon" }: ComingSoonProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="rounded-2xl bg-brand-50 p-5">
        <Construction className="size-10 text-brand-600" />
      </div>
      <h2 className="mt-5 font-serif text-2xl font-bold text-ink-900">
        {title}
      </h2>
      <p className="mt-2 max-w-sm text-sm text-ink-500">
        Fitur ini sedang dalam pengembangan dan akan segera hadir.
      </p>
    </div>
  );
}
