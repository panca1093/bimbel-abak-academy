"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, BookOpen, PlayCircle, Receipt, ShoppingCart } from "lucide-react";
import { useCartStore } from "@/stores/cart";

const TABS = [
  { href: "/", label: "Beranda", icon: Home },
  { href: "/catalog", label: "Katalog", icon: BookOpen },
  { href: "/courses", label: "Kursus", icon: PlayCircle },
  { href: "/orders", label: "Pesanan", icon: Receipt },
  { href: "/cart", label: "Keranjang", icon: ShoppingCart },
];

export function BottomNav() {
  const pathname = usePathname();
  const cartCount = useCartStore((s) => s.count);

  return (
    <nav className="fixed inset-x-0 bottom-0 z-40 flex h-16 items-stretch border-t border-line bg-surface/95 backdrop-blur md:hidden">
      {TABS.map((tab) => {
        const active =
          tab.href === "/" ? pathname === "/" : pathname.startsWith(tab.href);
        const Icon = tab.icon;
        return (
          <Link
            key={tab.href}
            href={tab.href}
            className={
              "relative flex flex-1 flex-col items-center justify-center gap-1 text-[10px] font-medium transition-colors " +
              (active ? "text-brand-700" : "text-ink-500")
            }
          >
            <span className="relative">
              <Icon size={22} />
              {tab.href === "/cart" && cartCount > 0 && (
                <span className="absolute -right-2 -top-1 flex size-4 items-center justify-center rounded-full bg-brand-600 text-[9px] font-bold text-white">
                  {cartCount > 9 ? "9+" : cartCount}
                </span>
              )}
            </span>
            <span>{tab.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}