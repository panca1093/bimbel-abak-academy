"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRouter } from "next/navigation";
import { ShoppingCart, User, LogOut, ChevronDown } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { useAuthStore } from "@/stores/auth";
import { useCartStore } from "@/stores/cart";
import { useLogout } from "@/lib/hooks/auth";

const NAV_LINKS = [
  { href: "/", label: "Beranda" },
  { href: "/catalog", label: "Katalog" },
  { href: "/courses", label: "Kursus" },
  { href: "/orders", label: "Pesanan" },
];

function AbakMark({ size = 28 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 120 120"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="abak academy"
      className="text-brand-600"
    >
      <circle cx="44" cy="34" r="15" fill="currentColor" />
      <path d="M22 104 Q22 64 44 64 Q66 64 66 104 Z" fill="currentColor" />
      <path d="M62 104 Q62 78 80 78 Q98 78 98 104 Z" fill="#1E978A" />
      <path d="M80 44 L96 51 L80 58 L64 51 Z" fill="#D99A2B" />
      <circle cx="80" cy="62" r="11" fill="#1E978A" />
      <rect x="79" y="44" width="2.5" height="9" fill="#D99A2B" />
    </svg>
  );
}

export function Header() {
  const pathname = usePathname();
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const cartCount = useCartStore((s) => s.count);
  const logout = useLogout();

  const initial = (user?.name ?? user?.email ?? user?.username ?? "A")
    .trim()
    .charAt(0)
    .toUpperCase();

  function handleLogout() {
    logout.mutate(undefined, {
      onSettled: () => {
        router.replace("/login");
      },
    });
  }

  return (
    <header className="sticky top-0 z-40 hidden border-b border-line bg-surface/90 backdrop-blur md:block">
      <div className="mx-auto flex h-16 max-w-6xl items-center gap-6 px-6">
        <Link href="/" className="flex items-center gap-2">
          <AbakMark size={28} />
          <span className="font-serif text-lg font-extrabold tracking-tight text-ink-900">
            abak
            <span className="ml-1 text-[0.7em] uppercase tracking-[0.18em] text-gold">
              academy
            </span>
          </span>
        </Link>

        <nav className="flex flex-1 items-center gap-1">
          {NAV_LINKS.map((link) => {
            const active =
              link.href === "/"
                ? pathname === "/"
                : pathname.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className={
                  "rounded-lg px-3 py-2 text-sm font-medium transition-colors " +
                  (active
                    ? "bg-brand-50 text-brand-700"
                    : "text-ink-600 hover:bg-brand-50/60 hover:text-brand-700")
                }
              >
                {link.label}
              </Link>
            );
          })}
        </nav>

        <Link
          href="/cart"
          className="relative flex size-9 items-center justify-center rounded-full text-ink-700 hover:bg-brand-50 hover:text-brand-700"
          aria-label="Keranjang"
        >
          <ShoppingCart size={20} />
          {cartCount > 0 && (
            <Badge className="absolute -right-1 -top-1 size-5 justify-center rounded-full bg-brand-600 px-0 py-0 text-[10px] font-bold text-white">
              {cartCount > 9 ? "9+" : cartCount}
            </Badge>
          )}
        </Link>

        <DropdownMenu>
          <DropdownMenuTrigger className="flex items-center gap-2 rounded-full py-1 pl-1 pr-2 outline-none hover:bg-brand-50">
            <Avatar size="sm">
              <AvatarFallback className="bg-brand-600 text-xs font-semibold text-white">
                {initial}
              </AvatarFallback>
            </Avatar>
            <ChevronDown size={14} className="text-ink-500" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel className="truncate text-ink-900">
              {user?.name ?? user?.email ?? "Akun"}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => router.push("/profile")}>
              <User size={16} />
              <span className="ml-2">Profil</span>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut size={16} />
              <span className="ml-2">Keluar</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}