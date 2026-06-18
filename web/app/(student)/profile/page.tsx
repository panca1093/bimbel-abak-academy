"use client";

import { useEffect, useState } from "react";
import { AlertCircle, UserRound } from "lucide-react";
import { useProfile, useUpdateProfile } from "@/lib/hooks/students";
import { useAuthStore } from "@/stores/auth";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { ChangePasswordForm } from "@/components/profile/ChangePasswordForm";
import { toast } from "sonner";

function initials(name?: string): string {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/);
  return (parts[0]?.[0] ?? "") + (parts[1]?.[0] ?? "");
}

export default function ProfilePage() {
  const { data: profile, isLoading, isError, error, refetch } = useProfile();
  const updateUser = useAuthStore((s) => s.user);
  const updateProfile = useUpdateProfile();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    if (profile) {
      setName(profile.name ?? "");
      setEmail(profile.email ?? "");
    }
  }, [profile]);

  useEffect(() => {
    setHydrated(true);
  }, []);

  function handleSave(e: React.FormEvent) {
    e.preventDefault();
    updateProfile.mutate(
      { name, email },
      {
        onSuccess: (updated) => {
          toast.success("Profil tersimpan.");
          if (updated) {
            setName(updated.name ?? name);
            setEmail(updated.email ?? email);
          }
        },
        onError: (err) => {
          const message =
            err instanceof Error && err.message
              ? err.message
              : "Gagal menyimpan profil.";
          toast.error(message);
        },
      }
    );
  }

  const displayName = profile?.name ?? updateUser?.name ?? "";

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
          Profil
        </h1>
        <p className="mt-2 text-sm text-ink-500">
          Kelola nama, email, dan kata sandi akun Anda.
        </p>
      </header>

      {isError ? (
        <Card className="border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              Gagal memuat profil.
              {error instanceof Error && error.message ? ` ${error.message}` : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Coba lagi
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
          <Card className="px-6 py-6">
            <div className="mb-6 flex items-center gap-4">
              {isLoading ? (
                <Skeleton className="size-16 rounded-full" />
              ) : (
                <Avatar className="size-16 rounded-full">
                  <AvatarFallback className="size-16 rounded-full bg-brand-100 text-brand-700 font-semibold">
                    {initials(displayName) || "?"}
                  </AvatarFallback>
                </Avatar>
              )}
              <div className="min-w-0">
                {isLoading ? (
                  <Skeleton className="h-5 w-40" />
                ) : (
                  <div className="font-serif text-xl font-semibold text-ink-900">
                    {displayName || "Tanpa nama"}
                  </div>
                )}
                {isLoading ? (
                  <Skeleton className="mt-2 h-4 w-56" />
                ) : (
                  <div className="truncate text-sm text-ink-500">
                    {profile?.username ? `${profile.username} · ` : ""}
                    {profile?.email ?? ""}
                  </div>
                )}
              </div>
            </div>

            <form onSubmit={handleSave} className="flex flex-col gap-4" noValidate>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="name">Nama lengkap</Label>
                  <Input
                    id="name"
                    value={hydrated ? name : ""}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Nama lengkap"
                    disabled={isLoading || updateProfile.isPending}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={hydrated ? email : ""}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="email@contoh.com"
                    disabled={isLoading || updateProfile.isPending}
                  />
                </div>
              </div>

              <div className="pt-2">
                <Button type="submit" disabled={updateProfile.isPending || isLoading}>
                  {updateProfile.isPending ? "Menyimpan…" : "Simpan perubahan"}
                </Button>
              </div>
            </form>
          </Card>

          <Card className="px-6 py-6">
            <div className="mb-4 flex items-center gap-2">
              <UserRound className="size-5 text-brand-600" />
              <h2 className="font-serif text-lg font-semibold text-ink-900">
                Keamanan
              </h2>
            </div>
            <p className="mb-4 text-sm text-ink-500">
              Kata sandi baru minimal 8 karakter.
            </p>
            <ChangePasswordForm />
          </Card>
        </div>
      )}
    </div>
  );
}