"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { useMe } from "@/lib/hooks/auth";
import { adminHomeForRole } from "@/lib/auth-redirect";
import type { UserRole } from "@/lib/nav-config";

export default function AdminIndexPage() {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const storeRole = user?.role as UserRole | undefined;
  const me = useMe({ enabled: !storeRole });
  const role = storeRole ?? (me.data?.role as UserRole | undefined);

  useEffect(() => {
    if (!role) return;
    router.replace(adminHomeForRole(role));
  }, [role, router]);

  return null;
}
