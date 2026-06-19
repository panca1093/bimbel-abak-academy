"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "./AppSidebar";
import { AppHeader } from "./AppHeader";
import { cn } from "@/lib/utils";
import type { UserRole } from "@/lib/nav-config";

function useIsLg() {
  const [lg, setLg] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(min-width: 1024px)");
    function handler(event: MediaQueryListEvent | MediaQueryList) {
      setLg(event.matches);
    }
    handler(mq);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return lg;
}

interface AppShellProps {
  role: UserRole;
  children: React.ReactNode;
}

export function AppShell({ role, children }: AppShellProps) {
  const [expanded, setExpanded] = useState(false);
  const isLg = useIsLg();
  const collapsed = !isLg && !expanded;

  useEffect(() => {
    if (isLg) setExpanded(false);
  }, [isLg]);

  function toggleSidebar() {
    setExpanded((prev) => !prev);
  }

  return (
    <div className="flex min-h-screen bg-paper">
      <aside
        className={cn(
          "fixed left-0 top-0 z-30 flex h-screen flex-col border-r border-line bg-surface transition-all duration-200 ease-out",
          collapsed ? "w-16" : "w-[252px]"
        )}
      >
        <AppSidebar role={role} collapsed={collapsed} />
      </aside>

      <div
        className={cn(
          "flex min-w-0 flex-1 flex-col transition-[padding] duration-200 ease-out",
          collapsed ? "pl-16" : "pl-[252px]"
        )}
      >
        <AppHeader onMenuClick={toggleSidebar} />
        <main className="flex-1">
          <div className="mx-auto w-full max-w-[1340px] px-7 py-7 lg:px-8 lg:py-8">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
