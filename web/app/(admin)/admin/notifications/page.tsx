"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Bell,
  Send,
  Clock,
  Users,
  Tag,
  MoreHorizontal,
  Edit,
  Trash2,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { DICT } from "@/lib/i18n";

type NotificationType = "announcement" | "promo" | "exam";
type NotificationStatus = "draft" | "scheduled" | "sent";
type RecipientGroup = "all" | "students" | "admins";

interface Notification {
  id: string;
  title: string;
  message: string;
  type: NotificationType;
  recipients: RecipientGroup;
  status: NotificationStatus;
  scheduledAt?: string;
  sentAt?: string;
  readCount?: number;
  totalCount?: number;
}

const INITIAL_NOTIFICATIONS: Notification[] = [
  {
    id: "NTF-1001",
    title: "Tryout SNBT #12 dibuka",
    message:
      "Sesi check-in untuk Tryout SNBT Nasional 2026 #12 sudah dibuka. Masukkan token di halaman kompetisi.",
    type: "exam",
    recipients: "students",
    status: "sent",
    sentAt: "2026-07-12T07:00",
    readCount: 742,
    totalCount: 843,
  },
  {
    id: "NTF-1002",
    title: "Diskon 30% Paket UTBK",
    message:
      "Dapatkan diskon 30% untuk pembelian Paket UTBK Premium hingga akhir bulan ini.",
    type: "promo",
    recipients: "all",
    status: "scheduled",
    scheduledAt: "2026-06-25T10:00",
  },
  {
    id: "NTF-1003",
    title: "Pemeliharaan sistem",
    message:
      "Platform akan mengalami pemeliharaan singkat pada Sabtu, 21 Juni 2026 pukul 02.00–04.00 WIB.",
    type: "announcement",
    recipients: "all",
    status: "draft",
  },
  {
    id: "NTF-1004",
    title: "Jadwal kelas intensif",
    message: "Kelas intensif Matematika dimulai Senin depan pukul 19.00 WIB.",
    type: "announcement",
    recipients: "students",
    status: "sent",
    sentAt: "2026-06-18T15:30",
    readCount: 89,
    totalCount: 1303,
  },
];

const NOTIFICATION_TYPES: NotificationType[] = ["announcement", "promo", "exam"];
const RECIPIENT_GROUPS: RecipientGroup[] = ["all", "students", "admins"];

const TYPE_TONE: Record<NotificationType, string> = {
  announcement: "bg-info-bg text-info border-info",
  promo: "bg-success-bg text-success border-success",
  exam: "bg-warn-bg text-warn border-warn",
};

const TYPE_ICON: Record<NotificationType, React.ReactNode> = {
  announcement: <Bell className="size-4" />,
  promo: <Tag className="size-4" />,
  exam: <Clock className="size-4" />,
};

type DictKey = keyof (typeof DICT)["id"];

const TYPE_LABEL: Record<NotificationType, DictKey> = {
  announcement: "notification_announcement",
  promo: "notification_promo",
  exam: "notification_exam",
};

const RECIPIENT_LABEL: Record<RecipientGroup, DictKey> = {
  all: "notification_all_users",
  students: "notification_students",
  admins: "notification_admins",
};

const STATUS_TONE: Record<NotificationStatus, string> = {
  draft: "bg-ink-100 text-ink-600 border-line",
  scheduled: "bg-info-bg text-info border-info",
  sent: "bg-success-bg text-success border-success",
};

