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
      toast.success(t("config_toast_general_saved"));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("sys_save_failed");
      toast.error(msg);
    }
  };

  const handleSaveNotif = async () => {
    try {
      await updateConfig.mutateAsync(notifFields);
      toast.success(t("config_toast_notif_saved"));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("sys_save_failed");
      toast.error(msg);
    }
  };

  const handleSavePayment = async () => {
    try {
      await updateConfig.mutateAsync(paymentFields);
      toast.success(t("config_toast_payment_saved"));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("sys_save_failed");
      toast.error(msg);
    }
  };

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Settings}
          title={t("config_title")}
          description={t("sys_loading")}
        />
        <div className="py-12 text-center text-ink-500">{t("sys_loading_data")}</div>
      </div>
    );
  }

  if (error) {
    const msg =
      (error as { code?: string })?.code === "forbidden"
        ? t("sys_error_forbidden")
        : t("sys_error_load");
    return (
      <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={Settings}
          title={t("config_title")}
          description={t("sys_error_title")}
        />
        <div className="py-12 text-center text-ink-500">{msg}</div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={Settings}
        title={t("config_title")}
        description={t("config_subtitle")}
      />

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-6">
          <TabsTrigger value="general" className="text-xs">
            <Settings className="mr-1 size-4" />
            {t("config_tab_general")}
          </TabsTrigger>
          <TabsTrigger value="features" className="text-xs">
            <Shield className="mr-1 size-4" />
            {t("config_tab_features")}
          </TabsTrigger>
          <TabsTrigger value="notifications" className="text-xs">
            <Bell className="mr-1 size-4" />
            {t("config_tab_notifications")}
          </TabsTrigger>
          <TabsTrigger value="payment" className="text-xs">
            <CreditCard className="mr-1 size-4" />
            {t("config_tab_payment")}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <div className="md-card-outlined">
            <div className="space-y-4">
              <div>
                <Label>{t("config_general_app_name")}</Label>
                <Input
                  value={appFields.app_name}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_name: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>{t("config_general_address")}</Label>
                <Input
                  value={appFields.app_address}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_address: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>{t("config_general_logo_url")}</Label>
                <Input
                  value={appFields.app_logo_url}
                  onChange={(e) =>
                    setAppFields((f) => ({ ...f, app_logo_url: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>{t("config_general_contact_email")}</Label>
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
                <Label>{t("config_general_contact_phone")}</Label>
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
              label={t("config_feature_selfreg_label")}
              description={t("config_feature_selfreg_desc")}
            />
            <Toggle
              checked={features.otpMandatory}
              onChange={(v) => setFeatures((f) => ({ ...f, otpMandatory: v }))}
              label={t("config_feature_otp_label")}
              description={t("config_feature_otp_desc")}
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
              label={t("config_notif_store_label")}
              description={t("config_notif_store_desc")}
            />
            <Toggle
              checked={notifFields.notify_on_purchase_admin_exam === "true"}
              onChange={(v) =>
                setNotifFields((f) => ({
                  ...f,
                  notify_on_purchase_admin_exam: v ? "true" : "false",
                }))
              }
              label={t("config_notif_exam_label")}
              description={t("config_notif_exam_desc")}
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
                <Label>{t("config_payment_server_key")}</Label>
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
                      : t("config_payment_placeholder_server")
                  }
                />
                {paymentFields.midtrans_server_key === "***" && (
                  <div className="mt-1 text-xs text-ink-500">
                    {t("config_payment_mask_hint")}
                  </div>
                )}
              </div>
              <div>
                <Label>{t("config_payment_client_key")}</Label>
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
                      : t("config_payment_placeholder_client")
                  }
                />
                {paymentFields.midtrans_client_key === "***" && (
                  <div className="mt-1 text-xs text-ink-500">
                    {t("config_payment_mask_hint")}
                  </div>
                )}
              </div>
              <div>
                <Label>{t("config_payment_env")}</Label>
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
