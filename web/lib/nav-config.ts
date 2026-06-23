import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  Trophy,
  BookOpen,
  ShoppingBag,
  Receipt,
  User,
  Package,
  Library,
  Tag,
  BarChart3,
  Bell,
  FileQuestion,
  ClipboardList,
  BarChart,
  Users,
  School,
  Calendar,
  FileText,
  Building,
  Settings,
  ShieldCheck,
} from "lucide-react";

export type UserRole =
  | "student"
  | "admin_store"
  | "admin_exam"
  | "admin_school"
  | "super_admin";

export interface NavItem {
  labelKey: keyof (typeof import("./i18n").DICT)["id"];
  href: string;
  icon: LucideIcon;
  exact?: boolean;
  comingSoon?: boolean;
}

export interface NavGroup {
  titleKey?: keyof (typeof import("./i18n").DICT)["id"];
  items: NavItem[];
}

export type RoleNavConfig = NavGroup[];

export const ROLE_LABEL_KEYS: Record<UserRole, keyof (typeof import("./i18n").DICT)["id"]> = {
  student: "role_student",
  admin_store: "role_admin_store",
  admin_exam: "role_admin_exam",
  admin_school: "role_admin_school",
  super_admin: "role_super_admin",
};

export const ADMIN_ROLES: UserRole[] = [
  "admin_store",
  "admin_exam",
  "admin_school",
  "super_admin",
];

function cs(
  labelKey: keyof (typeof import("./i18n").DICT)["id"],
  href: string,
  icon: LucideIcon
): NavItem {
  return { labelKey, href, icon, comingSoon: true };
}

const STUDENT_NAV: RoleNavConfig = [
  {
    items: [
      { labelKey: "nav_dashboard", href: "/", icon: LayoutDashboard, exact: true },
      { labelKey: "nav_competition", href: "/competition", icon: Trophy },
      { labelKey: "nav_courses", href: "/courses", icon: BookOpen },
      { labelKey: "nav_store", href: "/catalog", icon: ShoppingBag },
      { labelKey: "nav_billing", href: "/orders", icon: Receipt },
      { labelKey: "nav_profile", href: "/profile", icon: User },
    ],
  },
];

const CONTENT_MANAGER_NAV: RoleNavConfig = [
  {
    titleKey: "role_admin_store",
    items: [
      { labelKey: "nav_dashboard", href: "/admin/store", icon: LayoutDashboard, exact: true },
      { labelKey: "admin_nav_products", href: "/admin/products", icon: Package },
      { labelKey: "admin_nav_courses", href: "/admin/courses", icon: Library },
      { labelKey: "admin_nav_orders", href: "/admin/orders", icon: Receipt },
      { labelKey: "promos", href: "/admin/promos", icon: Tag },
      { labelKey: "revenue", href: "/admin/revenue", icon: BarChart3 },
      { labelKey: "notifications", href: "/admin/notifications", icon: Bell },
    ],
  },
];

const ADMIN_EXAM_NAV: RoleNavConfig = [
  {
    titleKey: "role_admin_exam",
    items: [
      { labelKey: "question_bank", href: "/admin/exam/banks", icon: FileQuestion },
      { labelKey: "tests", href: "/admin/exam/tryouts", icon: ClipboardList },
      { labelKey: "analytics", href: "/admin/exam/analytics", icon: BarChart },
      { labelKey: "schedules", href: "/admin/exam/schedules", icon: Calendar },
    ],
  },
];

const ADMIN_SCHOOL_NAV: RoleNavConfig = [
  {
    titleKey: "role_admin_school",
    items: [
      { labelKey: "students", href: "/admin/school/students", icon: Users },
      { labelKey: "classes", href: "/admin/school/classes", icon: School },
      { labelKey: "reports", href: "/admin/school/reports", icon: FileText },
    ],
  },
];

const SUPER_ADMIN_NAV: RoleNavConfig = [
  {
    items: [
      { labelKey: "nav_dashboard", href: "/admin", icon: LayoutDashboard, exact: true },
    ],
  },
  ...CONTENT_MANAGER_NAV,
  ...ADMIN_EXAM_NAV,
  ...ADMIN_SCHOOL_NAV,
  {
    titleKey: "system",
    items: [
      { labelKey: "accounts", href: "/admin/system/accounts", icon: Users },
      { labelKey: "schools", href: "/admin/system/schools", icon: Building },
      { labelKey: "config", href: "/admin/system/config", icon: Settings },
      { labelKey: "audit", href: "/admin/system/audit", icon: ShieldCheck },
    ],
  },
];

export const NAV_CONFIG: Record<UserRole, RoleNavConfig> = {
  student: STUDENT_NAV,
  admin_store: CONTENT_MANAGER_NAV,
  admin_exam: ADMIN_EXAM_NAV,
  admin_school: ADMIN_SCHOOL_NAV,
  super_admin: SUPER_ADMIN_NAV,
};

export function roleLabelKey(role?: string): keyof (typeof import("./i18n").DICT)["id"] | undefined {
  if (!role) return undefined;
  return ROLE_LABEL_KEYS[role as UserRole];
}
