"use client";

import Link from "next/link";
import { Store, Receipt, Package, BarChart3, Clock, CheckCircle2 } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { StatCard } from "@/components/admin/StatCard";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminOrders } from "@/lib/hooks/admin-orders";
import { useAdminRevenue } from "@/lib/hooks/admin-revenue";
import { formatRupiah } from "@/lib/format";
import { useTranslation } from "@/lib/i18n";

export default function StoreDashboardPage() {
  const { t } = useTranslation();
  const { data: pendingOrders, isLoading: loadingPending } = useAdminOrders("pending");
  const { data: processingOrders, isLoading: loadingProcessing } = useAdminOrders("processing");
  const { data: revenue, isLoading: loadingRevenue } = useAdminRevenue();

  const quickActions = [
    { icon: Package, label: t("store_action_manage_products"), href: "/admin/products" },
    { icon: Receipt, label: t("store_action_view_orders"), href: "/admin/orders" },
    { icon: BarChart3, label: t("store_action_revenue_report"), href: "/admin/revenue" },
  ];

  return (
    <div className="space-y-8 fade-in">
      <AdminPageHeader
        icon={Store}
        title={t("store_title")}
        description={t("store_subtitle")}
      />

      {/* Stats */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-3">
        {loadingPending ? (
          <Skeleton className="h-28 w-full" />
        ) : (
          <StatCard
            label={t("store_stat_pending")}
            value={String(pendingOrders?.length ?? 0)}
            trend={t("store_stat_pending_trend")}
            accent="error"
            icon={Clock}
          />
        )}
        {loadingProcessing ? (
          <Skeleton className="h-28 w-full" />
        ) : (
          <StatCard
            label={t("store_stat_processing")}
            value={String(processingOrders?.length ?? 0)}
            trend={t("store_stat_processing_trend")}
            accent="secondary"
            icon={CheckCircle2}
          />
        )}
        {loadingRevenue ? (
          <Skeleton className="h-28 w-full" />
        ) : (
          <StatCard
            label={t("store_stat_revenue_30d")}
            value={formatRupiah(revenue?.total ?? 0)}
            trend={t("store_stat_revenue_30d_trend")}
            accent="primary"
            icon={BarChart3}
          />
        )}
      </div>

      {/* Quick actions */}
      <div className="md-card-outlined">
        <h3 className="text-title-large mb-6">{t("admin_home_quick_actions")}</h3>
        <div className="grid gap-3 sm:grid-cols-3">
          {quickActions.map((action) => (
            <Link
              key={action.href}
              href={action.href}
              className="flex items-center gap-3 p-4 rounded-[12px] border border-line hover:bg-surface-container transition-colors"
            >
              <div
                className="flex size-10 items-center justify-center rounded-[10px]"
                style={{
                  backgroundColor: "var(--md-sys-color-primary-container)",
                  color: "var(--md-sys-color-primary)",
                }}
              >
                <action.icon size={20} />
              </div>
              <span className="text-body font-medium">{action.label}</span>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
