"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { User, Mail, Lock, Eye, EyeOff, CheckCircle2, Loader2, AlertCircle } from "lucide-react";
import { toast } from "sonner";

import { useRegister } from "@/lib/hooks/auth";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type Errors = {
  name?: string;
  email?: string;
  password?: string;
  confirm?: string;
};

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function RegisterPage() {
  const router = useRouter();
  const register = useRegister();
  const [name, setName] = React.useState("");
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [showPw, setShowPw] = React.useState(false);
  const [showCf, setShowCf] = React.useState(false);
  const [errors, setErrors] = React.useState<Errors>({});

  const pwStrength =
    password.length === 0
      ? null
      : password.length < 6
        ? "weak"
        : password.length < 10
          ? "fair"
          : "strong";

  const strengthColor: Record<string, string> = {
    weak: "bg-danger",
    fair: "bg-warn",
    strong: "bg-success",
  };
  const strengthText: Record<string, string> = {
    weak: "Lemah",
    fair: "Sedang",
    strong: "Kuat",
  };
  const strengthTextColor: Record<string, string> = {
    weak: "text-danger",
    fair: "text-warn",
    strong: "text-success",
  };
  const strengthWidth: Record<string, string> = {
    weak: "w-1/3",
    fair: "w-2/3",
    strong: "w-full",
  };

  const validate = (): Errors => {
    const e: Errors = {};
    if (!name.trim()) e.name = "Nama wajib diisi";
    if (!EMAIL_RE.test(email)) e.email = "Email tidak valid";
    if (password.length < 8) e.password = "Min. 8 karakter";
    if (confirm !== password) e.confirm = "Kata sandi tidak cocok";
    return e;
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const errs = validate();
    setErrors(errs);
    if (Object.keys(errs).length > 0) return;

    try {
      await register.mutateAsync({ name: name.trim(), email: email.trim(), password });
      toast.success("Pendaftaran berhasil. Silakan masuk.");
      router.push("/login");
    } catch (err) {
      const msg =
        err instanceof ApiError ? err.message : "Gagal mendaftar. Coba lagi.";
      toast.error(msg);
    }
  };

  const loading = register.isPending;
  const passwordsMatch = confirm.length > 0 && confirm === password;

  return (
    <div className="w-full max-w-[372px]">
      <div className="mb-7">
        <div className="mb-2 text-[11.5px] font-bold uppercase tracking-[0.06em] text-success">
          Mulai sekarang
        </div>
        <h2 className="font-serif text-[27px] font-bold leading-tight tracking-[-0.01em] text-ink-900">
          Buat Akun Baru ✨
        </h2>
        <p className="mt-2 text-[13.5px] leading-[1.55] text-ink-500">
          Isi data di bawah untuk mulai belajar bersama Abak Academy.
        </p>
      </div>

      <form onSubmit={onSubmit} noValidate>
        <div className="mb-3.5">
          <Label htmlFor="name" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Nama Lengkap
          </Label>
          <div className="relative">
            <Input
              id="name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                setErrors((p) => ({ ...p, name: undefined }));
              }}
              placeholder="masukkan nama lengkap"
              autoComplete="name"
              className="h-11 pl-10"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">
              <User size={16} />
            </span>
          </div>
          {errors.name && (
            <div className="mt-1 text-[11.5px] text-danger">{errors.name}</div>
          )}
        </div>

        <div className="mb-3.5">
          <Label htmlFor="email" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Email
          </Label>
          <div className="relative">
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                setErrors((p) => ({ ...p, email: undefined }));
              }}
              placeholder="nama@email.com"
              autoComplete="email"
              className="h-11 pl-10"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">
              <Mail size={16} />
            </span>
          </div>
          {errors.email && (
            <div className="mt-1 text-[11.5px] text-danger">{errors.email}</div>
          )}
        </div>

        <div className="mb-1.5">
          <Label htmlFor="password" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Kata Sandi
          </Label>
          <div className="relative">
            <Input
              id="password"
              type={showPw ? "text" : "password"}
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setErrors((p) => ({ ...p, password: undefined }));
              }}
              placeholder="min. 8 karakter"
              autoComplete="new-password"
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
          {pwStrength && (
            <div className="mt-1.5 flex items-center gap-2">
              <div className="h-1 flex-1 overflow-hidden rounded-full bg-line">
                <div
                  className={`h-full rounded-full transition-all duration-200 ${strengthColor[pwStrength]} ${strengthWidth[pwStrength]}`}
                />
              </div>
              <span className={`text-[11px] font-bold ${strengthTextColor[pwStrength]}`}>
                {strengthText[pwStrength]}
              </span>
            </div>
          )}
          {errors.password && (
            <div className="mt-1 text-[11.5px] text-danger">{errors.password}</div>
          )}
        </div>

        <div className="mb-5">
          <Label htmlFor="confirm" className="mb-1.5 text-[12.5px] font-semibold text-ink-600">
            Konfirmasi Kata Sandi
          </Label>
          <div className="relative">
            <Input
              id="confirm"
              type={showCf ? "text" : "password"}
              value={confirm}
              onChange={(e) => {
                setConfirm(e.target.value);
                setErrors((p) => ({ ...p, confirm: undefined }));
              }}
              placeholder="••••••••"
              autoComplete="new-password"
              className="h-11 pl-10 pr-11"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">
              {passwordsMatch ? (
                <CheckCircle2 size={16} className="text-success" />
              ) : (
                <Lock size={16} />
              )}
            </span>
            <button
              type="button"
              onClick={() => setShowCf((p) => !p)}
              aria-label={showCf ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-ink-400 transition-colors hover:text-ink-600"
            >
              {showCf ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
          {errors.confirm && (
            <div className="mt-1 text-[11.5px] text-danger">{errors.confirm}</div>
          )}
        </div>

        <Button
          type="submit"
          disabled={loading}
          className="h-12 w-full text-[15px]"
        >
          {loading ? (
            <span className="flex items-center gap-2">
              <Loader2 size={16} className="animate-spin" />
              Mendaftar...
            </span>
          ) : (
            "Daftar & Verifikasi OTP"
          )}
        </Button>
      </form>

      <p className="mt-6 text-center text-[13px] text-ink-500">
        Sudah punya akun?{" "}
        <button
          type="button"
          onClick={() => router.push("/login")}
          className="font-bold text-success transition-colors hover:text-success/80"
        >
          Masuk
        </button>
      </p>
    </div>
  );
}