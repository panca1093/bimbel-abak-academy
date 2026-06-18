"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

export interface OtpInputProps {
  value: string[];
  onChange: (value: string[]) => void;
  length?: number;
  hasError?: boolean;
  disabled?: boolean;
}

export function OtpInput({
  value,
  onChange,
  length = 6,
  hasError = false,
  disabled = false,
}: OtpInputProps) {
  const refs = React.useRef<Array<HTMLInputElement | null>>([]);

  const focusAt = (i: number) => {
    const idx = Math.max(0, Math.min(length - 1, i));
    refs.current[idx]?.focus();
  };

  const setDigit = (i: number, raw: string) => {
    const digit = raw.replace(/\D/g, "").slice(-1);
    const next = [...value];
    while (next.length < length) next.push("");
    next[i] = digit;
    onChange(next.slice(0, length));
    if (digit && i < length - 1) focusAt(i + 1);
  };

  const handleKey = (i: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Backspace") {
      e.preventDefault();
      const next = [...value];
      if (next[i]) {
        next[i] = "";
        onChange(next);
      } else if (i > 0) {
        focusAt(i - 1);
        const prev = [...next];
        prev[i - 1] = "";
        onChange(prev);
      }
    } else if (e.key === "ArrowLeft") {
      e.preventDefault();
      focusAt(i - 1);
    } else if (e.key === "ArrowRight") {
      e.preventDefault();
      focusAt(i + 1);
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, length);
    if (!pasted) return;
    const next: string[] = Array.from({ length }, (_, k) => pasted[k] ?? "");
    onChange(next);
    focusAt(Math.min(pasted.length, length - 1));
  };

  return (
    <div className="flex gap-2.5" onPaste={handlePaste}>
      {Array.from({ length }, (_, i) => {
        const digit = value[i] ?? "";
        return (
          <input
            key={i}
            ref={(el) => {
              refs.current[i] = el;
            }}
            type="text"
            inputMode="numeric"
            autoComplete={i === 0 ? "one-time-code" : "off"}
            maxLength={1}
            value={digit}
            disabled={disabled}
            onChange={(e) => setDigit(i, e.target.value)}
            onKeyDown={(e) => handleKey(i, e)}
            onFocus={(e) => e.target.select()}
            className={cn(
              "h-14 w-12 rounded-[10px] text-center font-serif text-[22px] font-bold text-ink-900 outline-none transition-colors",
              "border-2",
              digit
                ? "border-brand-600 bg-brand-50"
                : hasError
                  ? "border-danger"
                  : "border-line bg-surface",
              "focus-visible:border-brand-600 focus-visible:ring-[3px] focus-visible:ring-brand-600/20",
              disabled && "opacity-60"
            )}
          />
        );
      })}
    </div>
  );
}