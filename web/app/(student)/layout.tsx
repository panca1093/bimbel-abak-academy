"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Script from "next/script";
import { useAuthStore } from "@/stores/auth";
import { AppShell } from "@/components/shell/AppShell";
import { ADMIN_ROLES } from "@/lib/nav-config";
import type { UserRole } from "@/lib/nav-config";

const SNAP_SRC =
  process.env.NEXT_PUBLIC_MIDTRANS_SNAP_URL ??
  "https://app.sandbox.midtrans.com/snap/snap.js";
const SNAP_CLIENT_KEY = process.env.NEXT_PUBLIC_MIDTRANS_CLIENT_KEY ?? "";

export default function StudentLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const router = useRouter();
  const [hydrated, setHydrated] = useState(false);

  const role = user?.role as UserRole | undefined;

  useEffect(() => {
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    if (!token) {
      router.replace("/login");
      return;
    }
    if (role && ADMIN_ROLES.includes(role)) {
      router.replace("/admin");
    }
  }, [hydrated, token, role, router]);

  if (!hydrated || !token || (role && ADMIN_ROLES.includes(role))) {
    return (
      <div className="flex min-h-screen items-center justify-center text-ink-500">
        Memuat…
      </div>
    );
  }

  return (
    <>
      <Script
        id="midtrans-snap"
        src={SNAP_SRC}
        strategy="afterInteractive"
        data-client-key={SNAP_CLIENT_KEY}
      />
      <AppShell role="student">{children}</AppShell>
    </>
  );
}
