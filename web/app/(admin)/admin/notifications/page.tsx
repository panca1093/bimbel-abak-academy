"use client";

import { useState } from "react";
import { Bell } from "lucide-react";
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
    <>
      <AnnouncementTable onCreateClick={handleCreate} onEdit={handleEdit} />
      <AnnouncementComposer
        open={composerOpen}
        onOpenChange={(open) => {
          if (!open) handleClose();
        }}
        announcement={editing}
      />
      <section className="mx-auto max-w-6xl px-4 pb-12 md:px-6">
        <PurchaseNotificationFeed />
      </section>
    </>
  );
}
