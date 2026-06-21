"use client";

import { useState } from "react";
import { Save, Settings, Bell, Shield, CreditCard, Mail } from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";

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

export default function SystemConfigPage() {
  const { t } = useTranslation();
  const [tab, setTab] = useState("general");
  const [general, setGeneral] = useState({
    platformName: "Akademi Bimbel",
    supportEmail: "support@abak.academy",
    timezone: "Asia/Jakarta",
  });
  const [features, setFeatures] = useState({
    selfReg: true,
    waNotif: true,
    emailNotif: true,
    otpMandatory: false,
  });
  const [payment, setPayment] = useState({
    provider: "Midtrans",
    sandbox: true,
  });

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
                  value={general.platformName}
                  onChange={(e) =>
                    setGeneral((g) => ({ ...g, platformName: e.target.value }))
                  }
                />
              </div>
              <div>
                <Label>Email dukungan</Label>
                <div className="flex items-center gap-2">
                  <Mail className="size-4 text-ink-400" />
                  <Input
                    type="email"
                    value={general.supportEmail}
                    onChange={(e) =>
                      setGeneral((g) => ({ ...g, supportEmail: e.target.value }))
                    }
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <button className="md-btn-filled">
                  <Save className="mr-1 size-4" />
                  Simpan
                </button>
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
              checked={features.emailNotif}
              onChange={(v) => setFeatures((f) => ({ ...f, emailNotif: v }))}
              label="Notifikasi email"
            />
            <Toggle
              checked={features.waNotif}
              onChange={(v) => setFeatures((f) => ({ ...f, waNotif: v }))}
              label="Notifikasi WhatsApp"
            />
          </div>
        </TabsContent>

        <TabsContent value="payment">
          <div className="md-card-outlined">
            <div className="space-y-4">
              <div>
                <Label>Payment gateway</Label>
                <Input value={payment.provider} readOnly />
              </div>
              <Toggle
                checked={payment.sandbox}
                onChange={(v) => setPayment((p) => ({ ...p, sandbox: v }))}
                label="Mode sandbox"
                description="Aktifkan untuk transaksi uji tanpa charge nyata."
              />
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
