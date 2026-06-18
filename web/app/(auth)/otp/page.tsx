"use client";

import * as React from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, ShieldCheck, Loader2, AlertCircle } from "lucide-react";

import { useVerifyOtp } from "@/lib/hooks/auth";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { OtpInput } from "@/components/auth/OtpInput";

function maskContact(contact: string): string {
  if (!contact) return "•••••";
  if (contact.includes("@")) {
    return contact.replace(/(.{2}).*(@.*)/, "$1•••$2");
  }
  return contact.replace(/(\+?\d{2,4})\d+(\d{3})/, "$1•••••$2");
}

function OtpForm() {
  const router = useRouter();
  const params = useSearchParams();
  const identifier = params.get("id") ?? "";

  const verify = useVerifyOtp();
  const [digits, setDigits] = React.useState<string[]>(Array(6).fill(""));
  const [error, setError] = React.useState<string | null>(null);

  const code = digits.join("");
  const allFilled = digits.every((d) => d !== "");

  const onVerify = async () => {
    if (code.length < 6) {
      setError("Masukkan 6 digit OTP");
      return;
    }
    setError(null);
    try {
      await verify.mutateAsync({ identifier, code });
      router.push("/");
    } catch (err) {
      const msg =
        err instanceof ApiError ? err.message : "Kode OTP tidak valid atau kedaluwarsa.";
      setError(msg);
    }
  };

  const loading = verify.isPending;

  return (
    <div className="w-full max-w-[372px]">
      <button
        type="button"
        onClick={() => router.push("/login")}
        className="mb-8 flex items-center gap-1.5 text-[13px] font-semibold text-ink-400 transition-colors hover:text-ink-600"
      >
        <ArrowLeft size={15} />
        Kembali
      </button>

      <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-[14px] bg-[linear-gradient(135deg,#EEF0FD,#DDE1FB)]">
        <ShieldCheck size={26} className="text-brand-600" />
      </div>

      <h2 className="font-serif text-[26px] font-bold leading-tight tracking-[-0.01em] text-ink-900">
        Verifikasi OTP
      </h2>
      <p className="mt-2.5 mb-7 text-[13.5px] leading-[1.6] text-ink-500">
        Kode 6 digit telah dikirim ke{" "}
        <strong className="text-ink-700">{maskContact(identifier)}</strong>. Kode berlaku
        selama 5 menit.
      </p>

      <OtpInput
        value={digits}
        onChange={setDigits}
        hasError={!!error}
        disabled={loading}
      />

      {error && (
        <div
          role="alert"
          className="mt-4 flex items-center gap-2 text-[12.5px] text-danger"
        >
          <AlertCircle size={14} className="shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <Button
        type="button"
        onClick={onVerify}
        disabled={loading || !allFilled}
        className="mt-6 h-12 w-full text-[15px]"
      >
        {loading ? (
          <span className="flex items-center gap-2">
            <Loader2 size={16} className="animate-spin" />
            Memverifikasi...
          </span>
        ) : (
          "Verifikasi & Masuk"
        )}
      </Button>

      <p className="mt-5 text-center text-[12px] leading-[1.6] text-ink-400">
        Tidak menerima kode? Periksa folder spam atau tunggu beberapa saat sebelum
        meminta ulang.
      </p>
    </div>
  );
}

export default function OtpPage() {
  return (
    <React.Suspense fallback={<div className="w-full max-w-[372px]" />}>
      <OtpForm />
    </React.Suspense>
  );
}