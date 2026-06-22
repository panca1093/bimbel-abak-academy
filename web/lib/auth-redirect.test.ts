import { describe, it, expect } from "vitest";
import { redirectForRole, adminHomeForRole } from "./auth-redirect";
import type { UserRole } from "./nav-config";

describe("redirectForRole", () => {
  it("returns '/' for student", () => {
    expect(redirectForRole("student")).toBe("/");
  });

  it("returns '/admin' for every admin role", () => {
    const adminRoles: UserRole[] = [
      "admin_store",
      "admin_exam",
      "admin_school",
      "super_admin",
    ];
    for (const role of adminRoles) {
      expect(redirectForRole(role)).toBe("/admin");
    }
  });

  it("defaults to '/' for unknown or missing role", () => {
    expect(redirectForRole("unknown")).toBe("/");
    expect(redirectForRole(undefined)).toBe("/");
    expect(redirectForRole(null)).toBe("/");
  });
});

describe("adminHomeForRole", () => {
  it("sends admin_store to /admin/store", () => {
    expect(adminHomeForRole("admin_store")).toBe("/admin/store");
  });

  it("sends admin_exam to first coming-soon exam item", () => {
    expect(adminHomeForRole("admin_exam")).toBe("/admin/exam/banks");
  });

  it("sends admin_school to first coming-soon school item", () => {
    expect(adminHomeForRole("admin_school")).toBe("/admin/school/students");
  });

  it("sends super_admin to first live admin item", () => {
    expect(adminHomeForRole("super_admin")).toBe("/admin/store");
  });
});
