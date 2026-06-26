"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Building,
  Users,
  MapPin,
  MoreHorizontal,
  Edit,
  Trash2,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";

type SchoolStatus = "active" | "inactive" | "pending";

interface SchoolOrg {
  id: string;
  name: string;
  city: string;
  province: string;
  status: SchoolStatus;
  students: number;
  classes: number;
  exams: number;
}

const INITIAL_SCHOOLS: SchoolOrg[] = [
  {
    id: "SCH-501",
    name: "SMAN 1 Jakarta",
    city: "Jakarta",
    province: "DKI Jakarta",
    status: "active",
    students: 142,
    classes: 6,
    exams: 24,
  },
  {
    id: "SCH-502",
    name: "SMAN 3 Bandung",
    city: "Bandung",
    province: "Jawa Barat",
    status: "active",
    students: 98,
    classes: 4,
    exams: 18,
  },
  {
    id: "SCH-503",
    name: "SMAN 2 Surabaya",
    city: "Surabaya",
    province: "Jawa Timur",
    status: "active",
    students: 76,
    classes: 3,
    exams: 15,
  },
  {
    id: "SCH-504",
    name: "SMAN 5 Yogyakarta",
    city: "Yogyakarta",
    province: "DI Yogyakarta",
    status: "pending",
    students: 0,
    classes: 0,
    exams: 0,
  },
];

const STATUS_TONE: Record<SchoolStatus, string> = {
  active: "bg-success-bg text-success border-success",
  inactive: "bg-ink-100 text-ink-600 border-line",
  pending: "bg-warn-bg text-warn border-warn",
};

export default function SystemSchoolsPage() {
  const { t, lang } = useTranslation();
  const numberLocale = lang === "en" ? "en-US" : "id-ID";
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<SchoolStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);

  const rows = useMemo(() => {
    return INITIAL_SCHOOLS.filter((s) => {
      const q = search.toLowerCase();
      const matchesSearch =
        search.trim() === "" ||
        s.name.toLowerCase().includes(q) ||
        s.city.toLowerCase().includes(q) ||
        s.province.toLowerCase().includes(q);
      const matchesStatus = statusFilter === "all" || s.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [search, statusFilter]);

  const stats = useMemo(() => {
    const students = INITIAL_SCHOOLS.reduce((sum, s) => sum + s.students, 0);
    const active = INITIAL_SCHOOLS.filter((s) => s.status === "active").length;
    return {
      total: INITIAL_SCHOOLS.length,
      students,
      active,
    };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={Building}
        title={t("schools_title")}
        description={t("schools_subtitle")}
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        }
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard label={t("schools_stat_total")} value={String(stats.total)} />
        <StatCard label={t("status_label_active")} value={String(stats.active)} />
        <StatCard label={t("schools_stat_students")} value={stats.students.toLocaleString(numberLocale)} />
      </div>

      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap gap-2">
          <FilterChip active={statusFilter === "all"} onClick={() => setStatusFilter("all")}>
            {t("tab_all")}
          </FilterChip>
          <FilterChip active={statusFilter === "active"} onClick={() => setStatusFilter("active")}>
            {t("status_label_active")}
          </FilterChip>
          <FilterChip active={statusFilter === "pending"} onClick={() => setStatusFilter("pending")}>
            {t("schools_status_pending")}
          </FilterChip>
          <FilterChip active={statusFilter === "inactive"} onClick={() => setStatusFilter("inactive")}>
            {t("status_label_inactive")}
          </FilterChip>
        </div>
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t("schools_search_placeholder")}
          className="h-9 w-[260px] text-xs sm:ml-auto"
        />
      </div>

      <div className="md-card-outlined">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">{t("schools_th_school")}</th>
                <th className="px-4 py-3">{t("schools_th_location")}</th>
                <th className="px-4 py-3">{t("accounts_th_status")}</th>
                <th className="px-4 py-3">{t("schools_th_students")}</th>
                <th className="px-4 py-3">{t("schools_th_classes")}</th>
                <th className="px-4 py-3">{t("schools_th_exams")}</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((s) => (
                <tr key={s.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 font-medium text-ink-900">
                      <Building className="size-4 text-brand-600" />
                      {s.name}
                    </div>
                    <div className="font-mono text-[11px] text-ink-500">{s.id}</div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 text-xs text-ink-700">
                      <MapPin className="size-3.5 text-ink-400" />
                      {s.city}, {s.province}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn("text-[11px] font-semibold capitalize", STATUS_TONE[s.status])}
                    >
                      {s.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    <span className="inline-flex items-center gap-1">
                      <Users className="size-3" />
                      {s.students.toLocaleString(numberLocale)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">{s.classes}</td>
                  <td className="px-4 py-3 text-xs text-ink-600">{s.exams}</td>
                  <td className="px-4 py-3 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon-xs">
                          <MoreHorizontal className="size-4 text-ink-500" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => setCreateOpen(true)}>
                          <Edit className="mr-2 size-4" />{t("schools_action_edit")}
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-danger">
                          <Trash2 className="mr-2 size-4" />{t("schools_action_delete")}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">{t("schools_dialog_create_title")}</DialogTitle>
            <DialogDescription>{t("schools_dialog_create_desc")}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>{t("schools_field_name")}</Label>
              <Input placeholder={t("schools_placeholder_name")} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>{t("schools_field_city")}</Label>
                <Input placeholder={t("schools_field_city")} />
              </div>
              <div>
                <Label>{t("schools_field_province")}</Label>
                <Input placeholder={t("schools_field_province")} />
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                {t("cancel")}
              </Button>
              <Button onClick={() => setCreateOpen(false)}>{t("create")}</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-lg border px-3 py-[7px] text-xs font-semibold transition-colors",
        active
          ? "border-brand-600 bg-brand-600 text-white"
          : "border-line bg-surface text-ink-600 hover:text-ink-900"
      )}
    >
      {children}
    </button>
  );
}
