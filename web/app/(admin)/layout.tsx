"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { useMe } from "@/lib/hooks/auth";
import { AppShell } from "@/components/shell/AppShell";
import { ADMIN_ROLES } from "@/lib/nav-config";
import type { UserRole } from "@/lib/nav-config";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const storeUser = useAuthStore((s) => s.user);
  const [hydrated, setHydrated] = useState(false);

  const storeRole = storeUser?.role as UserRole | undefined;
  const needsMeFetch = hydrated && !!token && !storeRole;
  const me = useMe({ enabled: needsMeFetch });

  const effectiveRole = storeRole ?? (me.data?.role as UserRole | undefined);

  useEffect(() => {
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    if (!token) {
      router.replace("/login");
      return;
    }
    if (me.isError) {
      router.replace("/login");
      return;
    }
    if (effectiveRole && !ADMIN_ROLES.includes(effectiveRole)) {
      router.replace("/");
    }
  }, [hydrated, token, effectiveRole, me.isError, router]);

  if (!hydrated || !token || !effectiveRole || !ADMIN_ROLES.includes(effectiveRole)) {
    return (
      <div className="flex min-h-screen items-center justify-center text-ink-500">
        Memuat…
      </div>
    );
  }

  return <AppShell role={effectiveRole}>{children}</AppShell>;
}
