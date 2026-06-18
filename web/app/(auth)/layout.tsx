"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { BrandPanel } from "@/components/auth/BrandPanel";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  const router = useRouter();

  useEffect(() => {
    if (token) router.replace("/");
  }, [token, router]);

  if (token) return null;

  return (
    <div className="flex min-h-screen">
      <BrandPanel mode="login" />
      <div className="flex flex-1 items-center justify-center bg-surface px-6 py-12">
        {children}
      </div>
    </div>
  );
}