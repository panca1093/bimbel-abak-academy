import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { FC } from "react";
import NotificationsPage from "./notifications/page";
import ExamTryoutsPage from "./exam/tryouts/page";
import ExamAnalyticsPage from "./exam/analytics/page";
import ExamSchedulesPage from "./exam/schedules/page";
import SchoolStudentsPage from "./school/students/page";
import SchoolClassesPage from "./school/classes/page";
import SchoolReportsPage from "./school/reports/page";
import SystemAccountsPage from "./system/accounts/page";
import SystemSchoolsPage from "./system/schools/page";
import SystemConfigPage from "./system/config/page";
import SystemAuditPage from "./system/audit/page";
import { NAV_CONFIG, type UserRole } from "@/lib/nav-config";

vi.mock("@/components/shell/ComingSoon", () => ({
  ComingSoon: ({ title }: { title?: string }) => (
    <div data-testid="coming-soon" data-title={title}>
      ComingSoon
    </div>
  ),
}));

const routeCases: [string, FC][] = [
  ["/admin/notifications", NotificationsPage],
  ["/admin/exam/tryouts", ExamTryoutsPage],
  ["/admin/exam/analytics", ExamAnalyticsPage],
  ["/admin/exam/schedules", ExamSchedulesPage],
  ["/admin/school/students", SchoolStudentsPage],
  ["/admin/school/classes", SchoolClassesPage],
  ["/admin/school/reports", SchoolReportsPage],
  ["/admin/system/accounts", SystemAccountsPage],
  ["/admin/system/schools", SystemSchoolsPage],
  ["/admin/system/config", SystemConfigPage],
  ["/admin/system/audit", SystemAuditPage],
];

describe("Coming-soon admin routes", () => {
  it.each(routeCases)("%s renders ComingSoon", (path, Page) => {
    render(<Page />);
    expect(screen.getByTestId("coming-soon")).toBeInTheDocument();
  });
});

describe("nav-config coming-soon flags", () => {
  it("flags every admin_exam and admin_school item as coming-soon", () => {
    for (const role of ["admin_exam", "admin_school"] as UserRole[]) {
      for (const group of NAV_CONFIG[role]) {
        for (const item of group.items) {
          expect(item.comingSoon).toBe(true);
        }
      }
    }
  });

  it("flags super_admin Exam, School, and System items as coming-soon", () => {
    const superAdminGroups = NAV_CONFIG.super_admin;
    for (const group of superAdminGroups) {
      if (group.titleKey === "role_admin_store") continue;
      for (const item of group.items) {
        expect(item.comingSoon).toBe(true);
      }
    }
  });

  it("marks Content Manager Notifications as coming-soon", () => {
    const cmItems = NAV_CONFIG.admin_store[0].items;
    const notifications = cmItems.find((i) => i.href === "/admin/notifications");
    expect(notifications?.comingSoon).toBe(true);
  });

  it("never flags live Content Manager items as coming-soon", () => {
    const liveHrefs = ["/admin", "/admin/products", "/admin/courses", "/admin/orders", "/admin/promos", "/admin/revenue"];
    for (const role of ["admin_store", "super_admin"] as UserRole[]) {
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