function fmtDateTime(iso?: string) {
  if (!iso) return "—";
  const d = new Date(iso);
  return d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function NotificationsPage() {
  const { t } = useTranslation();
  const [tab, setTab] = useState<NotificationStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [draft, setDraft] = useState<Partial<Notification>>({
    type: "announcement",
    recipients: "all",
  });

  const rows = useMemo(() => {
    if (tab === "all") return INITIAL_NOTIFICATIONS;
    return INITIAL_NOTIFICATIONS.filter((n) => n.status === tab);
  }, [tab]);

  const stats = useMemo(() => {
    return {
      total: INITIAL_NOTIFICATIONS.length,
      sent: INITIAL_NOTIFICATIONS.filter((n) => n.status === "sent").length,
      scheduled: INITIAL_NOTIFICATIONS.filter((n) => n.status === "scheduled").length,
      draft: INITIAL_NOTIFICATIONS.filter((n) => n.status === "draft").length,
    };
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {t("notifications")}
          </h1>
          <p className="mt-2 text-sm text-ink-500">
            Buat, jadwalkan, dan pantau notifikasi ke pengguna.
          </p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1 size-4" />
          {t("create")}
        </Button>
      </header>

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card className="p-5">
          <div className="text-xs text-ink-500">Total notifikasi</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.total}</div>
        </Card>
        <Card className="p-5">
          <div className="text-xs text-ink-500">Terkirim</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.sent}</div>
        </Card>
        <Card className="p-5">
          <div className="text-xs text-ink-500">Terjadwal</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.scheduled}</div>
        </Card>
        <Card className="p-5">
          <div className="text-xs text-ink-500">Draft</div>
          <div className="mt-1 text-2xl font-bold text-ink-900">{stats.draft}</div>
        </Card>
      </div>

      <Tabs value={tab} onValueChange={(v) => setTab(v as typeof tab)}>
        <TabsList className="mb-4">
          <TabsTrigger value="all" className="text-xs">
            {t("tab_all")}
          </TabsTrigger>
          <TabsTrigger value="sent" className="text-xs">
            {t("notification_sent")}
          </TabsTrigger>
          <TabsTrigger value="scheduled" className="text-xs">
            {t("notification_scheduled")}
          </TabsTrigger>
          <TabsTrigger value="draft" className="text-xs">
            {t("notification_draft")}
          </TabsTrigger>
        </TabsList>
      </Tabs>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600">
              <tr>
                <th className="px-4 py-3">{t("notification")}</th>
                <th className="px-4 py-3">{t("notification_type")}</th>
                <th className="px-4 py-3">{t("notification_recipients")}</th>
                <th className="px-4 py-3">{t("notification_status")}</th>
                <th className="px-4 py-3">Waktu</th>
                <th className="px-4 py-3">Dibaca</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {rows.map((n) => (
                <tr key={n.id} className="group hover:bg-surface-2">
                  <td className="px-4 py-3">
                    <div className="max-w-[280px] truncate font-medium text-ink-900">
                      {n.title}
                    </div>
                    <div className="font-mono text-[11px] text-ink-500">{n.id}</div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold",
                        TYPE_TONE[n.type]
                      )}
                    >
                      <span className="mr-1">{TYPE_ICON[n.type]}</span>
                      {t(TYPE_LABEL[n.type])}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center gap-1 text-xs text-ink-600">
                      <Users className="size-3" />
                      {t(RECIPIENT_LABEL[n.recipients])}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[11px] font-semibold capitalize",
                        STATUS_TONE[n.status]
                      )}
                    >
                      {n.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {n.status === "sent"
                      ? fmtDateTime(n.sentAt)
                      : fmtDateTime(n.scheduledAt)}
                  </td>
                  <td className="px-4 py-3 text-xs text-ink-600">
                    {n.readCount != null && n.totalCount != null ? (
                      <>
                        {n.readCount.toLocaleString("id-ID")} /{" "}
                        {n.totalCount.toLocaleString("id-ID")}
                      </>
                    ) : (
                      "—"
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon-xs">
                          <MoreHorizontal className="size-4 text-ink-500" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => setCreateOpen(true)}>
                          <Edit className="mr-2 size-4" />
                          {t("update")}
                        </DropdownMenuItem>
                        {n.status !== "sent" && (
                          <DropdownMenuItem>
                            <Send className="mr-2 size-4" />
                            {t("notification_send_now")}
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem className="text-danger">
                          <Trash2 className="mr-2 size-4" />
                          {t("cancel")}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="font-serif">
              {t("create")} {t("notification")}
            </DialogTitle>
            <DialogDescription>
              Form pembuatan notifikasi akan tersedia di iterasi berikutnya.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>{t("notification_title")}</Label>
              <Input
                value={draft.title || ""}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, title: e.target.value }))
                }
                placeholder="Judul notifikasi"
              />
            </div>
            <div>
              <Label>{t("notification_message")}</Label>
              <textarea
                value={draft.message || ""}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setDraft((d) => ({ ...d, message: e.target.value }))
                }
                placeholder="Isi pesan…"
                rows={4}
                className={cn(
                  "min-h-[96px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none transition-[color,box-shadow]",
                  "focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-brand-300/50",
                  "placeholder:text-muted-foreground disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
                )}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>{t("notification_type")}</Label>
                <Select
                  value={draft.type}
                  onValueChange={(v) =>
                    setDraft((d) => ({ ...d, type: v as NotificationType }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {NOTIFICATION_TYPES.map((tt) => (
                      <SelectItem key={tt} value={tt}>
                        {t(TYPE_LABEL[tt])}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>{t("notification_recipients")}</Label>
                <Select
                  value={draft.recipients}
                  onValueChange={(v) =>
                    setDraft((d) => ({ ...d, recipients: v as RecipientGroup }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {RECIPIENT_GROUPS.map((rg) => (
                      <SelectItem key={rg} value={rg}>
                        {t(RECIPIENT_LABEL[rg])}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => setCreateOpen(false)}
              >
                {t("cancel")}
              </Button>
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                {t("save")} {t("notification_draft")}
              </Button>
              <Button onClick={() => setCreateOpen(false)}>
                <Send className="mr-1 size-4" />
                {t("send")}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
