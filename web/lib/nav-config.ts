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
  label: string;
  href: string;
  icon: LucideIcon;
  exact?: boolean;
  comingSoon?: boolean;
}

export interface NavGroup {
  title: string;
  items: NavItem[];
}

export type RoleNavConfig = NavGroup[];

export const ROLE_LABELS: Record<UserRole, string> = {
  student: "Student",
  admin_store: "Content Manager",
  admin_exam: "Admin Exam",
  admin_school: "School Operator",
  super_admin: "Super Admin",
};

export const ADMIN_ROLES: UserRole[] = [
  "admin_store",
  "admin_exam",
  "admin_school",
  "super_admin",
];

function cs(label: string, href: string, icon: LucideIcon): NavItem {
  return { label, href, icon, comingSoon: true };
}

const STUDENT_NAV: RoleNavConfig = [
  {
    title: "",
    items: [
      { label: "Dashboard", href: "/", icon: LayoutDashboard, exact: true },
      cs("Competition", "/competition", Trophy),
      { label: "Courses", href: "/courses", icon: BookOpen },
      { label: "Store", href: "/catalog", icon: ShoppingBag },
      { label: "Billing/Orders", href: "/orders", icon: Receipt },
      { label: "Profile", href: "/profile", icon: User },
    ],
  },
];

const CONTENT_MANAGER_NAV: RoleNavConfig = [
  {
    title: "Content Manager",
    items: [
      { label: "Dashboard", href: "/admin", icon: LayoutDashboard, exact: true },
      { label: "Products", href: "/admin/products", icon: Package },
      { label: "Course Builder", href: "/admin/courses", icon: Library },
      { label: "Orders", href: "/admin/orders", icon: Receipt },
      { label: "Promos", href: "/admin/promos", icon: Tag },
      { label: "Revenue", href: "/admin/revenue", icon: BarChart3 },
      cs("Notifications", "/admin/notifications", Bell),
    ],
  },
];

const ADMIN_EXAM_NAV: RoleNavConfig = [
  {
    title: "Admin Exam",
    items: [
      cs("Question Banks", "/admin/exam/banks", FileQuestion),
      cs("Tryouts", "/admin/exam/tryouts", ClipboardList),
      cs("Analytics", "/admin/exam/analytics", BarChart),
      cs("Schedules", "/admin/exam/schedules", Calendar),
    ],
  },
];

const ADMIN_SCHOOL_NAV: RoleNavConfig = [
  {
    title: "School Operator",
    items: [
      cs("Students", "/admin/school/students", Users),
      cs("Classes", "/admin/school/classes", School),
      cs("Reports", "/admin/school/reports", FileText),
    ],
  },
];

const SUPER_ADMIN_NAV: RoleNavConfig = [
  ...CONTENT_MANAGER_NAV,
  ...ADMIN_EXAM_NAV,
  ...ADMIN_SCHOOL_NAV,
  {
    title: "System",
    items: [
      cs("Accounts", "/admin/system/accounts", Users),
      cs("Schools", "/admin/system/schools", Building),
      cs("Config", "/admin/system/config", Settings),
      cs("Audit", "/admin/system/audit", ShieldCheck),
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

export function roleDisplayName(role?: string): string {
  if (!role) return "User";
  return ROLE_LABELS[role as UserRole] ?? role;
}
