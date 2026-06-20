"use client";

import { useState } from "react";
import { useProducts } from "@/lib/hooks/products";
import type { ProductType } from "@/lib/types";
import { useTranslation } from "@/lib/i18n";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { ProductCard } from "@/components/catalog/ProductCard";

type TabValue = "all" | ProductType;

const TABS: { value: TabValue; labelKey: string }[] = [
  { value: "all", labelKey: "catalog_tab_all" },
  { value: "book", labelKey: "catalog_tab_book" },
  { value: "course", labelKey: "catalog_tab_course" },
  { value: "package", labelKey: "catalog_tab_competition" },
];

function CatalogGrid({ products }: { products: ReturnType<typeof useProducts>["data"] }) {
  const { t } = useTranslation();
  if (!products || products.length === 0) {
    return (
      <p className="py-16 text-center text-sm text-ink-500">
        {t("catalog_empty")}
      </p>
    );
  }
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {products.map((p) => (
        <ProductCard key={p.id} product={p} />
      ))}
    </div>
  );
}

function CatalogSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex flex-col overflow-hidden rounded-lg border border-line bg-surface">
          <Skeleton className="h-32 rounded-none" />
          <div className="flex flex-col gap-2 p-4">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-full" />
            <Skeleton className="mt-2 h-5 w-1/3" />
          </div>
        </div>
      ))}
    </div>
  );
}

export default function CatalogPage() {
  const { t } = useTranslation();
  const [tab, setTab] = useState<TabValue>("all");
  const type = tab === "all" ? undefined : tab;
  const { data, isLoading, isError, error, refetch } = useProducts(type);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-6">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">{t("catalog_title")}</h1>
        <p className="mt-1 text-sm text-ink-500">
          {t("catalog_subtitle")}
        </p>
      </header>

      <Tabs
        value={tab}
        onValueChange={(v) => setTab(v as TabValue)}
        className="mb-6"
      >
        <TabsList variant="line">
          {TABS.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value}>
              {t(tab.labelKey as any)}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {isError ? (
        <div className="rounded-lg border border-danger/30 bg-danger-bg px-5 py-4 text-sm text-danger">
          <p>{t("catalog_load_failed")} {(error as Error)?.message}</p>
          <button onClick={() => refetch()} className="mt-2 underline">
            {t("retry")}
          </button>
        </div>
      ) : isLoading ? (
        <CatalogSkeleton />
      ) : (
        <CatalogGrid products={data} />
      )}
    </div>
  );
}