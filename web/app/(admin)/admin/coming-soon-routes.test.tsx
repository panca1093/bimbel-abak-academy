import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { FC } from "react";
import NotificationsPage from "./notifications/page";
import { NAV_CONFIG } from "@/lib/nav-config";

vi.mock("@/components/shell/ComingSoon", () => ({
  ComingSoon: ({ title }: { title?: string }) => (
    <div data-testid="coming-soon" data-title={title}>
      ComingSoon
    </div>
  ),
}));

const routeCases: [string, FC][] = [
  ["/admin/notifications", NotificationsPage],
];

describe("Coming-soon admin routes", () => {
  it.each(routeCases)("%s renders ComingSoon", (path, Page) => {
    render(<Page />);
    expect(screen.getByTestId("coming-soon")).toBeInTheDocument();
  });
});

describe("nav-config coming-soon flags", () => {
  it("flags Notifications as the only remaining coming-soon item", () => {
    const allItems = [
      ...NAV_CONFIG.admin_exam.flatMap((g) => g.items),
      ...NAV_CONFIG.admin_school.flatMap((g) => g.items),
      ...NAV_CONFIG.super_admin
        .filter((g) => g.titleKey === "system")
        .flatMap((g) => g.items),
      ...NAV_CONFIG.admin_store[0].items,
    ];
    const onlyComingSoon = allItems.filter((i) => i.comingSoon);
    expect(onlyComingSoon).toHaveLength(1);
    expect(onlyComingSoon[0].href).toBe("/admin/notifications");
  });

  it("never flags live admin items as coming-soon", () => {
    const liveHrefs = [
      "/admin",
      "/admin/products",
      "/admin/courses",
      "/admin/orders",
      "/admin/promos",
      "/admin/revenue",
      "/admin/exam/banks",
      "/admin/exam/tryouts",
      "/admin/exam/analytics",
      "/admin/exam/schedules",
      "/admin/school/students",
      "/admin/school/classes",
      "/admin/school/reports",
      "/admin/system/accounts",
      "/admin/system/schools",
      "/admin/system/config",
      "/admin/system/audit",
    ];
    for (const role of ["admin_store", "admin_exam", "admin_school", "super_admin"] as const) {
      for (const group of NAV_CONFIG[role]) {
        for (const item of group.items) {
          if (liveHrefs.includes(item.href)) {
            expect(item.comingSoon).not.toBe(true);
          }
        }
      }
    }
  });
});
