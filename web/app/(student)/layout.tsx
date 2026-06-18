"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Script from "next/script";
import { useAuthStore } from "@/stores/auth";
import { Header } from "@/components/shell/Header";
import { BottomNav } from "@/components/shell/BottomNav";

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
  const router = useRouter();
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (hydrated && !token) {
      router.replace("/login");
    }
  }, [hydrated, token, router]);

  if (!hydrated) {
    return null;
  }

  if (!token) {
    return null;
  }

  return (
    <>
      <Script
        id="midtrans-snap"
        src={SNAP_SRC}
        strategy="afterInteractive"
        data-client-key={SNAP_CLIENT_KEY}
      />
      <div className="flex min-h-screen flex-col bg-paper">
        <Header />
        <main className="flex-1 pb-20 md:pb-0">{children}</main>
        <BottomNav />
      </div>
    </>
  );
}