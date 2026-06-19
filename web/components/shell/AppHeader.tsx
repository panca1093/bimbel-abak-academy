"use client";

import { useRouter } from "next/navigation";
import {
  Menu,
  Search,
  Globe,
  Moon,
  Bell,
  LogOut,
} from "lucide-react";
import { useAuthStore } from "@/stores/auth";
import { useLogout } from "@/lib/hooks/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { roleDisplayName } from "@/lib/nav-config";
import { cn } from "@/lib/utils";

interface AppHeaderProps {
  onMenuClick?: () => void;
}

export function AppHeader({ onMenuClick }: AppHeaderProps) {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const logout = useLogout();

  const initial = (user?.name ?? user?.email ?? user?.username ?? "A")
    .trim()
    .charAt(0)
    .toUpperCase();
  const roleLabel = roleDisplayName(user?.role);

  function handleLogout() {
    logout.mutate(undefined, {
      onSettled: () => {
        router.replace("/login");
      },
    });
  }

  return (
    <header className="sticky top-0 z-20 flex h-16 shrink-0 items-center gap-3 border-b border-line bg-surface/90 px-4 backdrop-blur lg:px-8">
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={onMenuClick}
        aria-label="Toggle sidebar"
        className="lg:hidden"
      >
        <Menu className="size-5 text-ink-600" />
      </Button>

      <div className="flex flex-1 items-center gap-2">
        <Search className="size-4 shrink-0 text-ink-400" />
        <Input
          type="search"
          placeholder="Cari menu atau materi..."
          className="h-9 w-full max-w-md"
        />
      </div>

      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          aria-label="Ganti bahasa"
          className="hidden sm:flex"
        >
          <Globe className="size-5 text-ink-600" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          aria-label="Mode gelap"
          className="hidden sm:flex"
        >
          <Moon className="size-5 text-ink-600" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="relative"
          aria-label="Notifikasi"
        >
          <Bell className="size-5 text-ink-600" />
          <span className="absolute right-1.5 top-1.5 size-2 rounded-full bg-danger" />
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className={cn(
                "flex items-center gap-2 px-2",
                "hidden md:flex"
              )}
            >
              <Avatar size="sm">
                <AvatarFallback className="bg-brand-600 text-xs font-semibold text-white">
                  {initial}
                </AvatarFallback>
              </Avatar>
              <span className="max-w-[120px] truncate text-sm font-medium text-ink-900">
                {user?.name ?? "Akun"}
              </span>
              <Badge
                variant="secondary"
                className="text-[10px] font-normal"
              >
                {roleLabel}
              </Badge>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel className="truncate text-ink-900">
              {user?.name ?? "Akun"}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut className="size-4" />
              <span className="ml-2">Keluar</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
