"use client";

import { useEffect, useState, useCallback } from "react";
import { useAdminNotifications, useMarkNotificationRead } from "@/lib/hooks/admin-notifications";
import { useTranslation } from "@/lib/i18n";
import type { AdminNotification } from "@/lib/hooks/admin-notifications";

function formatAmount(amount: number): string {
  return `Rp${amount.toLocaleString("id-ID")}`;
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function PurchaseNotificationFeed() {
  const { t } = useTranslation();
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [cursor, setCursor] = useState<string | undefined>();
  const [allItems, setAllItems] = useState<AdminNotification[]>([]);
  const [hasMore, setHasMore] = useState(true);

  const query = useAdminNotifications({ unreadOnly, cursor });
  const markRead = useMarkNotificationRead();

  useEffect(() => {
    if (!query.data) return;
    const items = query.data.data ?? [];
    if (!cursor) {
      setAllItems(items);
    } else {
      setAllItems((prev) => [...prev, ...items]);
    }
    setHasMore(Boolean(query.data.next_cursor));
  }, [query.data, cursor]);

  const handleToggleUnreadOnly = useCallback(() => {
    setUnreadOnly((prev) => !prev);
    setCursor(undefined);
    setAllItems([]);
    setHasMore(true);
  }, []);

  const handleLoadMore = useCallback(() => {
    if (query.data?.next_cursor) {
      setCursor(query.data.next_cursor);
    }
  }, [query.data]);

  const handleMarkRead = useCallback(
    (id: string) => {
      markRead.mutate(id);
    },
    [markRead]
  );

  if (query.isFetching && !query.data) {
    return <div className="p-4 text-center color-on-surface-variant">Memuat...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-title-medium">{t("notification")}</h2>
        <button
          type="button"
          onClick={handleToggleUnreadOnly}
          className={`text-label rounded-md px-3 py-1.5 transition-colors ${
            unreadOnly
              ? "bg-primary text-on-primary"
              : "bg-surface-container-high text-on-surface-variant"
          }`}
          aria-label={t("notification_unread_only")}
        >
          {t("notification_unread_only")}
        </button>
      </div>

      {allItems.length === 0 && !query.isFetching ? (
        <p className="p-4 text-center color-on-surface-variant text-body-medium">
          {"Belum ada notifikasi"}
        </p>
      ) : (
        <ul className="divide-y divide-outline-variant">
          {allItems.map((notif) => (
            <li
              key={notif.id}
              onClick={() => handleMarkRead(notif.id)}
              className={`flex cursor-pointer items-center justify-between gap-4 px-4 py-3 transition-colors hover:bg-surface-container-high ${
                !notif.read ? "bg-primary-container/20" : ""
              }`}
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="truncate text-body-medium font-medium">
                    {notif.student_name}
                  </p>
                  {!notif.read && (
                    <span className="inline-block size-2 shrink-0 rounded-full bg-primary" />
                  )}
                </div>
                <p className="text-body-small color-on-surface-variant">
                  {"Pesanan"}: {notif.order_id.slice(0, 8)}...
                </p>
                <p className="text-body-small color-on-surface-variant">
                  {formatTimestamp(notif.created_at)}
                </p>
              </div>
              <div className="shrink-0 text-right">
                <p className="text-body-medium font-semibold">{formatAmount(notif.amount)}</p>
              </div>
            </li>
          ))}
        </ul>
      )}

      {hasMore && (
        <div className="flex justify-center pt-2">
          <button
            type="button"
            onClick={handleLoadMore}
            disabled={query.isFetching}
            className="text-label rounded-md bg-surface-container-high px-4 py-2 color-on-surface-variant transition-colors hover:bg-surface-container-highest disabled:opacity-50"
          >
            {query.isFetching ? "Memuat..." : "Muat lebih banyak"}
          </button>
        </div>
      )}
    </div>
  );
}
