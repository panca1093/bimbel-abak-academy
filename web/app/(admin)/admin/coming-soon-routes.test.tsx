import { describe, it, expect } from "vitest";
import {
  ADMIN_ROLES,
  NAV_CONFIG,
  type NavGroup,
  type NavItem,
} from "@/lib/nav-config";

describe("coming soon routes", () => {
  it("should have no remaining coming-soon admin routes", () => {
    const isNavItem = (item: NavItem | NavGroup): item is NavItem =>
      "labelKey" in item;

    const isComingSoon = (item: NavItem | NavGroup): boolean => {
      if (isNavItem(item)) return item.comingSoon ?? false;
      return item.items.some(isComingSoon);
    };

    const adminGroups = ADMIN_ROLES.map((role) => NAV_CONFIG[role]).flat();
    expect(adminGroups.some(isComingSoon)).toBe(false);
  });
});
