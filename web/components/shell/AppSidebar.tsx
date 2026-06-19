"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronDown, LogOut } from "lucide-react";
import { useAuthStore } from "@/stores/auth";
import { useLogout } from "@/lib/hooks/auth";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { AbakMark } from "./AbakMark";
import {
  NAV_CONFIG,
  roleDisplayName,
  type UserRole,
  type NavItem,
} from "@/lib/nav-config";
import { cn } from "@/lib/utils";

interface AppSidebarProps {
  role: UserRole;
  collapsed?: boolean;
}

function isNavActive(pathname: string, item: NavItem): boolean {
  if (item.exact) return pathname === item.href;
  return pathname === item.href || pathname.startsWith(item.href + "/");
}

function NavItemLink({ item, active }: { item: NavItem; active: boolean }) {
  return (
    <Link
      href={item.href}
      title={item.label}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
        active
          ? "bg-brand-50 font-semibold text-brand-700"
          : "font-medium text-ink-600 hover:bg-brand-50/60 hover:text-brand-700"
      )}
    >
      <item.icon
        className={cn(
          "size-5 shrink-0",
          active ? "text-brand-600" : "text-ink-500"
        )}
      />
      <span className="truncate">{item.label}</span>
      {item.comingSoon && (
        <Badge
          variant="outline"
          className="ml-auto px-1.5 py-0 text-[10px] font-normal text-ink-400"
        >
          Akan datang
        </Badge>
      )}
    </Link>
  );
}

function RailItemLink({ item, active }: { item: NavItem; active: boolean }) {
  return (
    <Link
      href={item.href}
      title={item.label}
      className={cn(
        "flex size-10 items-center justify-center rounded-lg transition-colors",
        active
          ? "bg-brand-50 text-brand-600"
          : "text-ink-500 hover:bg-brand-50/60 hover:text-brand-700"
      )}
    >
      <item.icon className="size-5" />
    </Link>
  );
}

export function AppSidebar({ role, collapsed = false }: AppSidebarProps) {
  const pathname = usePathname();
  const user = useAuthStore((s) => s.user);
  const logout = useLogout();
  const groups = useMemo(() => NAV_CONFIG[role] ?? [], [role]);
  const [open, setOpen] = useState(() => groups.map(() => true));

  const initial = useMemo(() => {
    const source = user?.name ?? user?.email ?? user?.username ?? "A";
    return source.trim().charAt(0).toUpperCase();
  }, [user]);

  const roleLabel = roleDisplayName(role ?? user?.role);

  function handleLogout() {
    logout.mutate(undefined);
  }

  function toggleGroup(index: number) {
    setOpen((prev) => {
      const next = [...prev];
      next[index] = !next[index];
      return next;
    });
  }

  const flatItems = useMemo(
    () => groups.flatMap((g) => g.items),
    [groups]
  );

  return (
    <>
      <div className="flex h-16 shrink-0 items-center justify-center px-4 lg:justify-start">
        <Link href="/" className="flex items-center gap-2">
          <AbakMark size={28} />
          <span className="hidden font-serif text-lg font-extrabold tracking-tight text-ink-900 lg:block">
            abak
            <span className="ml-1 text-[0.7em] uppercase tracking-[0.18em] text-gold">
              academy
            </span>
          </span>
        </Link>
      </div>

      <Separator className="shrink-0" />

      {collapsed ? (
        <nav className="flex flex-1 flex-col items-center gap-2 overflow-y-auto px-2 py-3">
          {flatItems.map((item) => (
            <RailItemLink
              key={item.href + item.label}
              item={item}
              active={isNavActive(pathname, item)}
            />
          ))}
        </nav>
      ) : (
        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-3 py-3">
          {groups.map((group, index) => (
            <div key={group.title || `group-${index}`}>
              {group.title && (
                <button
                  type="button"
                  onClick={() => toggleGroup(index)}
                  className="flex w-full items-center justify-between px-3 py-2 text-left"
                >
                  <span className="text-[10px] font-bold uppercase tracking-wider text-ink-400">
                    {group.title}
                  </span>
                  <ChevronDown
                    className={cn(
                      "size-4 text-ink-400 transition-transform",
                      open[index] && "rotate-180"
                    )}
                  />
                </button>
              )}
              <div
                className={cn(
                  "flex flex-col gap-1",
                  group.title && !open[index] && "hidden"
                )}
              >
                {group.items.map((item) => (
                  <NavItemLink
                    key={item.href + item.label}
                    item={item}
                    active={isNavActive(pathname, item)}
                  />
                ))}
              </div>
            </div>
          ))}
        </nav>
      )}

      <div className="mt-auto border-t border-line">
        {collapsed ? (
          <div className="flex flex-col items-center gap-2 p-2">
            <Avatar size="sm">
              <AvatarFallback className="bg-brand-600 text-xs font-semibold text-white">
                {initial}
              </AvatarFallback>
            </Avatar>
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={handleLogout}
              aria-label="Keluar"
            >
              <LogOut className="size-4 text-ink-500" />
            </Button>
          </div>
        ) : (
          <div className="flex items-center gap-3 p-4">
            <Avatar size="sm">
              <AvatarFallback className="bg-brand-600 text-xs font-semibold text-white">
                {initial}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-ink-900">
                {user?.name ?? "User"}
              </p>
              <Badge
                variant="secondary"
                className="mt-0.5 px-1.5 py-0 text-[10px] font-normal"
              >
                {roleLabel}
              </Badge>
            </div>
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={handleLogout}
              aria-label="Keluar"
            >
              <LogOut className="size-4 text-ink-500" />
            </Button>
          </div>
        )}
      </div>
    </>
  );
}
