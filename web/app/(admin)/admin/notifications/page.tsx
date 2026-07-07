"use client";

import { useState } from "react";
import { Bell } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { AnnouncementTable } from "@/components/admin/AnnouncementTable";
import { AnnouncementComposer } from "@/components/admin/AnnouncementComposer";
import { PurchaseNotificationFeed } from "@/components/admin/PurchaseNotificationFeed";
import type { Announcement } from "@/lib/hooks/admin-announcements";

export default function NotificationsPage() {
  const [composerOpen, setComposerOpen] = useState(false);
  const [editing, setEditing] = useState<Announcement | null>(null);

  const handleCreate = () => {
    setEditing(null);
    setComposerOpen(true);
  };

  const handleEdit = (ann: Announcement) => {
    setEditing(ann);
    setComposerOpen(true);
  };

  const handleClose = () => {
    setComposerOpen(false);
    setEditing(null);
  };

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={Bell}
        title="Notifikasi"
        description="Pantau notifikasi pembelian dan kelola pengumuman."
      />

      <section className="mx-auto max-w-6xl px-4 md:px-6">
        <PurchaseNotificationFeed />
      </section>

      <AnnouncementTable onCreateClick={handleCreate} onEdit={handleEdit} />

      <AnnouncementComposer
        open={composerOpen}
        onOpenChange={(open) => {
          if (!open) handleClose();
        }}
        announcement={editing}
      />
    </div>
  );
}
