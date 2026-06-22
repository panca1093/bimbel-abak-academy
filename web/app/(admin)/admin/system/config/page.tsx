"use client";

import { useEffect, useRef, useState } from "react";
import { Save, Settings, Bell, Shield, CreditCard, Mail } from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import {
  useAdminSystemConfig,
  useUpdateSystemConfig,
} from "@/lib/hooks/admin-config";

interface ToggleProps {
  checked: boolean;
  onChange: (v: boolean) => void;
  label: string;
  description?: string;
}

function Toggle({ checked, onChange, label, description }: ToggleProps) {
  return (
    <div className="flex items-start justify-between gap-4 py-3">
      <div className="space-y-0.5">
        <div className="text-sm font-medium text-ink-900">{label}</div>
        {description && (
          <div className="text-xs text-ink-500">{description}</div>
        )}
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={cn(
          "relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors",
          checked ? "bg-brand-600" : "bg-line"
        )}
      >
        <span
          className={cn(
            "pointer-events-none block size-5 rounded-full bg-white shadow ring-0 transition-transform",
            checked ? "translate-x-5" : "translate-x-0"
          )}
        />
      </button>
    </div>
  );
}

const INITIAL_APP = {
  app_name: "",
  app_address: "",
  app_logo_url: "",
  app_contact_email: "",
  app_contact_phone: "",
};

const INITIAL_NOTIF = {
  notify_on_purchase_admin_store: "false",
  notify_on_purchase_admin_exam: "false",
};

const INITIAL_PAYMENT = {
  midtrans_server_key: "",
  midtrans_client_key: "",
  midtrans_env: "sandbox",
};

