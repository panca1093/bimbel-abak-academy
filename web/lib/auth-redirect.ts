import { ADMIN_ROLES, NAV_CONFIG, type UserRole } from "./nav-config";

export function redirectForRole(role?: string | null): string {
  if (ADMIN_ROLES.includes(role as UserRole)) return "/admin";
  return "/";
}

export function adminHomeForRole(role: UserRole): string {
  if (role === "admin_store") return "/admin/products";
  const items = NAV_CONFIG[role]?.flatMap((group) => group.items) ?? [];
  const first = items.find((item) => item.href !== "/admin");
  return first?.href ?? "/admin/products";
}
