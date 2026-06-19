"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { User, Lock, Eye, EyeOff, Loader2, AlertCircle } from "lucide-react";

import { useLogin } from "@/lib/hooks/auth";
import { redirectForRole } from "@/lib/auth-redirect";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function LoginPage() {
  const router = useRouter();
  const login = useLogin();
  const [identifier, setIdentifier] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [showPw, setShowPw] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      const data = await login.mutateAsync({ identifier, password });
      if (data.otp_required) {
        if (data.pending_token && typeof window !== "undefined") {
          sessionStorage.setItem("abak-pending-token", data.pending_token);
        }
        router.push(`/otp?id=${encodeURIComponent(identifier)}`);
        return;
      }
      router.push(redirectForRole(data.user?.role));
    } catch (err) {
      const msg =
        err instanceof ApiError ? err.message : "Gagal masuk. Coba lagi.";
      setError(msg);
    }
  };

  const loading = login.isPending;

  return (
    <div className="w-full max-w-[372px]">
      <div className="mb-7">
        <div className="mb-2 text-[11.5px] font-bold uppercase tracking-[0.06em] text-success">
          Selamat datang kembali
        </div>
        <h2 className="font-serif text-[27px] font-bold leading-tight tracking-[-0.01em] text-ink-900">
          Masuk ke Akun Kamu 👋
        </h2>
        <p className="mt-2 text-[13.5px] leading-[1.55] text-ink-500">
          Masukkan email/username dan kata sandi untuk melanjutkan.
        </p>
      </div>

      <form onSubmit={onSubmit} noValidate>
        <div className="mb-4">
          <Label htmlFor="identifier" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Email atau Username
          </Label>
          <div className="relative">
            <Input
              id="identifier"
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              placeholder="email atau username"
              autoComplete="username"
              required
              className="h-11 pl-10"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">
              <User size={16} />
            </span>
          </div>
        </div>

        <div className="mb-2.5">
          <Label htmlFor="password" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Kata Sandi
          </Label>
          <div className="relative">
            <Input
              id="password"
              type={showPw ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
              required
              className="h-11 pl-10 pr-11"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">
              <Lock size={16} />
            </span>
            <button
              type="button"
              onClick={() => setShowPw((p) => !p)}
              aria-label={showPw ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-ink-400 transition-colors hover:text-ink-600"
            >
              {showPw ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </div>

        {error && (
          <div
            role="alert"
            className="mb-4 mt-3 flex items-center gap-2 rounded-[10px] border border-danger/20 bg-danger-bg px-3 py-2.5 text-[12.5px] text-danger"
          >
            <AlertCircle size={14} className="shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <Button
          type="submit"
          disabled={loading}
          className="mt-5 h-12 w-full text-[15px]"
        >
          {loading ? (
            <span className="flex items-center gap-2">
              <Loader2 size={16} className="animate-spin" />
              Masuk...
            </span>
          ) : (
            "Masuk"
          )}
        </Button>
      </form>

      <p className="mt-6 text-center text-[13px] text-ink-500">
        Belum punya akun?{" "}
        <button
          type="button"
          onClick={() => router.push("/register")}
          className="font-bold text-success transition-colors hover:text-success/80"
        >
          Daftar sekarang
        </button>
      </p>
    </div>
  );
}