export default function SystemConfigPage() {
  const { t } = useTranslation();
  const [tab, setTab] = useState("general");
  const configLoaded = useRef(false);

  const [features, setFeatures] = useState({
    selfReg: true,
    otpMandatory: false,
  });

  const [appFields, setAppFields] = useState(INITIAL_APP);
  const [notifFields, setNotifFields] = useState(INITIAL_NOTIF);
  const [paymentFields, setPaymentFields] = useState(INITIAL_PAYMENT);

  const { data: config, isLoading, error } = useAdminSystemConfig();
  const updateConfig = useUpdateSystemConfig();

  useEffect(() => {
    if (config && !configLoaded.current) {
      configLoaded.current = true;
      setAppFields({
        app_name: config.app_name ?? "",
        app_address: config.app_address ?? "",
        app_logo_url: config.app_logo_url ?? "",
        app_contact_email: config.app_contact_email ?? "",
        app_contact_phone: config.app_contact_phone ?? "",
      });
      setNotifFields({
        notify_on_purchase_admin_store:
          config.notify_on_purchase_admin_store ?? "false",
        notify_on_purchase_admin_exam:
          config.notify_on_purchase_admin_exam ?? "false",
      });
      setPaymentFields({
        midtrans_server_key: config.midtrans_server_key ?? "",
        midtrans_client_key: config.midtrans_client_key ?? "",
        midtrans_env: (config.midtrans_env as "sandbox" | "production") ?? "sandbox",
      });
    }
  }, [config]);

  const handleSaveGeneral = async () => {
    try {
      await updateConfig.mutateAsync(appFields);
      toast.success("Pengaturan umum disimpan");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal menyimpan";
      toast.error(msg);
    }
  };

  const handleSaveNotif = async () => {
    try {
      await updateConfig.mutateAsync(notifFields);
      toast.success("Pengaturan notifikasi disimpan");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal menyimpan";
      toast.error(msg);
    }
  };

  const handleSavePayment = async () => {
    try {
      await updateConfig.mutateAsync(paymentFields);
      toast.success("Pengaturan pembayaran disimpan");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal menyimpan";
      toast.error(msg);
    }
  };

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Settings}
          title="Konfigurasi Sistem"
          description="Memuat…"
        />
        <div className="py-12 text-center text-ink-500">Memuat data…</div>
      </div>
    );
  }

  if (error) {
    const msg =
      (error as { code?: string })?.code === "forbidden"
        ? "Akses ditolak. Hanya Super Admin yang dapat mengakses halaman ini."
        : "Gagal memuat data. Coba refresh halaman.";
    return (
      <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Settings}
          title="Konfigurasi Sistem"
          description="Terjadi kesalahan"
        />
        <div className="py-12 text-center text-ink-500">{msg}</div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={Settings}
        title="Konfigurasi Sistem"
        description="Pengaturan platform dan fitur global."
      />

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-6">
          <TabsTrigger value="general" className="text-xs">
            <Settings className="mr-1 size-4" />
            Umum
          </TabsTrigger>
          <TabsTrigger value="features" className="text-xs">
            <Shield className="mr-1 size-4" />
            Fitur
          </TabsTrigger>
          <TabsTrigger value="notifications" className="text-xs">
            <Bell className="mr-1 size-4" />
            Notifikasi
          </TabsTrigger>
          <TabsTrigger value="payment" className="text-xs">
            <CreditCard className="mr-1 size-4" />
            Pembayaran
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <div className="md-card-outlined">
            <div className="space-y-4">
              <div>
                <Label>Nama platform</Label>
                <Input
                  value={appFields.app_name}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_name: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>Alamat</Label>
                <Input
                  value={appFields.app_address}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_address: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>URL Logo</Label>
                <Input
                  value={appFields.app_logo_url}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_logo_url: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>Email kontak</Label>
                <div className="flex items-center gap-2">
                  <Mail className="size-4 text-ink-400" />
                  <Input
                    type="email"
                    value={appFields.app_contact_email}
                    onChange={(e) =>
                      setAppFields((f) => ({
                        ...f,
                        app_contact_email: e.target.value,
                      }))
                    }
                  />
                </div>
              </div>
              <div>
                <Label>Telepon kontak</Label>
                <Input
                  value={appFields.app_contact_phone}
                  onChange={(e) =>
                    setAppFields((f) => ({
                      ...f,
                      app_contact_phone: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="flex justify-end">
                <Button
                  onClick={handleSaveGeneral}
                  disabled={updateConfig.isPending}
                >
                  <Save className="mr-1 size-4" />
                  {t("save")}
                </Button>
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="features">
          <div className="md-card-outlined">
            <Toggle
              checked={features.selfReg}
              onChange={(v) => setFeatures((f) => ({ ...f, selfReg: v }))}
              label="Registrasi mandiri siswa"
              description="Izinkan siswa mendaftar langsung dari halaman publik."
            />
            <Toggle
              checked={features.otpMandatory}
              onChange={(v) => setFeatures((f) => ({ ...f, otpMandatory: v }))}
              label="OTP wajib saat login"
              description="Kirimkan kode verifikasi untuk setiap percobaan login."
            />
          </div>
        </TabsContent>

        <TabsContent value="notifications">
          <div className="md-card-outlined">
            <Toggle
              checked={notifFields.notify_on_purchase_admin_store === "true"}
              onChange={(v) =>
                setNotifFields((f) => ({
                  ...f,
                  notify_on_purchase_admin_store: v ? "true" : "false",
                }))
              }
              label="Notifikasi pembelian (Store Manager)"
              description="Kirim notifikasi ke admin store saat ada pembelian baru."
            />
            <Toggle
              checked={notifFields.notify_on_purchase_admin_exam === "true"}
              onChange={(v) =>
                setNotifFields((f) => ({
                  ...f,
                  notify_on_purchase_admin_exam: v ? "true" : "false",
                }))
              }
              label="Notifikasi pembelian (Admin Exam)"
              description="Kirim notifikasi ke admin exam saat ada pembelian baru."
            />
            <div className="flex justify-end pt-4">
              <Button
                onClick={handleSaveNotif}
                disabled={updateConfig.isPending}
              >
                <Save className="mr-1 size-4" />
                {t("save")}
              </Button>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="payment">
          <div className="md-card-outlined">
            <div className="space-y-4">
              <div>
                <Label>Midtrans Server Key</Label>
                <Input
                  type="password"
                  value={paymentFields.midtrans_server_key}
                  onChange={(e) =>
                    setPaymentFields((f) => ({
                      ...f,
                      midtrans_server_key: e.target.value,
                    }))
                  }
                  placeholder={
                    paymentFields.midtrans_server_key === "***"
                      ? "***"
                      : "Isi server key"
                  }
                />
                {paymentFields.midtrans_server_key === "***" && (
                  <div className="mt-1 text-xs text-ink-500">
                    Biarkan *** untuk tidak mengubah
                  </div>
                )}
              </div>
              <div>
                <Label>Midtrans Client Key</Label>
                <Input
                  type="password"
                  value={paymentFields.midtrans_client_key}
                  onChange={(e) =>
                    setPaymentFields((f) => ({
                      ...f,
                      midtrans_client_key: e.target.value,
                    }))
                  }
                  placeholder={
                    paymentFields.midtrans_client_key === "***"
                      ? "***"
                      : "Isi client key"
                  }
                />
                {paymentFields.midtrans_client_key === "***" && (
                  <div className="mt-1 text-xs text-ink-500">
                    Biarkan *** untuk tidak mengubah
                  </div>
                )}
              </div>
              <div>
                <Label>Lingkungan Midtrans</Label>
                <Select
                  value={paymentFields.midtrans_env}
                  onValueChange={(v) =>
                    setPaymentFields((f) => ({
                      ...f,
                      midtrans_env: v,
                    }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="sandbox">Sandbox</SelectItem>
                    <SelectItem value="production">Production</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex justify-end">
                <Button
                  onClick={handleSavePayment}
                  disabled={updateConfig.isPending}
                >
                  <Save className="mr-1 size-4" />
                  {t("save")}
                </Button>
              </div>
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
